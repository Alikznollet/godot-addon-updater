@tool
class_name SRDataAccess
extends Node

## Note: this is computed from the asset folder. Retrieve via get_plugin_path()
static var plugin_path: String = ""
static var shown_res_path_warning: bool = false
static var request_in_process: bool = false

static var id_gen_counter: int = 0

static var cached_srd_data: Array[SRData]
static var cached_config: SRConfigHandler.Config
static var cache_dirty: bool = true

static func prepare_path(target_path: String) -> bool:
	if target_path.begins_with("res://") && !shown_res_path_warning:
		push_warning("SRDataAccess: accessing paths at 'res://' may not work properly for exported builds (used path: '", target_path, "')")
		shown_res_path_warning = true
		
	var dir_end_idx: int = max(target_path.rfind("\\"), target_path.rfind("/")) + 1
	var target_dir: String = target_path.substr(0, dir_end_idx)
	
	if !DirAccess.dir_exists_absolute(target_dir):
		DirAccess.make_dir_recursive_absolute(target_dir)
		
	return true

static func get_plugin_path() -> String:
	if plugin_path.is_empty():
		plugin_path = (new().get_script() as Script).resource_path.get_base_dir()
	return plugin_path

static func read_all_sr_data(config: SRConfigHandler.Config) -> Array[SRData]:
	var current_sr_data: Array[SRData] = []
	match config.data_source_type:
		SRConfigHandler.DataSourceType.JSON: current_sr_data = read_all_srd_from_json_file(config)
		SRConfigHandler.DataSourceType.HTTP: current_sr_data = await SRDataAccessHttp.read_all_srd_from_http(config)

		_: push_warning("SRDataAccess: No handling implemented for data source type '", SRConfigHandler.DataSourceType.keys()[config.data_source_type], "'. Saving data ignored.")
	return current_sr_data

static func create_srd(config: SRConfigHandler.Config, srd: SRData) -> void:
	match config.data_source_type:
		SRConfigHandler.DataSourceType.JSON: create_srd_json_file(config, srd)
		SRConfigHandler.DataSourceType.HTTP: SRDataAccessHttp.create_srd_http(config, srd)
		_: push_warning("SRDataAccess: No handling implemented for data source type '", SRConfigHandler.DataSourceType.keys()[config.data_source_type], "'. Saving data ignored.")

static func update_srd(config: SRConfigHandler.Config, srd: SRData) -> void:
	match config.data_source_type:
		SRConfigHandler.DataSourceType.JSON: update_srd_json_file(config, srd)
		SRConfigHandler.DataSourceType.HTTP: SRDataAccessHttp.update_srd_http(config, srd)
		_: push_warning("SRDataAccess: No handling implemented for data source type '", SRConfigHandler.DataSourceType.keys()[config.data_source_type], "'. Saving data ignored.")

static func save_all_srd(config: SRConfigHandler.Config, sr_data: Array[SRData]) -> void:
	match config.data_source_type:
		SRConfigHandler.DataSourceType.JSON: save_all_srd_json_file(config, sr_data)
		SRConfigHandler.DataSourceType.HTTP: SRDataAccessHttp._save_all_srd_http(config, sr_data)
		_: push_warning("SRDataAccess: No handling implemented for data source type '", SRConfigHandler.DataSourceType.keys()[config.data_source_type], "'. Saving data ignored.")

static func delete_srd(config: SRConfigHandler.Config, srd: SRData) -> void:
	match config.data_source_type:
		SRConfigHandler.DataSourceType.JSON: delete_srd_json_file(config, srd)
		SRConfigHandler.DataSourceType.HTTP: SRDataAccessHttp.delete_srd_http(config, srd)
		_: push_warning("SRDataAccess: No handling implemented for data source type '", SRConfigHandler.DataSourceType.keys()[config.data_source_type], "'. Saving data ignored.")

