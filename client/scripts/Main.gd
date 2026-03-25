extends Node2D

# Map is 20x11 tiles - coordinates normalized 0-1
const MAP_WIDTH := 20.0
const MAP_HEIGHT := 11.0

var network_client: Node
var game_canvas: Node2D

@onready var status_label: Label = $UI/StatusLabel
@onready var gold_label: Label = $UI/HUD/VBoxContainer/GoldLabel
@onready var lives_label: Label = $UI/HUD/VBoxContainer/LivesLabel
@onready var wave_label: Label = $UI/HUD/VBoxContainer/WaveLabel
@onready var score_label: Label = $UI/HUD/VBoxContainer/ScoreLabel
@onready var start_wave_button: Button = $UI/HUD/StartWaveButton
@onready var tower_buttons_container: HBoxContainer = $UI/TowerPanel/VBoxContainer/ButtonsContainer

var waypoints: Array = []
var current_towers: Array = []
var current_enemies: Array = []
var player_gold: int = 0
var player_lives: int = 20
var player_score: int = 0
var current_wave: int = 0
var selected_tower_type: String = "raiz"
var game_started: bool = false
var hover_tile: Vector2 = Vector2(-1, -1)
var hover_valid: bool = false
var waiting_for_ready := true

const TOWER_COLORS = {
	"raiz": Color.CYAN,
	"brilhante": Color.YELLOW,
	"tank": Color.GREEN,
	"coach": Color.MAGENTA,
	"hacker": Color.GOLD
}

const ENEMY_COLORS = {
	"boleto": Color.RED,
	"imposto": Color.ORANGE,
	"taxa": Color.YELLOW,
	"multa": Color.PURPLE,
	"execucao": Color.DARK_RED
}

var tower_nodes: Dictionary = {}
var enemy_nodes: Dictionary = {}

func _ready() -> void:
	var NetworkClientScript = load("res://scripts/NetworkClient.gd")
	network_client = NetworkClientScript.new()
	add_child(network_client)
	
	network_client.connected.connect(_on_connected)
	network_client.connection_failed.connect(_on_connection_failed)
	network_client.snapshot_received.connect(_on_snapshot_received)
	network_client.game_data_received.connect(_on_game_data_received)
	network_client.wave_started.connect(_on_wave_started)
	network_client.wave_completed.connect(_on_wave_completed)
	network_client.match_ended.connect(_on_match_ended)
	network_client.command_accepted.connect(_on_command_accepted)
	network_client.command_rejected.connect(_on_command_rejected)
	
	network_client.connect_to_server("ws://127.0.0.1:8080/ws")
	
	game_canvas = $GameCanvas
	
	setup_tower_buttons()
	start_wave_button.pressed.connect(_on_start_wave_pressed)

func setup_tower_buttons() -> void:
	var tower_types = ["raiz", "brilhante", "tank", "coach", "hacker"]
	var tower_names = {
		"raiz": "Careca Raiz\n$100",
		"brilhante": "Careca Brilhante\n$250",
		"tank": "Careca Tank\n$150",
		"coach": "Careca Coach\n$300",
		"hacker": "Careca Hacker\n$200"
	}
	
	for tower_type in tower_types:
		var btn = Button.new()
		btn.text = tower_names.get(tower_type, tower_type)
		btn.pressed.connect(_on_tower_button_pressed.bind(tower_type))
		tower_buttons_container.add_child(btn)

func _on_connected() -> void:
	status_label.text = "Conectado! Esperando dados do jogo..."

func _on_connection_failed(reason: String) -> void:
	status_label.text = "Erro: %s" % reason

func _on_game_data_received(data: Dictionary) -> void:
	waypoints = data.get("waypoints", [])
	var tower_types = data.get("towerTypes", {})
	
	queue_redraw()
	
	status_label.text = "Clique em PRONTO para iniciar!"
	start_wave_button.text = "PRONTO"
	start_wave_button.disabled = false

func _on_snapshot_received(data: Dictionary) -> void:
	current_towers = data.get("towers", [])
	current_enemies = data.get("enemies", [])
	current_wave = data.get("waveNumber", 0)
	
	var players = data.get("players", [])
	if players.size() > 0:
		var my_player = players[0]
		player_gold = my_player.get("gold", 0)
		player_lives = my_player.get("lives", 0)
		player_score = my_player.get("score", 0)
	
	_update_hud()
	_render_game()

