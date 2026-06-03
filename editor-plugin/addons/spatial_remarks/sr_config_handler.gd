class_name SRConfigHandler
extends Object

enum DataSourceType {
	JSON,
	HTTP
}

enum NodeReferenceDisplay {
	PATH,
	NAME,
	NONE
}

enum AdditionalDataOption {
	NEVER,
	CHOOSE,
	ALWAYS
}

class Config:
	
	# General
	var author: String
	#var project_name: String #UNUSED
	var project_version: String

	# Data
	var data_source_type: DataSourceType
	var data_location: String
	var screenshot_location: String
	var http_api: Dictionary[SRDataAccessHttp.Api, String]

	# Remark Creation
	var author_editable: bool
	var show_position: bool
	var scene_name_display: NodeReferenceDisplay
	var show_target_node: bool
	var target_node_selectable: bool
	var max_num_query_results: int
	var max_3d_range: float
	var create_screenshots: AdditionalDataOption
	var include_logs: AdditionalDataOption
	var max_log_lines: int
	
	#var create_dbsnapshot: AdditionalDataOption #UNUSED
	#var category_selectable: bool #UNUSED

	# Display
	var collision_layer_number: int
	var use_raycast_for_display: bool
	
	static func get_default() -> Config:
		var default_config: Config = Config.new()
		default_config.author = "unknown"
		#default_config.project_name = "unspecified"
		default_config.project_version = "unspecified"
		default_config.data_source_type = DataSourceType.JSON
		default_config.data_location = "res://spatial_remarks/spatial_remarks.json"
		default_config.screenshot_location = "res://spatial_remarks/"
		default_config.http_api = {
			SRDataAccessHttp.Api.CREATE: "create",
			SRDataAccessHttp.Api.SAVE_ALL: "save_all",
			SRDataAccessHttp.Api.READ_ALL: "read_all",
			SRDataAccessHttp.Api.UPDATE: "update",
			SRDataAccessHttp.Api.DELETE: "delete"
		}
		
		default_config.author_editable = false
		default_config.show_position = false
		default_config.scene_name_display = NodeReferenceDisplay.NONE
		default_config.show_target_node = false
		default_config.target_node_selectable = false
		default_config.max_num_query_results = 20
		default_config.max_3d_range = 20.0
		default_config.create_screenshots = AdditionalDataOption.NEVER
		default_config.include_logs = AdditionalDataOption.NEVER
		default_config.max_log_lines = 50
		#default_config.create_dbsnapshot = AdditionalDataOption.NEVER
		#default_config.category_selectable = false
		default_config.collision_layer_number = 32
		default_config.use_raycast_for_display = false
		return default_config
		
const PROJECT_SETTINGS_IDENTIFIER: String = "[project]"

## The file path for the default configuration below the addon path. This will usually be addons/spatial_remarks/DEFAULT_ADDON_CFG_PATH
const DEFAULT_ADDON_CFG_PATH: String = "/sr_config.cfg"

static func load_config() -> Config:
	# default configuration
	var config: Config = Config.get_default()

	var default_filepath: String = SRDataAccess.get_plugin_path() + DEFAULT_ADDON_CFG_PATH

	var base_config_file: ConfigFile = get_config_file(default_filepath)
	if base_config_file == null:
		push_warning("Could not determine default config file. Using default configuration")
		return config

	# standard configuration
	_update_config_from_cfg_file(config, base_config_file)
	
	# export override
	if !OS.has_feature("editor"):
		var export_cfg_override_path: String = _load_value(base_config_file, "config", "export_override")
		override_config_from_path(config, export_cfg_override_path)
	
	# override
	var cfg_override_path: String = _load_value(base_config_file, "config", "override")
	override_config_from_path(config, cfg_override_path)

	return config
	
static func override_config_from_path(config: Config, override_cfg_path: String) -> void:
	# only override if the path exists
	if !override_cfg_path.strip_edges().is_empty():
		var override_cfg_file: ConfigFile = get_config_file(override_cfg_path)
		if override_cfg_file != null:
			_update_config_from_cfg_file(config, override_cfg_file)
				
