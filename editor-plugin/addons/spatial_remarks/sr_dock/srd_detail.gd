@tool
class_name SRDDetail
extends HBoxContainer

const SELECT_FIELD_ROW_SCENE_PATH: String = "/sr_dock/select_field_row.tscn"

signal update_srd(srd: SRData)
signal delete_srd(srd: SRData)
signal unselect_srd()

static var _select_field_row_scene: PackedScene

@export var _text_edit: TextEdit
@export var _field_vbc: VBoxContainer

var _srd: SRData

var _field_rows: Dictionary[SRData.Field, SelectFieldRow] = {}
var _current_field: SRData.Field = SRData.Field.TEXT

func _ready() -> void:
	_select_field_row_scene = load(SRDataAccess.get_plugin_path() + SELECT_FIELD_ROW_SCENE_PATH) as PackedScene

	for field: SRData.Field in SRData.Field.values():
		var select_field_row: SelectFieldRow = _select_field_row_scene.instantiate() as SelectFieldRow
		_field_vbc.add_child(select_field_row)
		select_field_row.init_field(null, field)
		select_field_row.field_selected.connect(_on_field_selected)
		_field_rows[field] = select_field_row

func load_srd(srd_new: SRData) -> void:
	_srd = srd_new
	if srd_new != null:
		_text_edit.text = srd_new.text

	for field: SRData.Field in SRData.Field.values():
		_field_rows[field].init_field(srd_new, field)

	_on_field_selected(_current_field)

func _on_save_button_pressed() -> void:
	var has_dirty_field: bool = false
	for field: SRData.Field in SRData.Field.values():
		if _field_rows[field].is_dirty:
			has_dirty_field = true
			_srd.set_field_from_str(field, _field_rows[field].get_field_value())
	
	if !has_dirty_field:
		print("SRDetail: nothing to update. Skip saving.")
		return
			
	update_srd.emit(_srd)
	
	for field: SRData.Field in SRData.Field.values():
		_field_rows[field].set_dirty(false)
		
func _on_delete_button_pressed() -> void:
	delete_srd.emit(_srd)

func _on_close_button_pressed() -> void:
	unselect_srd.emit()

func _on_text_edit_text_changed() -> void:
	_field_rows[_current_field].set_field_value_changed(_text_edit.text)

func _on_field_selected(new_field: SRData.Field) -> void:
	_field_rows[_current_field].set_selected(false)
	_current_field = new_field
	_field_rows[new_field].set_selected(true)
	_text_edit.text = _field_rows[new_field].get_field_value()
	_text_edit.editable = SRData.is_editable_in_editor(new_field)
		