func _on_wave_started(wave_number: int) -> void:
	status_label.text = "Wave %d iniciada!" % wave_number
	current_wave = wave_number
	waiting_for_ready = false
	start_wave_button.text = "PRÓXIMA WAVE"
	start_wave_button.disabled = false

func _on_wave_completed(wave_number: int) -> void:
	status_label.text = "Wave %d completada! Prepare para a próxima..." % wave_number

func _on_match_ended(victory: bool, wave_number: int) -> void:
	if victory:
		status_label.text = "VITÓRIA! Você completou todas as waves!"
	else:
		status_label.text = "DERROTA! Você sobreviveu até a wave %d" % wave_number
	start_wave_button.disabled = true

func _on_command_accepted(command_id: String, data: Dictionary) -> void:
	pass

func _on_command_rejected(command_id: String, reason: String) -> void:
	status_label.text = "Erro: %s" % reason

func _on_start_wave_pressed() -> void:
	if waiting_for_ready:
		network_client.send_player_ready()
		status_label.text = "Aguardando outros jogadores..."
	else:
		network_client.send_start_wave()
		status_label.text = "Iniciando próxima wave..."

func _on_tower_button_pressed(tower_type: String) -> void:
	selected_tower_type = tower_type
	status_label.text = "Torre selecionada: %s - Clique no mapa para posicionar" % tower_type

func _input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event = event as InputEventKey
		if key_event.keycode == KEY_D and key_event.pressed:
			network_client.debug_enabled = not network_client.debug_enabled
			status_label.text = "Debug: " + ("ON" if network_client.debug_enabled else "OFF")
	
	if event is InputEventMouseButton:
		var mouse_event = event as InputEventMouseButton
		if mouse_event.button_index == MOUSE_BUTTON_LEFT and mouse_event.pressed:
			var mouse_pos = get_viewport().get_mouse_position()
			
			# Convert viewport position to normalized game coordinates (0-1)
			var viewport_size = get_viewport_rect().size
			var norm_x = mouse_pos.x / viewport_size.x
			var norm_y = mouse_pos.y / viewport_size.y
			
			# Map normalized coordinates to game world
			var game_x = norm_x * MAP_WIDTH
			var game_y = norm_y * MAP_HEIGHT
			
			if game_x >= 0 and game_x < MAP_WIDTH and game_y >= 0 and game_y < MAP_HEIGHT:
				network_client.send_place_tower(selected_tower_type, game_x, game_y)

func _process(_delta: float) -> void:
	var mouse_pos = get_viewport().get_mouse_position()
	var viewport_size = get_viewport_rect().size
	
	# Convert to normalized coordinates (0-1)
	var norm_x = mouse_pos.x / viewport_size.x
	var norm_y = mouse_pos.y / viewport_size.y
	
	# Map to game world
	var game_x = norm_x * MAP_WIDTH
	var game_y = norm_y * MAP_HEIGHT
	
	if game_x >= 0 and game_x < MAP_WIDTH and game_y >= 0 and game_y < MAP_HEIGHT:
		hover_tile = Vector2(int(game_x), int(game_y))
		hover_valid = true
	else:
		hover_tile = Vector2(-1, -1)
		hover_valid = false
	
	queue_redraw()

func _update_hud() -> void:
	gold_label.text = "Ouro: %d" % player_gold
	lives_label.text = "Vidas: %d" % player_lives
	wave_label.text = "Wave: %d" % current_wave
	score_label.text = "Score: %d" % player_score

func _render_game() -> void:
	queue_redraw()
	render_towers()
	render_enemies()

func _draw() -> void:
	var viewport_size = get_viewport_rect().size
	var scale_x = viewport_size.x / MAP_WIDTH
	var scale_y = viewport_size.y / MAP_HEIGHT
	
	draw_background(scale_x, scale_y)
	draw_grid(scale_x, scale_y)
	draw_path(scale_x, scale_y)
	draw_hover_preview(scale_x, scale_y)

