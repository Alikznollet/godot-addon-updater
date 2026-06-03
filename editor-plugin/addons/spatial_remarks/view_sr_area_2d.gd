class_name ViewSRArea2D extends Area2D

var current_sr_note: SRNote2D = null

func _ready() -> void:
	var col_layer: int = SRHandler.get_collision_layer_number()
	collision_mask = 0
	if col_layer > 0:
		set_collision_mask_value(col_layer, true)
		

func _physics_process(delta: float) -> void:
	if !SRHandler.remarks_visible:
		return
		
	var overlapping_areas: Array = get_overlapping_areas()
	if overlapping_areas.size() > 0:
		overlapping_areas.sort_custom(sort_by_dist)
		var best_area: Area2D = overlapping_areas[0]
		if best_area is SRNote2D:
			var target_sr_note: SRNote2D = (best_area as SRNote2D)
			if target_sr_note == current_sr_note:
				return
			unhighlight_previous()
			current_sr_note = (best_area as SRNote2D)
			current_sr_note.set_highlighted(true)
			return

	unhighlight_previous()
	
func unhighlight_previous() -> void:
	if current_sr_note == null:
		return
	
	current_sr_note.set_highlighted(false)
	current_sr_note = null

func sort_by_dist(a: Area2D, b: Area2D) -> bool:
	return a.global_position.distance_squared_to(self.global_position) < b.global_position.distance_squared_to(self.global_position)
