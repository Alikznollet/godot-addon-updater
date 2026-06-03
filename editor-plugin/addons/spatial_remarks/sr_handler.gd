extends Node

const SR_NOTE_2D_SCENE_PATH: String = "/sr_note_2d.tscn"
const SR_NOTE_3D_SCENE_PATH: String = "/sr_note_3d.tscn"
const SR_CREATION_SCENE_PATH: String = "/sr_creation.tscn"
const SR_IG_OVERLAY_SCENE_PATH: String = "/sr_ingame_overlay.tscn"
const VIEW_SR_AREA_2D_SCENE_PATH: String = "/view_sr_area_2d.tscn"
const VIEW_SR_RAYCAST_3D_SCENE_PATH: String = "/view_sr_raycast_3d.tscn"

static var _sr_creation_scene: PackedScene
static var _sr_note_2d_scene: PackedScene
static var _sr_note_3d_scene: PackedScene
static var _sr_ingame_overlay_scene: PackedScene
static var _view_sr_area_2d_scene: PackedScene
static var _view_sr_raycast_3d_scene: PackedScene

var _config: SRConfigHandler.Config

var _sr_data: Dictionary[SRData, Node]
var _sr_editor_dirty: bool = false
var _sr_data_parent: Node3D = Node3D.new()
var _sr_ingame_overlay: SRIngameOverlay
var _sr_creation_node: SRCreation

var _currently_paused: bool
var _current_mouse_mode: Input.MouseMode
var _previous_mouse_pos: Vector2
var _view_sr_area_2d: ViewSRArea2D
var _view_sr_raycast_3d: ViewSRRaycast3D

var _start_remark_callback: Callable = Callable(_default_start_remark_callback)
var _end_remark_callback: Callable = Callable(_default_end_remark_callback)

var remark_input_active: bool = false
var remarks_visible: bool = false

var _active_camera_2d: Camera2D
var _active_camera_3d: Camera3D

var _http_request: HTTPRequest

func _ready() -> void:
	_http_request = HTTPRequest.new()
	add_child(_http_request)
	SRDataAccessHttp.http_request_node = _http_request
	
	_sr_creation_scene = load(SRDataAccess.get_plugin_path() + SR_CREATION_SCENE_PATH) as PackedScene
	_sr_note_2d_scene = load(SRDataAccess.get_plugin_path() + SR_NOTE_2D_SCENE_PATH) as PackedScene
	_sr_note_3d_scene = load(SRDataAccess.get_plugin_path() + SR_NOTE_3D_SCENE_PATH) as PackedScene
	_sr_ingame_overlay_scene = load(SRDataAccess.get_plugin_path() + SR_IG_OVERLAY_SCENE_PATH) as PackedScene
	_view_sr_area_2d_scene = load(SRDataAccess.get_plugin_path() + VIEW_SR_AREA_2D_SCENE_PATH) as PackedScene
	_view_sr_raycast_3d_scene = load(SRDataAccess.get_plugin_path() + VIEW_SR_RAYCAST_3D_SCENE_PATH) as PackedScene
	
	_previous_mouse_pos = get_viewport().get_mouse_position()

	process_mode = Node.PROCESS_MODE_ALWAYS
	_config = SRConfigHandler.load_config()
	var srds: Array[SRData] = await SRDataAccess.read_all_sr_data(_config)
	_sr_editor_dirty = true

	get_tree().current_scene.ready.connect(init.bind(srds))
	
func init(srds: Array[SRData]) -> void:
	_sr_data_parent.name = "SpatialRemarks"
	_sr_data_parent.visible = remarks_visible
	get_tree().current_scene.add_child(_sr_data_parent)

	_sr_ingame_overlay = _sr_ingame_overlay_scene.instantiate() as SRIngameOverlay
	_sr_data_parent.add_child(_sr_ingame_overlay)

	_add_all_srds(srds)

func _input(event: InputEvent) -> void:
	if event.is_action_pressed("create_sr"):
		start_remark_input()
	elif event.is_action_pressed("show_sr"):
		set_remarks_visible(!remarks_visible)
	
func set_author(author_new: String) -> void:
	_config.author = author_new

func set_custom_pause_callbacks(start_remark_callback: Callable, end_remark_callback: Callable) -> void:
	_start_remark_callback = start_remark_callback
	_end_remark_callback = end_remark_callback

#func add_data(srds: Array[SRData], do_save: bool = true) -> void:
	#_add_all_srds(srds)
	#_sr_editor_dirty = true
	#if do_save:
		#save()

func create_srd(srd: SRData) -> void:
	#_add_all_srds(srds)
	_sr_editor_dirty = true
	SRDataAccess.create_srd(_config, srd)
		
#func save() -> void:
	#SRDataAccess._save_srd(_config, _sr_data.keys())


func start_remark_input() -> void:
	if remark_input_active:
		return
	remark_input_active = true
	_sr_creation_node = _sr_creation_scene.instantiate() as SRCreation
	get_tree().current_scene.add_child(_sr_creation_node)
	_sr_creation_node.init(_config)
	_start_remark_callback.call()
	
func end_remark_input() -> void:
	_end_remark_callback.call()
	remark_input_active = false
	_sr_creation_node.queue_free()
	_sr_creation_node = null
	refresh_sr_data()
		