func draw_hover_preview(scale_x: float, scale_y: float) -> void:
	if not hover_valid:
		return
	
	var pos = Vector2(hover_tile.x * scale_x, hover_tile.y * scale_y)
	var preview_color = TOWER_COLORS.get(selected_tower_type, Color.WHITE)
	preview_color.a = 0.5
	
	draw_rect(Rect2(pos.x, pos.y, scale_x, scale_y), preview_color)
	
	var tower_costs = {"raiz": 100, "brilhante": 250, "tank": 150, "coach": 300, "hacker": 200}
	var cost = tower_costs.get(selected_tower_type, 100)
	var cost_color = Color.GREEN if player_gold >= cost else Color.RED
	
	draw_string(ThemeDB.fallback_font, pos + Vector2(5, scale_y - 8), "$%d" % cost, HORIZONTAL_ALIGNMENT_LEFT, -1, 12, cost_color)

func draw_background(scale_x: float, scale_y: float) -> void:
	var bg_color = Color(0.1, 0.15, 0.2)
	draw_rect(Rect2(0, 0, MAP_WIDTH * scale_x, MAP_HEIGHT * scale_y), bg_color)

func draw_grid(scale_x: float, scale_y: float) -> void:
	var grid_color = Color(0.2, 0.25, 0.3, 0.3)
	for x in range(int(MAP_WIDTH)):
		draw_line(Vector2(x * scale_x, 0), Vector2(x * scale_x, MAP_HEIGHT * scale_y), grid_color)
	for y in range(int(MAP_HEIGHT)):
		draw_line(Vector2(0, y * scale_y), Vector2(MAP_WIDTH * scale_x, y * scale_y), grid_color)

func draw_path(scale_x: float, scale_y: float) -> void:
	if waypoints.size() < 2:
		return
		
	var prev_pos = Vector2(waypoints[0].x * scale_x, waypoints[0].y * scale_y)
	for i in range(1, waypoints.size()):
		var wp = waypoints[i]
		var wp_pos = Vector2(wp.x * scale_x, wp.y * scale_y)
		var path_color = Color(0.4, 0.35, 0.3)
		var path_width = min(scale_x, scale_y) * 0.6
		draw_line(prev_pos, wp_pos, path_color, path_width)
		draw_circle(prev_pos, min(scale_x, scale_y) * 0.2, Color(0.5, 0.45, 0.4))
		prev_pos = wp_pos
	draw_circle(prev_pos, min(scale_x, scale_y) * 0.2, Color(0.5, 0.45, 0.4))

func render_towers() -> void:
	var viewport_size = get_viewport_rect().size
	var scale_x = viewport_size.x / MAP_WIDTH
	var scale_y = viewport_size.y / MAP_HEIGHT
	var current_ids = []
	
	for tower_data in current_towers:
		var tower_id = tower_data.get("id", "")
		current_ids.append(tower_id)
		
		var x: float = tower_data.get("x", 0)
		var y: float = tower_data.get("y", 0)
		var tower_type = tower_data.get("type", "raiz")
		var level: int = tower_data.get("level", 1)
		
		var pos = Vector2(x * scale_x + scale_x/2, y * scale_y + scale_y/2)
		
		if tower_id in tower_nodes:
			var node = tower_nodes[tower_id]
			node.position = node.position.lerp(pos, 0.3)
		else:
			var node = _create_tower_node(tower_type, level, scale_x, scale_y)
			node.position = pos
			add_child(node)
			tower_nodes[tower_id] = node
	
	var to_remove = []
	for id in tower_nodes:
		if not id in current_ids:
			to_remove.append(id)
	
	for id in to_remove:
		var node = tower_nodes[id]
		node.queue_free()
		tower_nodes.erase(id)

