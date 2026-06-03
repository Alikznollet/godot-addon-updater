@tool
class_name SRDataTable
extends VBoxContainer

const SR_DATA_ROW_SCENE_PATH: String = "/sr_dock/sr_data_row.tscn"

const RESIZE_OFFSET: int = 100

signal select_entry(srd: SRData)
signal jump_entry(srd: SRData)

static var _sr_data_row: PackedScene

@export var _content_vbc: VBoxContainer
@export var _data_headers_hbc: SRDataHeaders

var _field_text_sorted_ids: Array[int] = []
var _field_text_sizes: Dictionary[SRData.Field, int] = {}
var _field_text_eq_tresh_size: int = 0
var _field_text_full_size: int = 0
var _parent_size: int = 0

var _previous_size: int = 0
var _inner_resize: bool = false

var visible_columns: Array[SRData.Field] = [SRData.Field.TARGET_NODE, SRData.Field.AUTHOR, SRData.Field.TEXT, SRData.Field.GLOBAL_POSITION]

var _sr_data: Dictionary[SRData, SRDataRow] = {}

func _ready() -> void:
	_sr_data_row = load(SRDataAccess.get_plugin_path() + SR_DATA_ROW_SCENE_PATH) as PackedScene
	for field: SRData.Field in SRData.Field.values():
		if visible_columns.has(field) && !SRData.is_column_in_editor(field):
			visible_columns.erase(field)

func update_sr_data(sr_data_ar: Array[SRData]) -> void:
	for srdr: SRDataRow in _sr_data.values():
		srdr.queue_free()
		
	_sr_data.clear()
	
	for srd: SRData in sr_data_ar:
		var srdr: SRDataRow = _add_sr_data_row(srd)

	_update_field_data()
	
func _update_field_data() -> void:
	_field_text_sizes.clear()
	
	for col: SRData.Field in visible_columns:
		_field_text_sizes[col] = _data_headers_hbc.get_column_text_size(col)
	
	for srdr: SRDataRow in _sr_data.values():
		for col: SRData.Field in visible_columns:
			_field_text_sizes[col] = max(_field_text_sizes[col], srdr.get_field_text_size(col))

	_field_text_full_size = _field_text_sizes.values().reduce(func(a:int, b:int) -> int: return a+b)
	_field_text_eq_tresh_size = _field_text_sizes.values().min() * _field_text_sizes.size()
	_field_text_sorted_ids.assign(_field_text_sizes.keys())
	_field_text_sorted_ids.sort_custom(_sort_by_field_size)

	for col: SRData.Field in SRData.Field.values():
		if !SRData.is_column_in_editor(col):
			continue
		_set_column_visible(col, visible_columns.has(col), true)
	_compute_field_sizes()

func _add_sr_data_row(from_srd: SRData) -> SRDataRow:
	var srdr: SRDataRow = _sr_data_row.instantiate()
	_content_vbc.add_child(srdr)
	srdr.init(from_srd)
	
	var is_active_scene: bool = false
	if is_instance_valid(EditorInterface.get_edited_scene_root()):
		is_active_scene = from_srd.scene == EditorInterface.get_edited_scene_root().scene_file_path
		
	srdr.set_active_scene(is_active_scene)

	_sr_data[from_srd] = srdr
	srdr.select_entry.connect(_on_select_entry)
	srdr.jump_selection.connect(_on_jump_to_entry)
	return srdr
	
func _sort_by_field_size(a: int, b: int) -> bool:
	return _field_text_sizes[a] < _field_text_sizes[b]
	
func update_field_sizes(parent_size: int) -> void:
	# Prevent retriggered _on_resized() calls from setting field sizes
	if _inner_resize:
		return
	
	#var psize: int = (get_parent() as EditorDock).size.x
	if parent_size == _parent_size:
		return
	_parent_size = parent_size
	_compute_field_sizes()
	
func _compute_field_sizes() -> void:
	
	if _field_text_sizes.size() == 0:
		# Nothing to compute here
		return
	
	var num_fields: int = visible_columns.size()

	# compute the full space that is available
	# if this is too close to the current size, the minimum size set in the children will slow down/block too big resizes
	var max_size_x: int = max(_parent_size * 0.8, _parent_size - RESIZE_OFFSET)

	var target_sizes: Dictionary[SRData.Field, int] = {}
	# shortcut: all fields have reached their content length
	if max_size_x >= _field_text_full_size:
		var available_rests: float = (max_size_x-_field_text_full_size) as float / num_fields

		for field: SRData.Field in visible_columns:
			target_sizes[field] = _field_text_sizes[field] + floori(available_rests)
			
		_set_field_sizes(target_sizes)
		#_set_field_sizes(_field_text_sizes) # TODO: use this instead for keeping the table centered?
		return
		
	var reserved_x_size: float = max_size_x as float / num_fields # the size that each field can have
	
	# shortcut: no field has reached its content length
	if max_size_x <= _field_text_eq_tresh_size:
		_set_column_sizes_uniform(reserved_x_size)
		return

	var remaining_size: int = max_size_x
	
	# share the rest size among the other column
	var remaining_fields: int = num_fields
	for column: SRData.Field in visible_columns:
		var target_size_used: int = mini(floori(reserved_x_size), _field_text_sizes[column])
		target_sizes[column] = target_size_used
		remaining_size -= target_size_used
		remaining_fields -= 1
		if remaining_fields > 0:
			reserved_x_size = remaining_size as float / remaining_fields
		else: 
			reserved_x_size = 0
			
	_set_field_sizes(target_sizes)

func _set_field_sizes(sizes: Dictionary[SRData.Field, int]) -> void:
	_inner_resize = true

	for elem: SRDataRow in _sr_data.values():
		for field: SRData.Field in sizes.keys():
			elem.set_field_size(field, sizes[field])

	for field: SRData.Field in sizes.keys():
		_data_headers_hbc.set_column_size(field, sizes[field])
	_inner_resize = false

func _set_column_sizes_uniform(uni_size: int) -> void:
	_inner_resize = true
	for elem: SRDataRow in _sr_data.values():
		for field: SRData.Field in visible_columns:
			elem.set_field_size(field, uni_size)

	for field: SRData.Field in visible_columns:
		_data_headers_hbc.set_column_size(field, uni_size)
	_inner_resize = false

func _set_column_visible(column: SRData.Field, is_visible_new: bool, initial: bool = false) -> void:
	if !initial:
		var has_column: bool = visible_columns.has(column)
		
		if has_column == is_visible_new:
			return

	if is_visible_new && !visible_columns.has(column):
		visible_columns.append(column)

	elif !is_visible_new && visible_columns.has(column):
		visible_columns.erase(column)

	for elem: SRDataRow in _sr_data.values():
		elem.set_field_visible(column, is_visible_new)

	_data_headers_hbc.set_column_visible(column, is_visible_new)

	if !initial:
		_update_field_data()

func _on_select_entry(srd: SRData) -> void:
	select_entry.emit(srd)

func _on_jump_to_entry(srd: SRData) -> void:
	jump_entry.emit(srd)

func set_selected(srd: SRData, is_selected_new: bool) -> void:
	if !_sr_data.has(srd):
		return
		
	_sr_data[srd].set_selected(is_selected_new)

func toggle_column_visible(column: SRData.Field) -> void:
	_set_column_visible(column, !visible_columns.has(column))
	_set_column_visible(column, !visible_columns.has(column))
