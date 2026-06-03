@tool
class_name SelectFieldRow
extends HBoxContainer

const SELECTED_BASE_COLOR: Color = Color("474747b4")
const SELECTED_DIRTY_COLOR: Color = Color("595337b4")
const NORMAL_BASE_COLOR: Color = Color("1a1a1a99")
const NORMAL_DIRTY_COLOR: Color = Color("4a442cb4")

signal field_selected(field: SRData.Field)

@export var _field_button: Button
@export var _field_value_label: Label

var _field: SRData.Field
var _field_name: String
var _field_value: String

var _is_selected: bool = false
var is_dirty: bool = false

func _ready() -> void:
	set_selected(false)
	set_dirty(false)

func init_field(srd: SRData, field: SRData.Field) -> void:
	_field = field
	_field_name = SRData.get_field_name(field)
	_field_value = "???"
	if srd != null:
		set_field_value(srd.get_field_value_str(field))
	set_dirty(false)
	update_status()	

func set_selected(is_selected_new: bool) -> void:
	_is_selected = is_selected_new
	update_status()
	
func set_dirty(is_dirty_new: bool) -> void:
	is_dirty = is_dirty_new
	var unsaved_char: String = ""
	if is_dirty:
		unsaved_char = " *"
	_field_button.text = _field_name + unsaved_char
	update_status()

func set_field_value(value_new: String) -> void:
	_field_value = value_new
	_field_value_label.text = get_first_line(value_new)

func get_first_line(value: String) -> String:
	return value.substr(0, value.find("\n"))

func set_field_value_changed(value_new: String) -> void:
	set_field_value(value_new)
	set_dirty(true)

func get_field_value() -> String:
	return _field_value

func _on_field_button_pressed() -> void:
	field_selected.emit(_field)

func update_status() -> void:
	if _is_selected:
		_set_color(SELECTED_DIRTY_COLOR if is_dirty else SELECTED_BASE_COLOR)
		return

	_set_color(NORMAL_DIRTY_COLOR if is_dirty else NORMAL_BASE_COLOR)

func _set_color(color_new: Color) -> void:
	(_field_button.get("theme_override_styles/normal") as StyleBox).bg_color = color_new
	var highlight_color: Color = Color(color_new)
	highlight_color.v = highlight_color.v + 0.1
	(_field_button.get("theme_override_styles/hover") as StyleBox).bg_color = highlight_color
