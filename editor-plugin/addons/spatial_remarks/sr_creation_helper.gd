class_name SRCreationHelper
extends Object

const FULL_COL_MASK: int = 0xFFFFFFFF

const CIRCLE_SHAPE_2D_RADIUS: int = 20

const MAX_DISTANCE_3D: int = 50

const CONE_RESOLUTION: int = 20
const CONE_NEAR_RADIUS: float = 100

const RAY_3D_MAX_ITERATIONS: int = 5
const RAY_3D_MIN_OFFSET: float = 0.1

const ORIGIN_POS_NAME: String = "Self"

class TargetNodeData:
	var node: Node
	var target_position: Vector3
	var target_name: String
	var is_2d: bool
		
	func _init(node_new: Node, target_position_new: Vector3, is_2d_new: bool, custom_name_new: String = "") -> void:
		node = node_new
		target_position = target_position_new
		is_2d = is_2d_new
		target_name = custom_name_new if !custom_name_new.is_empty() else node_new.name

var _viewport: Viewport
var _config: SRConfigHandler.Config

func _init(viewport: Viewport, config: SRConfigHandler.Config) -> void:
	_viewport = viewport
	_config = config

func query_possible_targets(target_pos: Vector2) -> Array[TargetNodeData]:
	if SRHandler.is_active_3d():
		return _query_possible_targets_3d(target_pos)
		
	if SRHandler.is_active_2d():
		return _query_possible_targets_2d(target_pos)
		
	return _query_possible_targets_no_camera()

func _query_possible_targets_no_camera() -> Array[TargetNodeData]:
	var control: Control = _viewport.gui_get_hovered_control()
	if control == null:
		return []
	var results: Array[TargetNodeData] = [TargetNodeData.new(control, to_vec3(control.position), true)]
	return results

func _query_possible_targets_2d(target_pos: Vector2) -> Array[TargetNodeData]:
	var camera_2d: Camera2D = SRHandler.get_current_camera_2d()

	var space_state: PhysicsDirectSpaceState2D = _viewport.get_world_2d().direct_space_state
	var query: PhysicsShapeQueryParameters2D = PhysicsShapeQueryParameters2D.new()
	query.collide_with_areas = true
	query.collide_with_bodies = true
	query.transform = Transform2D(0, camera_2d.global_position)
	query.collision_mask = get_all_layers_except_for(_config.collision_layer_number)
	query.shape_rid = get_circle_shape_2d()

	var intersects: Array[Dictionary] = space_state.intersect_shape(query)
	var results: Array[TargetNodeData] = [TargetNodeData.new(camera_2d, to_vec3(camera_2d.global_position), true, ORIGIN_POS_NAME)]

	for intersect: Dictionary in intersects:
		if intersect.has("collider"):
			var target: Node2D = (intersect["collider"] as Node2D)
			#var pos: Vector2 = target.global_position as Vector2 
			
			results.append(TargetNodeData.new(target, to_vec3(target_pos), true))

	results.append_array(_query_possible_targets_no_camera())	
	return results


func _query_possible_targets_3d(target_pos: Vector2) -> Array[TargetNodeData]:
	var camera_3d: Camera3D = SRHandler.get_current_camera_3d()

	var collision_results: Dictionary[Node3D, TargetNodeData]
	
	# find raycast results
	for ray_intersect_result: Dictionary in get_raycast_collision_results(target_pos):
		var target: Node3D = (ray_intersect_result["collider"] as Node3D)
		
		if target == camera_3d:
			continue

		collision_results[target] = TargetNodeData.new(target, ray_intersect_result["position"], false)
		
	# find cone results, but only add nodes that don't exist yet as intersection point is not computed here
	for intersect: Dictionary in get_cone_point_collision_results(target_pos, camera_3d):
		if !intersect.has("collider"):
			continue
		
		var target: Node3D = (intersect["collider"] as Node3D)
		if collision_results.has(target):
			continue
			
		if target == camera_3d:
			continue
		
		collision_results[target] = TargetNodeData.new(target, target.global_position, false)

	# convert to TargetNodeData, sort by distance, reduce to max_num_query_results as specified by config
	var results: Array[TargetNodeData] = collision_results.values()
	results.sort_custom(sort_by_distance_to.bind(camera_3d.global_position))
	
	var no_camera_results: Array[TargetNodeData] = _query_possible_targets_no_camera()
	var num_remaining_query_results: int = _config.max_num_query_results - no_camera_results.size()

	if results.size() > num_remaining_query_results:
		results = results.slice(0, num_remaining_query_results)
	
	results.append_array(no_camera_results)
	results.insert(0, TargetNodeData.new(camera_3d, camera_3d.global_position, false, ORIGIN_POS_NAME))

	return results
	