static func _update_config_from_cfg_file(config: Config, config_file: ConfigFile) -> void:

	# General
	var author: String = _load_value(config_file, "general", "author")
	if !author.is_empty():
		config.author = author

	#var project_name: String = _load_value(config_file, "general", "project_name", "application/config/name")
	#if !project_name.is_empty():
		#config.project_name = project_name

	var project_version: String = _load_value(config_file, "general", "project_version", "application/config/version")
	if !project_version.is_empty():
		config.project_version = project_version

	# Data		
	var data_source_type: String = _load_value(config_file, "data", "source_type")
	if !data_source_type.is_empty():
		config.data_source_type = _to_data_source_type(data_source_type)
		
	var data_location: String = _load_value(config_file, "data", "location")
	if !data_location.is_empty():
		config.data_location = data_location
		
	var screenshot_location: String = _load_value(config_file, "data", "screenshot_location")
	if !screenshot_location.is_empty():
		config.screenshot_location = screenshot_location
				
	var http_read_all: String = _load_value(config_file, "data", "http_read_all")
	if !http_read_all.is_empty():
		config.http_api[SRDataAccessHttp.Api.READ_ALL] = http_read_all
		
	var http_create: String = _load_value(config_file, "data", "http_create")
	if !http_create.is_empty():
		config.http_api[SRDataAccessHttp.Api.CREATE] = http_create
		
	var http_save_all: String = _load_value(config_file, "data", "http_save_all")
	if !http_save_all.is_empty():
		config.http_api[SRDataAccessHttp.Api.SAVE_ALL] = http_save_all
	
	var http_update: String = _load_value(config_file, "data", "http_update")
	if !http_update.is_empty():
		config.http_api[SRDataAccessHttp.Api.UPDATE] = http_update

	var http_delete: String = _load_value(config_file, "data", "http_delete")
	if !http_delete.is_empty():
		config.http_api[SRDataAccessHttp.Api.DELETE] = http_delete
		
	# Remark Creation
	var author_editable: String = _load_value(config_file, "creation", "author_editable")
	if !author_editable.is_empty():
		config.author_editable = str_to_var(author_editable)

	var show_position: String = _load_value(config_file, "creation", "show_position")
	if !show_position.is_empty():
		config.show_position = str_to_var(show_position)
		
	var scene_name_display: String = _load_value(config_file, "creation", "scene_name_display")
	if !scene_name_display.is_empty():
		config.scene_name_display = _to_node_reference_display(scene_name_display)
		
	var show_target_node: String = _load_value(config_file, "creation", "show_target_node")
	if !show_target_node.is_empty():
		config.show_target_node = str_to_var(show_target_node)

		
	var target_node_selectable: String = _load_value(config_file, "creation", "target_node_selectable")
	if !target_node_selectable.is_empty():
		config.target_node_selectable = str_to_var(target_node_selectable)
				
	var max_num_query_results: String = _load_value(config_file, "creation", "max_num_query_results")
	if !max_num_query_results.is_empty():
		config.max_num_query_results = str_to_var(max_num_query_results)

	var max_3d_range: String = _load_value(config_file, "creation", "max_3d_range")
	if !max_3d_range.is_empty():
		config.max_3d_range = str_to_var(max_3d_range)

	var create_screenshots: String = _load_value(config_file, "creation", "create_screenshots")
	if !create_screenshots.is_empty():
		config.create_screenshots = _to_additional_data_option(create_screenshots)
	
	var include_logs: String = _load_value(config_file, "creation", "include_logs")
	if !include_logs.is_empty():
		config.include_logs = _to_additional_data_option(include_logs)
		
	var max_log_lines: String = _load_value(config_file, "creation", "max_log_lines")
	if !max_log_lines.is_empty():
		config.max_log_lines = str_to_var(max_log_lines)
	
	#var create_dbsnapshot: String = _load_value(config_file, "creation", "create_dbsnapshot")
	#if !create_dbsnapshot.is_empty():
		#config.create_dbsnapshot = _to_additional_data_option(create_dbsnapshot)	
	
	#var category_selectable: String = _load_value(config_file, "creation", "category_selectable")
	#if !category_selectable.is_empty():
		#config.category_selectable = str_to_var(category_selectable)	
	
	# Display
	var collision_layer_number: String = _load_value(config_file, "display", "collision_layer_number")
	if !collision_layer_number.is_empty():
		config.collision_layer_number = str_to_var(collision_layer_number)

	var use_raycast_for_display: String = _load_value(config_file, "display", "use_raycast_for_display")
	if !use_raycast_for_display.is_empty():
		config.use_raycast_for_display = str_to_var(use_raycast_for_display)

static func _load_value(config: ConfigFile, config_section: String, config_value: String, project_settings_path: String = "") -> String:
	var cfg_value: Variant = config.get_value(config_section, config_value, "")
	if !cfg_value is String:
		cfg_value = var_to_str(config.get_value(config_section, config_value, ""))
	if cfg_value.strip_edges().is_empty():
		return ""
		
	if !project_settings_path.is_empty() && cfg_value.strip_edges().to_lower() == PROJECT_SETTINGS_IDENTIFIER:
		var project_settings_value: String = ProjectSettings.get_setting(project_settings_path, "") as String
		if project_settings_value.strip_edges().is_empty():
			return ""
		
		cfg_value = project_settings_value
	return cfg_value

static func _to_data_source_type(dst_string: String) -> DataSourceType:
	if dst_string.to_lower().strip_edges() == "json":
		return DataSourceType.JSON
	if dst_string.to_lower().strip_edges() == "http":
		return DataSourceType.HTTP
		
	push_warning("SRDataHandler: DataSourceType '", dst_string, "' not recognized. Defaulting to JSON")
	return DataSourceType.JSON

static func get_config_file(path: String) -> ConfigFile:
	if !FileAccess.file_exists(path):
		push_warning("SRDataHandler: Could not load SR configuration file from '", path, "', file does not exist.")
		return null

	var config_file: ConfigFile = ConfigFile.new()
	var status: Error = config_file.load(path)
	if status != OK:
		push_warning("SRDataHandler: Could not load SR configuration file from '", path, "', error was: ", status)
		return null
	
	return config_file

static func _to_node_reference_display(nrd_string: String) -> NodeReferenceDisplay:
	if nrd_string.to_lower().strip_edges() == "path":
		return NodeReferenceDisplay.PATH
	elif nrd_string.to_lower().strip_edges() == "name":
		return NodeReferenceDisplay.NAME

	return NodeReferenceDisplay.NONE
	
static func _to_additional_data_option(ado_string: String) -> AdditionalDataOption:
	if ado_string.to_lower().strip_edges() == "always":
		return AdditionalDataOption.ALWAYS
	elif ado_string.to_lower().strip_edges() == "choose":
		return AdditionalDataOption.CHOOSE

	return AdditionalDataOption.NEVER
