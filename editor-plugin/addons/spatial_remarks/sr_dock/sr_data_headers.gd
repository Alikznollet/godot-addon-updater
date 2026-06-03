@tool
class_name SRDataHeaders
extends HBoxContainer

@export var _version_label: Label
@export var _position_label: Label
@export var _target_node_label: Label
@export var _scene_label: Label
@export var _author_label: Label
@export var _text_label: Label
@export var _creation_date_label: Label
@export var _logs_label: Label
@export var _screenshot_label: Label

var used_font: Font

func _ready() -> void:
	used_font = _version_label.get_theme_default_font()

func get_column_text_size(field: SRData.Field) -> int:
	return used_font.get_string_size(get_column(field).text).x + SRDataRow.TEXT_SIZE_OFFSET

func set_column_visible(column: SRData.Field, is_visible_new: bool) -> void:
	get_column(column).visible = is_visible_new
	
func set_column_size(column: SRData.Field, size_new: int) -> void:
	get_column(column).custom_minimum_size.x = size_new

func get_column(column: SRData.Field) -> Label:
	match column:
		SRData.Field.GLOBAL_POSITION: return _position_label
		SRData.Field.TEXT: return _text_label
		SRData.Field.AUTHOR: return _author_label
		SRData.Field.TARGET_NODE: return _target_node_label
		SRData.Field.SCENE: return _scene_label
		SRData.Field.PROJECT_VERSION: return _version_label
		SRData.Field.CREATION_DATE: return _creation_date_label
		SRData.Field.LOGS: return _logs_label
		SRData.Field.SCREENSHOT: return _screenshot_label
		
	push_warning("handling for column ", SRData.Field.keys()[column], " not implemented")
	return null
