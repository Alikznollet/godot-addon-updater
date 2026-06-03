@tool
class_name SRDock
extends PanelContainer

const CAMERA_NOTE_VIEW_DISTANCE: float = 8.0
const CAMERA_MOVE_TIME: float = 0.35

const SR_NOTE_2D_SCENE_PATH: String = "/sr_note_2d.tscn"
const SR_NOTE_3D_SCENE_PATH: String = "/sr_note_3d.tscn"

static var _sr_note_2d_scene: PackedScene
static var _sr_note_3d_scene: PackedScene

@export var _srd_detail: SRDDetail
@export var _sr_data_table: SRDataTable

var _config: SRConfigHandler.Config

var _sr_data_ar: Array[SRData]

var _node_in_scene: Node
var _instantiated_nodes: Dictionary[SRData, Node] = {}
var _selected_srd: SRData = null

# if this is viewed as a scene in the editor and not in the dock, some things might not work
var _is_dock_child: bool = false
var _file_dialog: EditorFileDialog
var col_selection: PopupMenu

var jump_node_tween: Tween = null
var editor_selection: EditorSelection
var _http_request_node: HTTPRequest

func _ready() -> void:
	_http_request_node = HTTPRequest.new()
	add_child(_http_request_node)
	SRDataAccessHttp.http_request_node = _http_request_node
	
	_sr_note_2d_scene = load(SRDataAccess.get_plugin_path() + SR_NOTE_2D_SCENE_PATH) as PackedScene
	_sr_note_3d_scene = load(SRDataAccess.get_plugin_path() + SR_NOTE_3D_SCENE_PATH) as PackedScene
	_srd_detail.visible = false
	_config = SRConfigHandler.load_config()
	
	if !get_parent() is EditorDock:
		return
	editor_selection = EditorInterface.get_selection()
	(get_parent() as EditorDock).resized.connect(_on_resized)
	_is_dock_child = true
	editor_selection.selection_changed.connect(on_selection_changed)
	
func scene_changed() -> void:
	load_sr_data()
	
func load_sr_data(reload_config: bool = false) -> void:
	if reload_config:
		_config = SRConfigHandler.load_config()
		
	if get_parent() is EditorDock:
		_is_dock_child = true

	_sr_data_ar = await SRDataAccess.read_all_sr_data(_config)
	SRDataAccess.sort_by_id(_sr_data_ar)
	
	_sr_data_table.update_sr_data(_sr_data_ar)
	_update_sr_node_display()
	if _selected_srd != null:
		var date: String = _selected_srd.creation_date
		var author: String = _selected_srd.author
		var selected: bool = false
		for srd: SRData in _sr_data_ar:
			if srd.creation_date == date && srd.author == author:
				_on_select_entry(srd)
				selected = true
		if !selected:
			_on_unselect_entry()

func _update_sr_node_display() -> void:
	if !_is_dock_child:
		return
	cleanup_nodes()
	
	_node_in_scene = Node.new()
	
	if EditorInterface.get_edited_scene_root() != null:
		EditorInterface.get_edited_scene_root().add_child(_node_in_scene)

	for srd: SRData in _sr_data_ar:
		_add_sr_node_to_scene(srd)
		
func _add_sr_node_to_scene(from_srd: SRData) -> void:
	if from_srd.scene != EditorInterface.get_edited_scene_root().scene_file_path:
		return
	
	var target_node: Node
	if from_srd.is_2d:
		#if !EditorInterface.get_edited_scene_root() is Node2D:
			#return
		target_node = _sr_note_2d_scene.instantiate() as SRNote2D
		_node_in_scene.add_child(target_node, false, InternalMode.INTERNAL_MODE_FRONT)
		(target_node as SRNote2D).init(from_srd, 0, true)
		
	else:
		#if !EditorInterface.get_edited_scene_root() is Node3D:
			#return
		target_node = _sr_note_3d_scene.instantiate() as SRNote3D
		_node_in_scene.add_child(target_node, false, InternalMode.INTERNAL_MODE_FRONT)
		(target_node as SRNote3D).init(from_srd, 0, true)
	target_node.owner = EditorInterface.get_edited_scene_root()
	_instantiated_nodes[from_srd] = target_node

func cleanup_nodes() -> void:
	for node_obj: Node in _instantiated_nodes.values():
		if is_instance_valid(node_obj):
			(node_obj).queue_free()
	
	_instantiated_nodes.clear()

