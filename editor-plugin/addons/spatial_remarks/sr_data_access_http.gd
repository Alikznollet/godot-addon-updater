@tool
class_name SRDataAccessHttp
extends Node

enum Api {
	READ_ALL,
	CREATE,
	UPDATE,
	SAVE_ALL,
	DELETE
}

static var request_in_process: bool = false

static var http_request_node: HTTPRequest # TODO: might need to replace this with httpclient. If so, figure out how to poll

#static var _http_client: HTTPClient
#static var _http_url: String

#static func ensure_http_client_initialized(url: String, port: int = -1, tls_options: TLSOptions = null) -> void:
	#if _http_client != null && _http_url == url:
		#return
		#
	#_http_client = HTTPClient.new()
	#_http_url = url
	#
	#_http_client.connect_to_host(url, port, tls_options)

static func read_all_srd_from_http(config: SRConfigHandler.Config) -> Array[SRData]:
	#if http_request_node == null:
		#push_warning("SRDataHandler: http_request node not initialized. Skip loading SRDs.")
		#return []
		#
	if request_in_process:
		push_warning("SRDataAccessHttp: http_request node is already processing another process.")
	request_in_process = true
	
	var http_url: String = config.data_location
	#ensure_http_client_initialized(http_url)
	
	# TODO: this is a WIP function

	var http_request: String = http_url + config.http_api[Api.READ_ALL]

	var error: Error = http_request_node.request(http_request, [], HTTPClient.METHOD_GET)
	if error != OK:
		push_error("HTTP Request threw an error: ", error)
		request_in_process = false
		return []
				
	var results: Array = await http_request_node.request_completed
	var result_code: int = results[0] as int
	var response_code: int = results[1] as int
	var headers: PackedStringArray = results[2] as PackedStringArray
	var body: PackedByteArray = results[3] as PackedByteArray
	
	if result_code != HTTPRequest.Result.RESULT_SUCCESS:
		push_error("SRDataAccessHttp: HTTP Request result unsuccessful, code is '", result_code, "")#HTTPRequest.Result.keys()[result_code] + "')")
		request_in_process = false
		return []
	
	var json = JSON.new()
	var parse_result: Error = json.parse(body.get_string_from_utf8())
	request_in_process = false
	
	if parse_result != OK:
		push_error("SRDataAccessHttp: JSON Parse Error during http read request: '" + json.get_error_message() + "'  at line " + str(json.get_error_line()))
		return []
		
	var sr_data_from_json: Array[SRData] = str_to_var(json.get_data())
	
	return sr_data_from_json
	
static func create_srd_http(config: SRConfigHandler.Config, sr_data: SRData) -> void:
	#if http_request_node == null:
		#push_warning("SRDataHandler: http_request node not initialized. Skip saving single SRD.")
	#
	if request_in_process:
		push_warning("SRDataAccessHttp: http_request node is already processing another process.")
	request_in_process = true
	
	var http_url: String = config.data_location
	#ensure_http_client_initialized(http_url)

	# TODO: this is a WIP function
	
	var json_string: String = JSON.stringify(var_to_str(sr_data))
	var http_request: String = http_url + config.http_api[Api.CREATE]
	var error: Error = http_request_node.request(http_request, [], HTTPClient.METHOD_POST, json_string)
	if error != OK:
		push_error("HTTP Request '" + http_request +"' threw an error: ", error)
		request_in_process = false
		return

	var results: Array = await http_request_node.request_completed
	var result_code: int = results[0] as int
	var response_code: int = results[1] as int
	var headers: PackedStringArray = results[2] as PackedStringArray
	var body: PackedByteArray = results[3] as PackedByteArray
		
	request_in_process = false



static func update_srd_http(config: SRConfigHandler.Config, sr_data: SRData) -> void:
	#if http_request_node == null:
		#push_warning("SRDataHandler: http_request node not initialized. Skip saving single SRD.")
	#
	if request_in_process:
		push_warning("SRDataAccessHttp: http_request node is already processing another process.")
	request_in_process = true
	
	var http_url: String = config.data_location
	#ensure_http_client_initialized(http_url)

	# TODO: this is a WIP function
	
	var json_string: String = JSON.stringify(var_to_str(sr_data))
	var http_request: String = http_url + config.http_api[Api.UPDATE]
	var error: Error = http_request_node.request(http_request, [], HTTPClient.METHOD_POST, json_string)
	if error != OK:
		push_error("HTTP Request '" + http_request +"' threw an error: ", error)
		request_in_process = false
		return

	var results: Array = await http_request_node.request_completed
	var result_code: int = results[0] as int
	var response_code: int = results[1] as int
	var headers: PackedStringArray = results[2] as PackedStringArray
	var body: PackedByteArray = results[3] as PackedByteArray
		
	request_in_process = false
	
static func _save_all_srd_http(config: SRConfigHandler.Config, sr_data: Array[SRData]) -> void:
	#if http_request_node == null:
		#push_warning("SRDataHandler: http_request node not initialized. Skip saving single SRD.")
	#
	if request_in_process:
		push_warning("SRDataAccessHttp: http_request node is already processing another process.")
	request_in_process = true
	
	var http_url: String = config.data_location
	#ensure_http_client_initialized(http_url)

	# TODO: this is a WIP function
	
	var json_string: String = JSON.stringify(var_to_str(sr_data))
	var http_request: String = http_url + config.http_api[Api.SAVE_ALL]
	var error: Error = http_request_node.request(http_request, [], HTTPClient.METHOD_POST, json_string)
	if error != OK:
		push_error("HTTP Request '" + http_request +"' threw an error: ", error)
		request_in_process = false
		return

	var results: Array = await http_request_node.request_completed
	var result_code: int = results[0] as int
	var response_code: int = results[1] as int
	var headers: PackedStringArray = results[2] as PackedStringArray
	var body: PackedByteArray = results[3] as PackedByteArray
		
	request_in_process = false
	
static func delete_srd_http(config: SRConfigHandler.Config, sr_data: SRData) -> void:
	#if http_request_node == null:
		#push_warning("SRDataHandler: http_request node not initialized. Skip saving single SRD.")
	#
	if request_in_process:
		push_warning("SRDataAccessHttp: http_request node is already processing another process.")
	request_in_process = true
	
	var http_url: String = config.data_location
	#ensure_http_client_initialized(http_url)

	# TODO: this is a WIP function
	
	var json_string: String = JSON.stringify(var_to_str(sr_data.id))
	var http_request: String = http_url + config.http_api[Api.DELETE]
	var error: Error = http_request_node.request(http_request, [], HTTPClient.METHOD_POST, json_string)
	if error != OK:
		push_error("HTTP Request '" + http_request +"' threw an error: ", error)
		request_in_process = false
		return

	var results: Array = await http_request_node.request_completed
	var result_code: int = results[0] as int
	var response_code: int = results[1] as int
	var headers: PackedStringArray = results[2] as PackedStringArray
	var body: PackedByteArray = results[3] as PackedByteArray
		
	request_in_process = false
