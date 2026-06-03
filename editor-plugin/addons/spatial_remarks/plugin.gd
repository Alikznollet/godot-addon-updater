@tool
extends EditorPlugin

const PLUGIN_NAME: String = "spatial_remarks"

## Autoloads used by the plugin
static var AUTOLOADS: Dictionary[String,String] = {
	SRDataAccess.get_plugin_path() + "/sr_handler.gd": "SRHandler",
}

static func get_inputs() -> Dictionary[String, Array]:
	return {
	"create_sr": [input_key(KEY_F3)],
	"show_sr": [input_key(KEY_F4)],
}

func _enable_plugin() -> void:
	add_autoloads()
	add_inputs()
	EditorInterface.set_plugin_enabled("spatial_remarks/sr_dock", true)
	print("Spatial Remarks enabled.")

	ProjectSettings.save()
	EditorInterface.restart_editor(true)
	
func _disable_plugin() -> void:
	remove_autoloads()
	remove_inputs()
	EditorInterface.set_plugin_enabled("spatial_remarks/sr_dock", false)
	print("Spatial Remarks disabled.")

	ProjectSettings.save()
	EditorInterface.restart_editor(true)

func _enter_tree() -> void:
	if Engine.is_editor_hint():
		pass

func _exit_tree() -> void:
	if Engine.is_editor_hint():
		pass

func add_autoloads() -> void:
	print("Adding autoloads...")
	for autoload_path: String in AUTOLOADS.keys():
		add_autoload_singleton(AUTOLOADS[autoload_path], autoload_path)

func remove_autoloads() -> void:
	print("Removing autoloads...")
	for autoload_name: String in AUTOLOADS.values():
		remove_autoload_singleton(autoload_name)

func add_inputs() -> bool:
	print("Adding inputs...")
	var is_any_changed: bool = false
	var added_inputs: Array[String] = []
	var inputs: Dictionary[String, Array] = get_inputs()
	for input_name in inputs:
		if ProjectSettings.has_setting("input/" + input_name):
			print("add_inputs(): ProjectSettings already has an input '", input_name, "'. Skipped adding this input.")
			continue
		var input: Dictionary = {
			"deadzone": 0.5,
			"events": inputs[input_name]
		}
		ProjectSettings.set_setting("input/" + input_name, input)
		added_inputs.append("'" + input_name + "'")
		is_any_changed = true
	
	if is_any_changed:
		print("Added the following inputs to the input map: " + ", ".join(added_inputs), ". You may need to restart Godot to see these changes.")
		
	return is_any_changed

func remove_inputs() -> bool:
	print("Removing inputs...")
	var is_any_changed: bool = false
	var removed_inputs: Array[String] = []
	for input_name in get_inputs():
		if !ProjectSettings.has_setting("input/" + input_name):
			print("remove_inputs(): ProjectSettings does not have input '", input_name, "'. Skipped removing this input.")
			continue
			
		ProjectSettings.set_setting("input/" + input_name, null)
		removed_inputs.append("'" + input_name + "'")
		is_any_changed = true
		
	if is_any_changed:
		print("Removed the following inputs from the input map: " + ", ".join(removed_inputs), ". You may need to restart Godot to see these changes.")
				
	return is_any_changed


static func input_key(key_code: Key) -> InputEventKey:
	var in_event_key: InputEventKey = InputEventKey.new()
	in_event_key.keycode = key_code
	return in_event_key
