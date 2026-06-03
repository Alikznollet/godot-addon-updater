@tool
extends EditorPlugin

const PLUGIN_NAME: String = "SRDock"
const DOCK_FOLDER: String = "/sr_dock"

# A class member to hold the dock during the plugin life cycle.
var dock: EditorDock
var _sr_dock: SRDock

func _enter_tree():
	# Initialization of the plugin goes here.
	_sr_dock = load(get_dock_plugin_path() + "/sr_dock.tscn").instantiate() as SRDock
	dock = EditorDock.new()
	dock.add_child(_sr_dock)
	dock.title = "Spatial Remarks"

	# Note that LEFT_UL means the left of the editor, upper-left dock.
	dock.default_slot = EditorDock.DOCK_SLOT_RIGHT_BL
	
	# Allow the dock to be on the left or right of the editor, and to be made floating.
	dock.available_layouts = EditorDock.DOCK_LAYOUT_VERTICAL | EditorDock.DOCK_LAYOUT_FLOATING

	add_dock(dock)
	scene_changed.connect(_on_scene_changed)
	
func _exit_tree():
	_sr_dock.cleanup_nodes()
	remove_dock(dock)

	dock.queue_free()

	scene_changed.disconnect(_on_scene_changed)
	scene_closed
	
func _on_scene_changed(_node) -> void:
	_sr_dock.scene_changed()

static func get_dock_plugin_path() -> String:
	return SRDataAccess.get_plugin_path() + DOCK_FOLDER


#func _handles(object: Object) -> bool:
	#return object is SRNote
##
#func _get_gizmo_name():
	#return "SpatialRemarks3D"
	##
#func _has_gizmo(node):
	#return node is SRNote
	##
#
#func _init():
	#pass
	#create_material("main", Color(1, 0, 0))
	#create_handle_material("handles")

#var handles: PackedVector3Array
#
#func _redraw(gizmo):
	#gizmo.clear()

	#var node3d = gizmo.get_node_3d()

	#var lines = PackedVector3Array()

	#lines.push_back(Vector3(0, 1, 0))
	#lines.push_back(Vector3(0, node3d., 0))

#	if SRHandler.sr_editor_dirty:
#		handles = PackedVector3Array()
	#	for sr: SRData in SRHandler.sr_data:
#		handles.push_back(sr.global_position)
		
	#handles.push_back(Vector3(0, node3d.my_custom_value, 0))

	#gizmo.add_lines(lines, get_material("main", gizmo), false)
	#gizmo.add_handles(handles, get_material("handles", gizmo), [])
	#
#func _forward_3d_draw_over_viewport(viewport_control: Control) -> void:
	#viewport_control.draw_circle(viewport_control.get_local_mouse_position(), 64, Color.WHITE)
#
#func _forward_canvas_draw_over_viewport(overlay):
	## Draw a circle at the cursor's position.
	#overlay.draw_circle(overlay.get_local_mouse_position(), 64, Color.WHITE)
	#
#func _get_handle_name(gizmo: EditorNode3DGizmo, handle_id: int, secondary: bool) -> String:
	#return "HANDLLE"
	#
#func _get_handle_value(gizmo: EditorNode3DGizmo, handle_id: int, secondary: bool) -> Variant:
	#return 99
#