func _create_tower_node(tower_type: String, level: int, scale_x: float, scale_y: float) -> Node2D:
	var node = Node2D.new()
	
	var color = TOWER_COLORS.get(tower_type, Color.WHITE)
	var radius = min(scale_x, scale_y) * 0.35
	
	var circle = Polygon2D.new()
	circle.polygon = _create_circle_polygon(Vector2.ZERO, radius, 16)
	circle.color = color.darkened(0.2 + level * 0.1)
	circle.position = -radius * Vector2.ONE
	node.add_child(circle)
	
	var inner = Polygon2D.new()
	inner.polygon = _create_circle_polygon(Vector2.ZERO, radius * 0.5, 12)
	inner.color = color.lightened(0.3)
	inner.position = -radius * 0.5 * Vector2.ONE
	node.add_child(inner)
	
	if level > 1:
		var level_text = Label.new()
		level_text.text = "★" + str(level - 1)
		level_text.add_theme_font_size_override("font_size", 12)
		level_text.position = Vector2(-8, -20)
		node.add_child(level_text)
	
	return node

func render_enemies() -> void:
	var viewport_size = get_viewport_rect().size
	var scale_x = viewport_size.x / MAP_WIDTH
	var scale_y = viewport_size.y / MAP_HEIGHT
	var current_ids = []
	
	for enemy_data in current_enemies:
		var enemy_id = enemy_data.get("id", "")
		current_ids.append(enemy_id)
		
		var x: float = enemy_data.get("x", 0)
		var y: float = enemy_data.get("y", 0)
		var enemy_type = enemy_data.get("type", "boleto")
		var hp: int = enemy_data.get("hp", 0)
		var max_hp: int = enemy_data.get("maxHp", 100)
		
		var pos = Vector2(x * scale_x + scale_x/2, y * scale_y + scale_y/2)
		
		if enemy_id in enemy_nodes:
			var node = enemy_nodes[enemy_id]
			node.position = node.position.lerp(pos, 0.3)
			_update_enemy_hp(node, hp, max_hp, scale_x, scale_y)
		else:
			var node = _create_enemy_node(enemy_type, hp, max_hp, scale_x, scale_y)
			node.position = pos
			add_child(node)
			enemy_nodes[enemy_id] = node
	
	var to_remove = []
	for id in enemy_nodes:
		if not id in current_ids:
			to_remove.append(id)
	
	for id in to_remove:
		var node = enemy_nodes[id]
		node.queue_free()
		enemy_nodes.erase(id)

func _create_enemy_node(enemy_type: String, hp: int, max_hp: int, scale_x: float, scale_y: float) -> Node2D:
	var node = Node2D.new()
	
	var color = ENEMY_COLORS.get(enemy_type, Color.RED)
	var radius = min(scale_x, scale_y) * 0.25
	
	var circle = Polygon2D.new()
	circle.polygon = _create_circle_polygon(Vector2.ZERO, radius, 12)
	circle.color = color
	circle.position = -radius * Vector2.ONE
	node.add_child(circle)
	
	_update_enemy_hp(node, hp, max_hp, scale_x, scale_y)
	
	return node

func _update_enemy_hp(node: Node2D, hp: int, max_hp: int, scale_x: float, scale_y: float) -> void:
	while node.get_child_count() > 1:
		node.get_child(1).queue_free()
	
	if hp < max_hp:
		var hp_bar_bg = ColorRect.new()
		hp_bar_bg.color = Color.BLACK
		hp_bar_bg.size = Vector2(min(scale_x, scale_y) * 0.5, 4)
		hp_bar_bg.position = Vector2(-min(scale_x, scale_y) * 0.25, -min(scale_x, scale_y) * 0.4)
		node.add_child(hp_bar_bg)
		
		var hp_bar = ColorRect.new()
		var hp_percent = float(hp) / float(max_hp)
		hp_bar.color = Color.GREEN if hp_percent > 0.5 else Color.YELLOW if hp_percent > 0.25 else Color.RED
		hp_bar.size = Vector2(min(scale_x, scale_y) * 0.5 * hp_percent, 4)
		hp_bar.position = Vector2(-min(scale_x, scale_y) * 0.25, -min(scale_x, scale_y) * 0.4)
		node.add_child(hp_bar)

func _create_circle_polygon(center: Vector2, radius: float, segments: int) -> PackedVector2Array:
	var polygon = PackedVector2Array()
	for i in range(segments):
		var angle = (float(i) / float(segments)) * TAU
		var point = center + Vector2(cos(angle), sin(angle)) * radius
		polygon.append(point)
	return polygon