#TODO: remarks within the project are not included in the export due to default settings.
static func read_all_srd_from_json_file(config: SRConfigHandler.Config) -> Array[SRData]:
	var json_path: String = config.data_location
	
	if json_path.begins_with("res://") && !shown_res_path_warning:
		push_warning("SRDataAccess: accessing paths at 'res://' may not work properly for exported builds (used path: '", json_path, "')")
		shown_res_path_warning = true

	if !FileAccess.file_exists(json_path):
		print("SRDataAccess: No spatial remark data found at '", json_path ,"'")
		return []# We don't have a file to load.
	
	var json_file_access: FileAccess = FileAccess.open(json_path, FileAccess.READ)
	
	var all_srds: Array[SRData] = _do_read_all_srds(json_file_access, config)
	json_file_access.close()
	print("SRDataAccess: loaded ", all_srds.size(), " spatial remark" + ("s" if all_srds.size() != 1 else "") + ".")
	
	return all_srds
	
static func import_srds_from_json_file(json_path: String) -> Array[SRData]:
	if !FileAccess.file_exists(json_path):
		print("SRDataAccess: No spatial remark data found at '", json_path ,"'")
		return []# We don't have a file to load.
	
	var json_file_access: FileAccess = FileAccess.open(json_path, FileAccess.READ)
	
	var all_srds: Array[SRData] = _do_read_all_srds(json_file_access, null)
	json_file_access.close()
	print("SRDataAccess: imported ", all_srds.size(), " spatial remark" + ("s" if all_srds.size() != 1 else "") + ".")
	
	return all_srds
	
static func create_srd_json_file(config: SRConfigHandler.Config, srd: SRData) -> void:
	if srd.id != SRData.NO_ID:
		push_error("SRData to be created already has an id: ", srd.id, ". Stopped saving..")
		return
	
	var target_path: String = config.data_location
	var path_available = prepare_path(target_path)
	if !path_available:
		push_warning("SRDataAccess: could not save to json file")
		return
			
	var all_srds: Array[SRData] = []
	var json_file_access: FileAccess
		
	if !FileAccess.file_exists(target_path):
		print("SRDataAccess: No spatial remark data found at '", target_path ,"'")
		json_file_access = FileAccess.open(config.data_location, FileAccess.WRITE)
	else: 
		json_file_access = FileAccess.open(config.data_location, FileAccess.READ_WRITE)
		all_srds = _do_read_all_srds(json_file_access, config)

	if json_file_access == null:
		var open_error: Error = FileAccess.get_open_error()
		push_warning("SRDataAccess: could not open file, open error was: ", open_error)

	_create_new_id(srd, all_srds)
	all_srds.append(srd)
	var json_string: String = JSON.stringify(var_to_str(all_srds))
	json_file_access.seek(0)
	json_file_access.store_line(json_string)
	json_file_access.resize(json_file_access.get_position())
	json_file_access.close()
	cache_dirty = true
	
static func update_srd_json_file(config: SRConfigHandler.Config, srd: SRData) -> void:
	if srd.id == SRData.NO_ID:
		push_error("could not update srd because it has no id!")
		return
	
	var target_path: String = config.data_location
	var path_available = prepare_path(target_path)
	if !path_available:
		push_warning("SRDataAccess: could not save to json file")
		return
			
	var json_file_access: FileAccess = FileAccess.open(config.data_location, FileAccess.READ_WRITE)
	var all_srds: Array[SRData] = _do_read_all_srds(json_file_access, config)

	var updated: bool = false
	for existing_idx: int in all_srds.size():
		var existing_srd: SRData = all_srds[existing_idx]
		if existing_srd.id == srd.id:
			all_srds[existing_idx] = srd
			updated = true
			break
			
	if !updated:
		push_warning("could not find srd id '" + str(srd.id) + "' for update in SRDataAccess!")
		json_file_access.close()
		return
		
	var json_string: String = JSON.stringify(var_to_str(all_srds))
	json_file_access.seek(0)
	json_file_access.store_line(json_string)
	json_file_access.resize(json_file_access.get_position())
	json_file_access.close()
	cache_dirty = true

