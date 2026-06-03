class_name SRCreation
extends CanvasLayer

const NO_VALUE: String = "-"

@export var _position_label: RichTextLabel
@export var _position_value_label: RichTextLabel
@export var _target_object_label: RichTextLabel
@export var _target_object_option_button: OptionButton
@export var _target_object_value_label: RichTextLabel
@export var _scene_label: RichTextLabel
@export var _scene_value_label: RichTextLabel
@export var _author_line_edit: LineEdit
@export var _author_value_label: RichTextLabel
@export var _creation_time_value_label: RichTextLabel
@export var _remark_text_edit: TextEdit
@export var _background_rect: ColorRect
@export var _canvas_group: CanvasGroup
@export var _background_mask: TextureRect
@export var _include_logs_check_box: CheckBox
@export var _include_screenshots_check_box: CheckBox
@export var _remark_preview_sprite_2d: Sprite2D
@export var _remark_preview_sprite_3d: Sprite3D
@export var _show_remarks_check_button: CheckButton

var _remark_to_create: SRData
var _config: SRConfigHandler.Config
var _possible_targets: Array[SRCreationHelper.TargetNodeData]
var _log_lines: String = ""
var _screenshot: Image = null

var _preview_tween: Tween = null

func _enter_tree() -> void:
	process_mode = Node.PROCESS_MODE_ALWAYS

	visible = false

func _ready() -> void:
	
	if InputMap.has_action("show_sr"):
		_show_remarks_check_button.text = _show_remarks_check_button.text + " ("+InputMap.action_get_events("show_sr")[0].as_text()+")"
	set_remarks_visible_button_pressed(SRHandler.remarks_visible)
	_background_rect.size = Vector2(ProjectSettings.get_setting("display/window/size/viewport_width"), ProjectSettings.get_setting("display/window/size/viewport_height"))
	_background_rect.position = -_background_rect.size / 2
	_canvas_group.position = _background_rect.size / 2
	_select_target(-1, false, false)
	
func init(config: SRConfigHandler.Config) -> void:
	_config = config

	if config.create_screenshots != SRConfigHandler.AdditionalDataOption.NEVER:
		_screenshot = create_screenshot(get_viewport())
		
	visible = true
	_remark_to_create = collect_sr_data(config)#, target_position, scene_name, possible_targets[0]) #
	var mouse_pos: Vector2 = get_viewport().get_mouse_position()
	
	if config.include_logs != SRConfigHandler.AdditionalDataOption.NEVER:
		_log_lines = "\n".join(read_recent_logs_from_file(config))
	
	# AUTHOR
	_author_line_edit.text = _remark_to_create.author
	_author_value_label.visible = !config.author_editable
	_author_line_edit.visible = config.author_editable
	_author_line_edit.editable = config.author_editable

	# SCENE VALUE
	match config.scene_name_display:
		SRConfigHandler.NodeReferenceDisplay.PATH:
			_scene_value_label.text = get_viewport().get_tree().current_scene.scene_file_path
			_scene_value_label.visible = true
			_scene_label.visible = true
		SRConfigHandler.NodeReferenceDisplay.NAME:
			_scene_value_label.text = get_viewport().get_tree().current_scene.name
			_scene_value_label.visible = true
			_scene_label.visible = true
		_: # SRDataAccess.NodeReferenceDisplay.NONE:
			_scene_value_label.visible = false
			_scene_label.visible = false

	_remark_to_create.creation_date = Time.get_datetime_string_from_system(true, false)
	_creation_time_value_label.text = Time.get_datetime_string_from_system(false, true)
	
	var sr_creation_helper: SRCreationHelper = SRCreationHelper.new(get_viewport(), config)
	
	_possible_targets = sr_creation_helper.query_possible_targets(mouse_pos)
	var current_target_idx: int = _get_best_target(_possible_targets)

	_position_value_label.text = NO_VALUE
	_target_object_value_label.text = NO_VALUE

	for idx: int in _possible_targets.size():
		_target_object_option_button.add_item(_possible_targets[idx].target_name)
		
	_select_target(current_target_idx, true, false)

	#if current_target_idx >= 0:
		#_select_target(current_target_idx, true)
		#
		#_position_value_label.text = str(_possible_targets[current_target_idx].target_position)

	_target_object_label.visible = config.show_target_node
	_target_object_value_label.visible = config.show_target_node && (_possible_targets.size() == 0 || !config.target_node_selectable)
	_target_object_option_button.visible = config.show_target_node && _possible_targets.size() > 0 && config.target_node_selectable
	_target_object_option_button.disabled = _possible_targets.size() <= 1 || !config.target_node_selectable

	_position_label.visible = config.show_position
	_position_value_label.visible = config.show_position
	
	_background_mask.position = mouse_pos -_background_rect.size / 2 - _background_mask.size / 2

	_include_logs_check_box.visible = _config.include_logs == SRConfigHandler.AdditionalDataOption.CHOOSE
	_include_screenshots_check_box.visible = _config.create_screenshots == SRConfigHandler.AdditionalDataOption.CHOOSE

	#TODO: context?
	#_remark_to_create.context = ...
	
	#TODO: category
	#if config.category_selectable:
	#...

