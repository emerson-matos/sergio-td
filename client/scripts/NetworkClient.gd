class_name NetworkClient
extends Node

signal connected
signal connection_failed(reason: String)
signal snapshot_received(data: Dictionary)
signal game_data_received(data: Dictionary)
signal wave_started(wave_number: int)
signal wave_completed(wave_number: int)
signal match_ended(victory: bool, wave_number: int)
signal command_accepted(command_id: String, data: Dictionary)
signal command_rejected(command_id: String, reason: String)

var _socket := WebSocketPeer.new()
var _is_connecting := false
var _player_id := ""
var _connected := false

var waypoints: Array = []
var tower_types: Dictionary = {}
var enemy_types: Dictionary = {}
var debug_enabled := false

func connect_to_server(url: String) -> void:
	var err := _socket.connect_to_url(url)
	if err != OK:
		emit_signal("connection_failed", "erro ao iniciar websocket: %d" % err)
		return
	_is_connecting = true
	set_process(true)

func _process(_delta: float) -> void:
	_socket.poll()
	var state := _socket.get_ready_state()
	if _is_connecting and state == WebSocketPeer.STATE_OPEN:
		_is_connecting = false
		_connected = true
		emit_signal("connected")
		_player_id = _get_player_id_from_args()
		_send_hello()
	elif _is_connecting and state == WebSocketPeer.STATE_CLOSED:
		_is_connecting = false
		_connected = false
		emit_signal("connection_failed", "servidor fechou conexão")

	while _socket.get_available_packet_count() > 0:
		var packet := _socket.get_packet().get_string_from_utf8()
		_handle_server_message(packet)

func _handle_server_message(raw_packet: String) -> void:
	var parsed: Variant = JSON.parse_string(raw_packet)
	if parsed == null or not (parsed is Dictionary):
		return

	var msg_type: String = parsed.get("type", "")
	var payload: Dictionary = parsed.get("payload", {})
	
	if debug_enabled:
		print("[<<] ", msg_type, ": ", payload)
	
	match msg_type:
		"ACK_HELLO":
			waypoints = payload.get("waypoints", [])
			tower_types = payload.get("towerTypes", {})
			enemy_types = payload.get("enemyTypes", {})
			emit_signal("game_data_received", {
				"waypoints": waypoints,
				"towerTypes": tower_types,
				"enemyTypes": enemy_types,
				"matchId": payload.get("matchId", "")
			})
			
		"SNAPSHOT_STATE":
			var data: Dictionary = {
				"tick": payload.get("tick", 0),
				"waveNumber": payload.get("waveNumber", 0),
				"players": payload.get("players", []),
				"towers": payload.get("towers", []),
				"enemies": payload.get("enemies", [])
			}
			emit_signal("snapshot_received", data)
			
		"EVENT_WAVE_STARTED":
			var wave_num: int = payload.get("waveNumber", 1)
			emit_signal("wave_started", wave_num)
			
		"EVENT_WAVE_COMPLETED":
			var wave_num: int = payload.get("waveNumber", 1)
			emit_signal("wave_completed", wave_num)
			
		"EVENT_MATCH_ENDED":
			var victory: bool = payload.get("victory", false)
			var wave_num: int = payload.get("waveNumber", 0)
			emit_signal("match_ended", victory, wave_num)
			
		"EVENT_MATCH_STARTED":
			# Match started - waves will begin
			pass
			
		"LOBBY_STATE":
			pass

func _send_hello() -> void:
	_send_message("HELLO", {
		"playerId": _player_id
	})

func send_start_wave() -> void:
	_send_message("START_WAVE", {})

func send_player_ready() -> void:
	_send_message("PLAYER_READY", {})

func send_place_tower(tower_type: String, x: float, y: float, command_id: String = "") -> void:
	var timestamp = int(Time.get_unix_time_from_system() * 1000)
	var cmd_id = command_id if command_id != "" else "cmd_place_" + str(timestamp)
	_send_message("COMMAND_PLACE_TOWER", {
		"commandId": cmd_id,
		"towerType": tower_type,
		"x": x,
		"y": y
	})

func send_upgrade_tower(tower_id: String, command_id: String = "") -> void:
	_send_message("COMMAND_UPGRADE_TOWER", {
		"commandId": command_id if command_id != "" else "cmd_upgrade_%d" % Time.get_unix_time_from_system(),
		"towerId": tower_id
	})

func send_set_target(tower_id: String, target_mode: String, command_id: String = "") -> void:
	_send_message("COMMAND_SET_TARGET", {
		"commandId": command_id if command_id != "" else "cmd_target_%d" % Time.get_unix_time_from_system(),
		"towerId": tower_id,
		"targetMode": target_mode
	})

func send_sell_tower(tower_id: String, command_id: String = "") -> void:
	_send_message("COMMAND_SELL_TOWER", {
		"commandId": command_id if command_id != "" else "cmd_sell_%d" % Time.get_unix_time_from_system(),
		"towerId": tower_id
	})

func send_get_game_data() -> void:
	_send_message("GET_GAME_DATA", {})

func _get_player_id_from_args() -> String:
	for arg in OS.get_cmdline_user_args():
		if arg.begins_with("--player-id="):
			return arg.substr("--player-id=".length())
	return "p_%s" % OS.get_unique_id().replace("-", "_").substr(0, 8)

func get_player_id() -> String:
	return _player_id

func is_server_connected() -> bool:
	return _connected

func _send_message(message_type: String, payload: Dictionary) -> void:
	if not _connected:
		return
	
	if debug_enabled:
		print("[>>] ", message_type, ": ", payload)
		
	var msg := {
		"v": 1,
		"type": message_type,
		"ts": Time.get_unix_time_from_system(),
		"payload": payload
	}
	_socket.send_text(JSON.stringify(msg))
