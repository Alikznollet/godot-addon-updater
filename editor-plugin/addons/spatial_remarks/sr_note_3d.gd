@tool
class_name SRNote3D
extends Area3D

@export var _note_empty: CompressedTexture2D
@export var _note_visible: CompressedTexture2D
@export var _sprite_3d: Sprite3D

var _sr_data: SRData

var is_in_editor: bool = true

var normal_scale: Vector3
var highlighted_scale: Vector3

func init(sr_data: SRData, target_collision_layer_number: int, is_in_editor_new: bool = false) -> void:
	is_in_editor = is_in_editor_new
	_sr_data = sr_data	
	global_position = sr_data.global_position
	collision_layer = 0
	normal_scale = scale
	highlighted_scale = scale * 1.5
	if target_collision_layer_number > 0:
		set_collision_layer_value(target_collision_layer_number, true)
	#modulate = SRHandler.get_color_by_category(sr_data.category)

func set_highlighted(is_highlighted: bool) -> void:
	_sprite_3d.texture = _note_empty if is_highlighted else _note_visible
	scale = highlighted_scale if is_highlighted else normal_scale
	if is_highlighted:
		if !is_in_editor:
			SRHandler.show_sr_note(_sr_data)
	else:
		if !is_in_editor:
			SRHandler.hide_sr_note()

func _on_mouse_entered() -> void:
	set_highlighted(true)

func _on_mouse_exited() -> void:
	set_highlighted(false)