func _on_create_pressed() -> void:
	_remark_to_create.text = _remark_text_edit.text
	
	if _config.author_editable:
		_remark_to_create.author = _author_line_edit.text

	if _config.target_node_selectable:
		var idx: int = _target_object_option_button.selected
		if idx >= 0:
			_remark_to_create.target_node = get_viewport().get_tree().current_scene.get_path_to(_possible_targets[idx].node)
			_remark_to_create.global_position = _possible_targets[idx].target_position
			_remark_to_create.is_2d = _possible_targets[idx].is_2d

	var include_screenshot: bool = _config.create_screenshots == SRConfigHandler.AdditionalDataOption.ALWAYS || (_config.create_screenshots == SRConfigHandler.AdditionalDataOption.CHOOSE && _include_screenshots_check_box.button_pressed)
	var include_logs: bool = _config.include_logs == SRConfigHandler.AdditionalDataOption.ALWAYS || (_config.include_logs == SRConfigHandler.AdditionalDataOption.CHOOSE && _include_logs_check_box.button_pressed)

	if include_screenshot:
		save_screenshot(_screenshot, _config)

	if include_logs:
		_remark_to_create.log_lines = _log_lines

	SRHandler.create_srd(_remark_to_create)
	finish()
	
func _on_cancel_pressed() -> void:
	finish()
	
func finish() -> void:
	SRHandler.end_remark_input()
	queue_free()
	
func collect_sr_data(config: SRConfigHandler.Config) -> SRData:#, target_position: Vector3, scene_name: String, first_target: String) -> SRData:	
	var sr_data: SRData = SRData.new()
	sr_data.author = config.author
	#sr_data.project_name = _config.project_name # unused
	sr_data.project_version = config.project_version
	sr_data.scene = get_viewport().get_tree().current_scene.scene_file_path
	#sr_data.category = config.default_category #TODO
	#sr_data.context = context #TODO
	return sr_data

func _get_best_target(targets: Array[SRCreationHelper.TargetNodeData]) -> int:
	if targets.size() == 0:
		return -1
	if targets.size() > 1:
		return 1
	
	return 0

static func _get_node_name(node: Node) -> String:
	return node.name

func _on_target_object_option_button_item_selected(index: int) -> void:
	_select_target(index, false)

func _select_target(target_index: int, set_target_option_button: bool = true, do_tween: bool = true) -> void:
	var is_2d: bool = target_index < 0 || _possible_targets[target_index].is_2d
	_remark_preview_sprite_3d.visible = target_index >= 0 && !is_2d
	_remark_preview_sprite_2d.visible = target_index >= 0 && is_2d
	
	#_remark
	#if _possible_targets[target_index].is_2d:
		#... use 2D node
	
	if target_index == -1:
		return
		
	if set_target_option_button:
		_target_object_option_button.select(target_index)

	var target_pos: Vector3 = _possible_targets[target_index].target_position
	_position_value_label.text = str(target_pos)
		
	move_preview_position(target_pos, is_2d, 0.1, do_tween)
	
	#_remark_to_create.target_node = _possible_targets[target_index].target_name # not needed? because we already set the target node on creation
	_target_object_value_label.text = _possible_targets[target_index].target_name
	#_remark_to_create.global_position = _possible_targets[current_target_idx].target_position

func move_preview_position(target_pos: Vector3, is_2d: bool, duration: float = 0.1, do_tween: bool = true) -> void:
	if is_instance_valid(_preview_tween):
		_preview_tween.kill()

	if is_2d:
		var target_pos_2d: Vector2 = SRCreationHelper.to_vec2(target_pos)
		if do_tween:
			_preview_tween = create_tween().set_trans(Tween.TRANS_CIRC)
			_preview_tween.tween_property(_remark_preview_sprite_2d, "global_position", target_pos_2d, duration)
		else:
			_remark_preview_sprite_2d.global_position = target_pos_2d
		
		return	
	else:
		if do_tween:
			_preview_tween = create_tween().set_trans(Tween.TRANS_CIRC)
			_preview_tween.tween_property(_remark_preview_sprite_3d, "global_position", target_pos, duration)
		else:
			_remark_preview_sprite_3d.global_position = target_pos
		
func create_screenshot(viewport: Viewport) -> Image:
	var image: Image = viewport.get_texture().get_image()
	return image
		
static func save_screenshot(screenshot: Image, config: SRConfigHandler.Config) -> void:
	if config.data_source_type == SRConfigHandler.DataSourceType.HTTP:
		push_warning("using remote data access for screenshots not implemented")
		return
	
	pass
	
func read_recent_logs_from_file(config: SRConfigHandler.Config) -> Array[String]:
	if !SRHandler.is_logging_enabled():
		return []
	
	var log_file_path: String = ProjectSettings.get("debug/file_logging/log_path")
	var log_file: FileAccess = FileAccess.open(log_file_path, FileAccess.READ)
	var log_lines: Array[String] = []

	while log_file.get_position() < log_file.get_length():
		var next_line: String = log_file.get_line()
		if next_line.strip_edges().is_empty():
			continue
		if log_lines.size() == config.max_log_lines:
			log_lines.pop_front()
		log_lines.append(next_line)
	
	return log_lines

func set_remarks_visible_button_pressed(remarks_visible: bool) -> void:
	_show_remarks_check_button.set_pressed_no_signal(remarks_visible)

func _on_show_remarks_check_button_toggled(toggled_on: bool) -> void:
	SRHandler.set_remarks_visible(toggled_on)