func _on_resized() -> void:
	var psize: int = (get_parent() as EditorDock).size.x	
	_sr_data_table.update_field_sizes(psize)

func _on_unselect_entry() -> void:
	if !_is_dock_child:
		return
		
	_sr_data_table.set_selected(_selected_srd, false)

	if _instantiated_nodes.has(_selected_srd):
		_instantiated_nodes[_selected_srd].set_highlighted(false)
	_srd_detail.visible = false
	_selected_srd = null
	
func _on_select_entry(entry: SRData) -> void:
	if !_is_dock_child:
		return
	if _selected_srd != null:
		_sr_data_table.set_selected(_selected_srd, false)
		if _instantiated_nodes.has(_selected_srd):
			_instantiated_nodes[_selected_srd].set_highlighted(false)
		
	_selected_srd = entry
	_srd_detail.visible = true
	_srd_detail.load_srd(entry)
	if _instantiated_nodes.has(entry):
		_instantiated_nodes[entry].set_highlighted(true)
	_sr_data_table.set_selected(entry, true)

func _on_sr_table_select_entry(srd: SRData) -> void:
	_on_select_entry(srd)
	
func _on_import_button_pressed() -> void:
	# show import selection menu
	# connect its finished-btn to do_import()
	if _file_dialog != null:
		_file_dialog.queue_free()
	
	_file_dialog = EditorFileDialog.new()
	var favorites: PackedStringArray = _file_dialog.get_favorite_list().duplicate()
	if !favorites.has(ProjectSettings.globalize_path("res://")):
		favorites.append(ProjectSettings.globalize_path("res://"))
	if !favorites.has(ProjectSettings.globalize_path("user://")):
		favorites.append(ProjectSettings.globalize_path("user://"))
	_file_dialog.set_favorite_list(favorites)
	_file_dialog.visible = false
	_file_dialog.file_mode = FileDialog.FILE_MODE_OPEN_FILES
	_file_dialog.display_mode = FileDialog.DISPLAY_LIST
	_file_dialog.access = FileDialog.ACCESS_FILESYSTEM
	_file_dialog.filters = ["*.json"]
	_file_dialog.ok_button_text = "Import"
	_file_dialog.files_selected.connect(_do_import_from_paths)
	EditorInterface.popup_dialog_centered(_file_dialog)
	#_file_dialog.visible = true

	#_file_dialog.popup_centered()
	
func _do_import_from_paths(paths: PackedStringArray) -> void:
	var added_data: Array[SRData] = []
	
	for path: String in paths:
		if !path.ends_with(".json"):
			push_warning("SRDock: loading non-json files is not supported currently. Path was: '", path, "'")
			continue		
		var res: Array[SRData] = SRDataAccess.import_srds_from_json_file(path)
		added_data.append_array(res)
	
	if added_data.size() > 0:
		_do_import(added_data)
	else:
		print("SRDock: No data found to import.")
	if _file_dialog != null:
		_file_dialog.queue_free()

func _on_refresh_button_pressed() -> void:
	load_sr_data(true)
	
func _do_import(added_data: Array[SRData]) -> void:
	for data: SRData in added_data:
		data.id = SRData.NO_ID #ensure we don't update existing data
	
	_sr_data_ar.append_array(added_data)
	SRDataAccess.save_all_srd(_config, _sr_data_ar)
	load_sr_data(false)

func _on_srd_detail_delete_srd(srd: SRData) -> void:
	SRDataAccess.delete_srd(_config, srd)
	_on_unselect_entry()
	load_sr_data(false)

func _on_srd_detail_update_srd(srd: SRData) -> void:
	SRDataAccess.update_srd(_config, srd)
	load_sr_data(false)

func _on_column_button_pressed() -> void:
	if col_selection != null:
		col_selection.queue_free()
		
	col_selection = PopupMenu.new()
	col_selection.hide_on_checkable_item_selection = false
	
	var visible_columns: Array[SRData.Field] = _sr_data_table.visible_columns
	var idx: int = 0
	for col: SRData.Field in SRData.Field.values():
		if !SRData.is_column_in_editor(col):
			continue
		var col_name: String = (SRData.Field.keys()[col] as String).to_pascal_case()
		col_selection.add_check_item(col_name, col)
		col_selection.set_item_checked(idx, visible_columns.has(col))
		idx += 1
	col_selection.id_pressed.connect(_toggle_col_selection)
	EditorInterface.popup_dialog(col_selection, Rect2i(get_screen_transform() * get_local_mouse_position(), Vector2i.ZERO))

