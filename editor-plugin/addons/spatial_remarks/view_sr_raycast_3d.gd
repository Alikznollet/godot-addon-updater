class_name ViewSRRaycast3D extends RayCast3D

var current_sr_note: SRNote3D = null

func _ready() -> void:
	var col_layer: int = SRHandler.get_collision_layer_number()
	collision_mask = 0
	if col_layer > 0:
		set_collision_mask_value(col_layer, true)
		

func _physics_process(delta: float) -> void:
	if !SRHandler.remarks_visible:
		return
	
	if is_colliding():
		var collider: Node3D = get_collider()
		if collider is SRNote3D:
			var target_sr_note: SRNote3D = (collider as SRNote3D)
			if target_sr_note == current_sr_note:
				return
			unhighlight_previous()
			current_sr_note = (collider as SRNote3D)
			current_sr_note.set_highlighted(true)
			return

	unhighlight_previous()
	
func unhighlight_previous() -> void:
	if current_sr_note == null:
		return
	
	current_sr_note.set_highlighted(false)
	current_sr_note = null