func _default_start_remark_callback() -> void:
	_currently_paused = get_tree().paused
	get_tree().paused = true
	_current_mouse_mode = Input.mouse_mode
	Input.mouse_mode = Input.MOUSE_MODE_VISIBLE

func _default_end_remark_callback() -> void:
	get_tree().paused = _currently_paused
	Input.mouse_mode = _current_mouse_mode

func set_remarks_visible(visible_new: bool) -> void:
	_sr_data_parent.visible = visible_new
	remarks_visible = visible_new
	
	if _sr_creation_node != null:
		_sr_creation_node.set_remarks_visible_button_pressed(visible_new)

	if !visible_new:
		hide_sr_note()

	if _config.use_raycast_for_display:
		set_sr_note_recognition(visible_new)
	else:
		show_mouse_for_sr_notes(visible_new)

func get_collision_layer_number() -> int:
	return _config.collision_layer_number

func _add_all_srds(srds: Array[SRData]) -> void:
	for srd: SRData in srds:
		_add_sr_node_to_tree(srd)

func _add_sr_node_to_tree(srd: SRData) -> void:
	if get_tree().current_scene.scene_file_path != srd.scene:
		_sr_data[srd] = null
		return
	var target_node_3d: SRNote3D = _sr_note_3d_scene.instantiate() as SRNote3D
	
	_sr_data_parent.add_child(target_node_3d)
	_sr_data[srd] = target_node_3d
	target_node_3d.init(srd, _config.collision_layer_number, false)

func refresh_sr_data() -> void:
	for srd: SRData in _sr_data.keys():
		_sr_data[srd].queue_free()
	
	_sr_data.clear()
	var srds: Array[SRData] = await SRDataAccess.read_all_sr_data(_config)
	_add_all_srds(srds)

func show_sr_note(srd: SRData) -> void:
	_sr_ingame_overlay.show_remark(srd)
	
func hide_sr_note() -> void:
	_sr_ingame_overlay.hide_remark()
	
func set_sr_note_recognition(is_visible_new: bool) -> void:
	if is_visible_new:
		
		if is_active_2d():
			var current_cam_2d: Camera2D = get_current_camera_2d()
			if current_cam_2d == null:
				push_warning("no camera2d found!")
				return
			_view_sr_area_2d = _view_sr_area_2d_scene.instantiate() as ViewSRArea2D
			current_cam_2d.add_child(_view_sr_area_2d)		
		else:
			var current_cam_3d: Camera3D = get_current_camera_3d()
			if current_cam_3d == null:
				push_warning("no camera3d found!")
				return
			_view_sr_raycast_3d = _view_sr_raycast_3d_scene.instantiate() as ViewSRRaycast3D
			current_cam_3d.add_child(_view_sr_raycast_3d)
	else:
		if _view_sr_raycast_3d != null:
			_view_sr_raycast_3d.queue_free()
	
func show_mouse_for_sr_notes(is_visible_new: bool) -> void:
	if !is_visible_new:
		_previous_mouse_pos = get_viewport().get_mouse_position()

		Input.mouse_mode = _current_mouse_mode
		hide_sr_note()
	else:
		var mouse_position: Vector2 = get_viewport().get_mouse_position()
		_current_mouse_mode = Input.mouse_mode
		Input.mouse_mode = Input.MOUSE_MODE_VISIBLE
		get_viewport().warp_mouse(_previous_mouse_pos) #TODO is this really better?

func set_active_camera_3d(active_camera_3d_new: Camera3D) -> void:
	var is_same_camera: bool = _active_camera_3d == active_camera_3d_new
	_active_camera_3d = active_camera_3d_new
	
	if !is_same_camera && _config.use_raycast_for_display && remarks_visible && _active_camera_3d != null:
		_view_sr_raycast_3d.queue_free()
		_view_sr_raycast_3d = _view_sr_raycast_3d_scene.instantiate() as ViewSRRaycast3D
		_active_camera_3d.add_child(_view_sr_raycast_3d)
		
func get_current_camera_3d() -> Camera3D:
	if _active_camera_3d != null:
		return _active_camera_3d
	
	return get_viewport().get_camera_3d()
	
func set_active_camera_2d(active_camera_2d_new: Camera2D) -> void:
	var is_same_camera: bool = _active_camera_2d == active_camera_2d_new
	_active_camera_2d = active_camera_2d_new

	if !is_same_camera && _config.use_raycast_for_display && remarks_visible && _active_camera_2d != null:
		_view_sr_area_2d.queue_free()
		_view_sr_area_2d = _view_sr_area_2d_scene.instantiate() as ViewSRArea2D
		_active_camera_2d.add_child(_view_sr_area_2d)
		
		
func get_current_camera_2d() -> Camera2D:
	if _active_camera_2d != null:
		return _active_camera_2d
	
	return get_viewport().get_camera_2d()

func is_active_3d() -> bool:
	if _active_camera_3d != null:
		return true
	if _active_camera_2d != null:
		return false
		
	return get_viewport().get_camera_3d() != null

func is_active_2d() -> bool:
	if _active_camera_2d != null:
		return true
	if _active_camera_3d != null:
		return false
		
	return get_viewport().get_camera_2d() != null

static func is_logging_enabled() -> bool:
	return ProjectSettings.get("debug/file_logging/enable_file_logging") || ProjectSettings.get("debug/file_logging/enable_file_logging.pc")