func _toggle_col_selection(col_id: int) -> void:
	_sr_data_table.toggle_column_visible(col_id as SRData.Field)
	if col_selection != null:
		var idx: int = col_selection.get_item_index(col_id)
		col_selection.set_item_checked(idx, _sr_data_table.visible_columns.has(col_id))
		
func _on_srd_detail_unselect_srd() -> void:
	_on_unselect_entry()

func _on_sr_data_table_jump_entry(srd: SRData) -> void:
	move_camera_to_node(srd)

func move_camera_to_node(srd: SRData) -> void:
	if !_instantiated_nodes.has(srd):
		return
		
	var editor_settings: EditorSettings = EditorInterface.get_editor_settings()
	var focus_shortcut: Shortcut = editor_settings.get_shortcut("spatial_editor/focus_selection")
	var focus_event: InputEvent = focus_shortcut.events[0] if focus_shortcut.events.size() > 0 else null
	#var selected_nodes: Array[Node] = editor_selection.get_selected_nodes()

	if focus_event != null:
		editor_selection.clear()
		editor_selection.add_node(_instantiated_nodes[srd])
		focus_event.pressed = true
		Input.parse_input_event(focus_event)	#
	else:
		if srd.is_2d:
			jump_node_local_2d(SRCreationHelper.to_vec2(srd.global_position))
		else:
			jump_node_local_3d(srd.global_position)
		
	await get_tree().create_timer(0.25).timeout
	#editor_selection.clear()
	var target_node_path: String = srd.target_node
	var has_target_node: bool = EditorInterface.get_edited_scene_root().has_node(target_node_path)
	if has_target_node:
		var target_node: Node = EditorInterface.get_edited_scene_root().get_node(target_node_path)
		editor_selection.add_node(target_node)
		#EditorInterface.get_inspector().edit(target_node)

func jump_node_local_2d(srd_pos: Vector2) -> void:
	var cam2d: Camera2D = EditorInterface.get_editor_viewport_3d().get_camera_2d()
	if cam2d == null:
		return
	
	if is_instance_valid(jump_node_tween):
		jump_node_tween.kill()

	jump_node_tween = create_tween().set_ease(Tween.EASE_OUT).set_trans(Tween.TRANS_CUBIC)
	jump_node_tween.tween_property(cam2d, "global_position", srd_pos, CAMERA_MOVE_TIME)
	#cam2d.global_position = Vector2.ZERO
	
func jump_node_local_3d(srd_pos: Vector3) -> void:
	var cam3d: Camera3D = EditorInterface.get_editor_viewport_3d().get_camera_3d()
	if cam3d == null:
		return
	
	if is_instance_valid(jump_node_tween):
		jump_node_tween.kill()

	jump_node_tween = create_tween().set_ease(Tween.EASE_OUT).set_trans(Tween.TRANS_CUBIC)
	
	var lookat: Vector3 = cam3d.global_transform.basis.z.normalized() * CAMERA_NOTE_VIEW_DISTANCE
	var target_position: Vector3 = srd_pos + lookat

	jump_node_tween.tween_property(cam3d, "global_position", target_position, CAMERA_MOVE_TIME)
	#cam3d.global_position = Vector3.ZERO

func on_selection_changed() -> void:
	var editor_selection_new: Array[Node] = editor_selection.get_selected_nodes()

	var non_sr_nodes: Array[Node] = []
	var sr_notes: Array[SRNote3D] = []
		
	for s_node: Node in editor_selection_new:
		if s_node is SRNote3D:
			sr_notes.append(s_node as SRNote3D)
		else:
			non_sr_nodes.append(s_node)

	var inspector: EditorInspector = EditorInterface.get_inspector()
	if !sr_notes.is_empty():
		var last_selection: SRNote3D = sr_notes[sr_notes.size()-1] as SRNote3D
		var srd: SRData = _instantiated_nodes.find_key(last_selection)
		_on_select_entry(srd)
		if non_sr_nodes.is_empty():
			#var target_node_path: String = srd.target_node
			#var has_target_node: bool = EditorInterface.get_edited_scene_root().has_node(target_node_path)
			#if has_target_node:
				#var target_node: Node = EditorInterface.get_edited_scene_root().get_node(target_node_path)
				#inspector.edit(target_node)
			#else:
			inspector.edit(null)
	else:
		return