func sort_by_distance_to(a: TargetNodeData, b: TargetNodeData, origin_position: Vector3) -> bool:
	return origin_position.distance_squared_to(a.target_position) < origin_position.distance_squared_to(b.target_position)
	
func get_raycast_collision_results(target_pos: Vector2) -> Array[Dictionary]:
	var camera_3d: Camera3D = SRHandler.get_current_camera_3d()
	var space_state: PhysicsDirectSpaceState3D = _viewport.get_world_3d().direct_space_state
	
	var look_dir: Vector3 = camera_3d.project_ray_normal(target_pos)
	var start: Vector3 = camera_3d.project_ray_origin(target_pos) + look_dir * RAY_3D_MIN_OFFSET
	var end: Vector3 = start + look_dir * MAX_DISTANCE_3D

	var collision_rids: Array[RID] = []
	
	var results: Array[Dictionary] = []
	
	for i in RAY_3D_MAX_ITERATIONS:
		var query: PhysicsRayQueryParameters3D = PhysicsRayQueryParameters3D.create(start, end)
		query.exclude = collision_rids
		var ray_intersect_result: Dictionary = space_state.intersect_ray(query)
		if ray_intersect_result.has("collider"):
			results.append(ray_intersect_result)
			collision_rids.append(ray_intersect_result["rid"] as RID)
		else:
			break

	return results

func get_cone_point_collision_results(target_pos: Vector2, camera_3d: Camera3D) -> Array[Dictionary]:
	var space_state: PhysicsDirectSpaceState3D = _viewport.get_world_3d().direct_space_state
	var query: PhysicsShapeQueryParameters3D = PhysicsShapeQueryParameters3D.new()
	
	query.collide_with_areas = true
	query.collide_with_bodies = true
	query.shape_rid = get_cone_shape_3d(camera_3d, target_pos)
	
	 # not needed, we already have the correct transform from the cone shape
	#query.transform = Transform3D(camera_3d.global_transform)
	
	query.collision_mask = get_all_layers_except_for(_config.collision_layer_number)

	return space_state.intersect_shape(query)
	
func get_cone_shape_3d(camera_3d: Camera3D, mouse_pos: Vector2) -> RID:
	var cone_shape_rid: RID = PhysicsServer3D.convex_polygon_shape_create()
	PhysicsServer3D.shape_set_data(cone_shape_rid, _generate_cone_points(camera_3d, mouse_pos, _config.max_3d_range, CONE_NEAR_RADIUS, CONE_RESOLUTION))

	return cone_shape_rid

static func _generate_cone_points(camera_3d: Camera3D, mouse_pos: Vector2, height: float, initial_radius: float, circle_resolution: int) -> PackedVector3Array:
	var points: PackedVector3Array = []
	for i in circle_resolution:
		var step_size: float = float(circle_resolution)/i
		var circle_point_vec2: Vector2 = mouse_pos + Vector2(cos(TAU/step_size) * initial_radius, sin(TAU/step_size)* initial_radius)
		var circle_point_near: Vector3 = camera_3d.project_ray_origin(circle_point_vec2) + camera_3d.project_ray_normal(circle_point_vec2) * 1.0
		var circle_point_far: Vector3 = camera_3d.project_ray_origin(circle_point_vec2) + camera_3d.project_ray_normal(circle_point_vec2) * height
		points.append(circle_point_near)
		points.append(circle_point_far)

	return points

static func get_circle_shape_2d() -> RID:
	var circle_shape_rid: RID = PhysicsServer2D.circle_shape_create()
	PhysicsServer2D.shape_set_data(circle_shape_rid, CIRCLE_SHAPE_2D_RADIUS)
	
	return circle_shape_rid

static func get_all_layers_except_for(layer_number: int) -> int:
	var bitflag: int = 1 << layer_number-1
	return 0xFFFFFFFF & ~bitflag

static func to_vec3(vec2: Vector2, z: float = 0.0) -> Vector3:
	return Vector3(vec2.x, vec2.y, z)
	
static func to_vec2(vec3: Vector3) -> Vector2:
	return Vector2(vec3.x, vec3.y)