static func save_all_srd_json_file(config: SRConfigHandler.Config, sr_data: Array[SRData]) -> void:
	var target_path: String = config.data_location
	var path_available = prepare_path(target_path)
	if !path_available:
		push_warning("SRDataAccess: could not save to json file")
		return
		
	for data: SRData in sr_data:
		if data.id == SRData.NO_ID:
			_create_new_id(data, sr_data)
	
	var json_file_access: FileAccess = FileAccess.open(config.data_location, FileAccess.WRITE)

	var json_string: String = JSON.stringify(var_to_str(sr_data))
	json_file_access.store_line(json_string)
	json_file_access.close()
	cache_dirty = true

static func delete_srd_json_file(config: SRConfigHandler.Config, srd: SRData) -> void:
	if srd.id == SRData.NO_ID:
		push_error("could not delete srd because it has no id!")
		return
	
	var target_path: String = config.data_location
	var path_available = prepare_path(target_path)
	if !path_available:
		push_warning("SRDataAccess: could not remove srd from json file")
		return
			
	var json_file_access: FileAccess = FileAccess.open(config.data_location, FileAccess.READ_WRITE)
	var all_srds: Array[SRData] = _do_read_all_srds(json_file_access, config)

	var removed: bool = false
	for existing_idx: int in all_srds.size():
		var existing_srd: SRData = all_srds[existing_idx]
		if existing_srd.id == srd.id:
			all_srds.remove_at(existing_idx)
			removed = true
			break
			
	if !removed:
		push_warning("could not delete given srd ", srd, " because it could not be found.")

	id_gen_counter = max(0, min(id_gen_counter, srd.id))

	var json_string: String = JSON.stringify(var_to_str(all_srds))
	json_file_access.seek(0)
	json_file_access.store_line(json_string)
	json_file_access.resize(json_file_access.get_position())
	json_file_access.close()
	cache_dirty = true

static func _do_read_all_srds(json_file_access: FileAccess, config: SRConfigHandler.Config) -> Array[SRData]:
	if config != null && config == cached_config && !cache_dirty:
		return cached_srd_data
	var all_srds: Array[SRData] = []
	while json_file_access.get_position() < json_file_access.get_length():
		var json_string: String = json_file_access.get_line()
		var json: JSON = JSON.new()
		var parse_result: Error = json.parse(json_string)
		if parse_result != OK:
			push_warning("SRDataAccess: JSON Parse Error: '" + json.get_error_message() + "'  at line " + str(json.get_error_line()))
			continue
		
		var sr_data_from_json: Array[SRData] = str_to_var(json.get_data())
		all_srds.append_array(sr_data_from_json)
	
	cached_config = config
	cache_dirty = false
	cached_srd_data = all_srds
	return all_srds
	
# sorts given array by id - in place
static func sort_by_id(srds: Array[SRData]) -> void:
	srds.sort_custom(_sort_by_id)


# Simple way for creating new ids. Checks for existence of ids in the given set, only creates non-negative ids
# Ids may be reused.
# Probably not a good way for HTTP interop
static func _create_new_id(srd: SRData, known_srds: Array[SRData]) -> void:
	if srd.id != SRData.NO_ID:
		push_warning("SRData already has an id: ", srd.id)
	
	var known_ids: Array[int]
	known_ids.assign(known_srds.map(_get_sr_id))
	
	while true:
		if !known_ids.has(id_gen_counter):
			break
		id_gen_counter += 1
		
	srd.id = id_gen_counter
	id_gen_counter += 1

static func _sort_by_id(a: SRData, b: SRData) -> bool:
	return a.id < b.id

static func _get_sr_id(srd: SRData) -> int:
	return srd.id
