@tool
class_name SRData
extends RefCounted

enum Field {
	AUTHOR,
	PROJECT_VERSION,
	SCENE,
	TARGET_NODE,
	GLOBAL_POSITION,
	CREATION_DATE,
	TEXT,
	LOGS,
	SCREENSHOT
}

const NO_ID: int = -1

var id: int = NO_ID
var author: String
#var project_name: String -> do we really need this?
var project_version: String
var scene: String # saved as path
var target_node: String # saved as relative path from the scene
var text: String
var global_position: Vector3 # maybe used as Container for Vector2, if is_2d is true 
var creation_date: String # this is stored in UTC but displayed in the local timezone
#var context: Dictionary
#var category: Category
var is_2d: bool = false
var log_lines: String = ""
var screenshot_name: String = ""

func duplicate() -> SRData:
	var res: SRData = SRData.new()
	res.author = self.author
	#res.project_name = self.project_name
	res.project_version = self.project_version
	res.scene = self.scene
	res.target_node = self.target_node
	res.text = self.text
	res.global_position = self.global_position
	res.creation_date = self.creation_date
	res.is_2d = self.is_2d
	res.log_lines = self.log_lines
	res.screenshot_name = self.screenshot_name
	#res.context = self.context.duplicate()
	#res.category = self.category
	return res

func update_from_reference(srd: SRData) -> void:
	self.author = srd.author
	self.project_version = srd.project_version
	self.scene = srd.scene
	self.target_node = srd.target_node
	self.text = srd.text
	self.global_position = srd.global_position
	self.creation_date = srd.creation_date
	self.is_2d = srd.is_2d
	self.log_lines = srd.log_lines
	self.screenshot_name = srd.screenshot_name

func get_field_value(field: Field) -> Variant:
	match field:
		Field.AUTHOR: return author
		Field.PROJECT_VERSION: return project_version
		Field.SCENE: return scene
		Field.TARGET_NODE: return target_node
		Field.GLOBAL_POSITION: return global_position
		Field.CREATION_DATE: return creation_date
		Field.TEXT: return text
		Field.LOGS: return log_lines
		Field.SCREENSHOT: return screenshot_name
		
	push_warning("SRData: no handling for ", Field.keys()[field], " implemented. Returning null.")
	return null

func get_global_position_tr(snap_steps: float = 0.01) -> String:
	var res: String = "(" + str(snappedf(global_position.x, snap_steps)) + ", " + str(snappedf(global_position.y, snap_steps))
	if !is_2d:
		res += ", " + str(snappedf(global_position.z, snap_steps))
	return res + ")"
	
func get_creation_date(use_local_time: bool = true) -> String:
	if !use_local_time:
		return creation_date.replace("T", " ") + " (UTC)"

	var time_zone_info: Dictionary = Time.get_time_zone_from_system()
	var bias_minutes: int = time_zone_info["bias"]
	var timezone_name: String = time_zone_info["name"]
	
	var date_time_sec: int = Time.get_unix_time_from_datetime_string(creation_date)
	date_time_sec += bias_minutes * 60 # convert to seconds
	var datetime_str: String = Time.get_datetime_string_from_unix_time(date_time_sec, true)
	
	return datetime_str + " ("+timezone_name+")"

func get_log_lines_tr() -> String:
	var first_log_line: String = log_lines.substr(0, log_lines.find("\n"))
	if first_log_line.length() > log_lines.length():
		first_log_line += "+"
	return first_log_line
	
func get_field_value_str(field: Field) -> Variant:
	match field:
		Field.GLOBAL_POSITION: return get_global_position_tr()
		Field.CREATION_DATE: return get_creation_date()
		#Field.LOGS: return get_log_lines_tr()

	return get_field_value(field)
	
func get_field_value_for_column(field: Field) -> String:
	match field:
		Field.SCREENSHOT: return get_exists_str(!screenshot_name.is_empty())
		Field.LOGS: return get_exists_str(!log_lines.strip_edges().is_empty())
	
	return get_field_value_str(field)
	
static func get_exists_str(exists: bool) -> String:
	return "yes" if exists else "-"
	
func set_field_from_str(field: Field, value: String) -> void:
	match field:
		Field.AUTHOR: author = value as String
		Field.PROJECT_VERSION: project_version = value as String
		Field.SCENE: scene = value as String
		Field.TARGET_NODE: target_node = value as String
		Field.GLOBAL_POSITION: set_global_position(value)
		#Field.CREATION_DATE: creation_date = value as String
		Field.TEXT: text = value as String
		_: push_warning("SRData: no handling for ", Field.keys()[field], " implemented")


func set_creation_time(value: String) -> void:
	# if allowed, timezone information would have to be extracted here.
	# possible solution: find the difference to get_field_value_str(Field.CREATION_DATE) and apply it to 
	push_warning("SRData: no handling for creation date settings implemented")
	#creation_date = ....

func set_global_position(value: String) -> void:
	# if allowed, timezone information would have to be extracted here.
	# possible solution: find the difference to get_field_value_str(Field.CREATION_DATE) and apply it to 
	push_warning("SRData: no handling for global position settings implemented")
	#global_position = ...

static func get_field_name(field: Field) -> String:
	match field:
		Field.PROJECT_VERSION: return "Version"
	
	return (Field.keys()[field] as String).to_pascal_case()

## if false, this field cannot be edited in the editor dock
static func is_editable_in_editor(field: Field) -> bool:
	match field:
		Field.CREATION_DATE: return false
		Field.GLOBAL_POSITION: return false
		Field.LOGS: return false
		Field.SCREENSHOT: return false

	return true

## if false, this field cannot be used/selected as column in the editor dock
static func is_column_in_editor(field: Field) -> bool:
	#match field:
		#Field.LOGS: return false
		#Field.SCREENSHOT: return false
		#Field.PROJECT_VERSION: return false
		
	return true
