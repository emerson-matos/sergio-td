class_name NetworkClient
extends Node

signal connected
signal connection_failed(reason: String)
signal snapshot_received(tick: int, enemies_count: int, towers_count: int)

var _socket := WebSocketPeer.new()
var _is_connecting := false
var _placed_tower_id := ""
var _upgrade_sent := false
var _sell_sent := false

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
		emit_signal("connected")
		_send_hello()
		_send_start_match()
	elif _is_connecting and state == WebSocketPeer.STATE_CLOSED:
		_is_connecting = false
		emit_signal("connection_failed", "servidor fechou conexão")

	while _socket.get_available_packet_count() > 0:
		var packet := _socket.get_packet().get_string_from_utf8()
		_handle_server_message(packet)

func _handle_server_message(raw_packet: String) -> void:
	var parsed: Variant = JSON.parse_string(raw_packet)
	if parsed == null or not (parsed is Dictionary):
		print("[server] mensagem inválida: %s" % raw_packet)
		return

	var message_type := String(parsed.get("type", ""))
	if message_type == "SNAPSHOT_STATE":
		var payload: Dictionary = parsed.get("payload", {})
		var tick := int(payload.get("tick", 0))
		var enemies: Array = payload.get("enemies", [])
		var towers: Array = payload.get("towers", [])
		emit_signal("snapshot_received", tick, enemies.size(), towers.size())
	elif message_type == "EVENT_WAVE_STARTED":
		_send_place_tower()
	elif message_type == "ACK_COMMAND":
		_handle_ack_command(parsed.get("payload", {}))

	print("[server] %s" % raw_packet)

func _handle_ack_command(payload_variant: Variant) -> void:
	if not (payload_variant is Dictionary):
		return
	var payload: Dictionary = payload_variant
	var accepted := bool(payload.get("accepted", false))
	if not accepted:
		return

	var tower_id := String(payload.get("towerId", ""))
	if tower_id != "":
		_placed_tower_id = tower_id

	if _placed_tower_id == "":
		return

	if not _upgrade_sent:
		_upgrade_sent = true
		_send_upgrade_tower(_placed_tower_id)
		return

	if not _sell_sent:
		_sell_sent = true
		_send_sell_tower(_placed_tower_id)

func _send_hello() -> void:
	_send_message("HELLO", {
		"client": "godot-week2"
	})

func _send_start_match() -> void:
	_send_message("START_MATCH", {
		"mode": "solo-dev"
	})

func _send_place_tower() -> void:
	_send_message("COMMAND_PLACE_TOWER", {
		"commandId": "cmd_place_1",
		"playerId": "p_1",
		"towerType": "dart",
		"x": 6.0,
		"y": 3.0
	})

func _send_upgrade_tower(tower_id: String) -> void:
	_send_message("COMMAND_UPGRADE_TOWER", {
		"commandId": "cmd_upgrade_1",
		"playerId": "p_1",
		"towerId": tower_id
	})

func _send_sell_tower(tower_id: String) -> void:
	_send_message("COMMAND_SELL_TOWER", {
		"commandId": "cmd_sell_1",
		"playerId": "p_1",
		"towerId": tower_id
	})

func _send_message(message_type: String, payload: Dictionary) -> void:
	var msg := {
		"v": 1,
		"type": message_type,
		"ts": Time.get_unix_time_from_system(),
		"payload": payload
	}
	_socket.send_text(JSON.stringify(msg))
