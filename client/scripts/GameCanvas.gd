class_name GameCanvas
extends Node2D

const TILE_SIZE := 64.0
const MAP_WIDTH := 20
const MAP_HEIGHT := 12

var tower_nodes: Dictionary = {}
var enemy_nodes: Dictionary = {}
var path_lines: Array = []

var tower_colors: Dictionary = {
	"raiz": Color.CYAN,
	"brilhante": Color.YELLOW,
	"tank": Color.GREEN,
	"coach": Color.MAGENTA,
	"hacker": Color.GOLD
}

var enemy_colors: Dictionary = {
	"boleto": Color.RED,
	"imposto": Color.ORANGE,
	"taxa": Color.YELLOW,
	"multa": Color.PURPLE,
	"execucao": Color.DARK_RED
}

func _draw() -> void:
	draw_background()
	draw_grid()
	draw_path_lines()

func _process(_delta: float) -> void:
	queue_redraw()

func draw_background() -> void:
	var bg_color := Color(0.1, 0.15, 0.2)
	draw_rect(Rect2(0, 0, MAP_WIDTH * TILE_SIZE, MAP_HEIGHT * TILE_SIZE), bg_color)

func draw_grid() -> void:
	var grid_color := Color(0.2, 0.25, 0.3, 0.3)
	for x in range(MAP_WIDTH):
		draw_line(Vector2(x * TILE_SIZE, 0), Vector2(x * TILE_SIZE, MAP_HEIGHT * TILE_SIZE), grid_color)
	for y in range(MAP_HEIGHT):
		draw_line(Vector2(0, y * TILE_SIZE), Vector2(MAP_WIDTH * TILE_SIZE, y * TILE_SIZE), grid_color)

func draw_path_segment(from: Vector2, to: Vector2) -> void:
	var path_color := Color(0.4, 0.35, 0.3)
	var path_width := TILE_SIZE * 0.6
	draw_line(from, to, path_color, path_width)
	
	draw_circle(from, TILE_SIZE * 0.25, Color(0.5, 0.45, 0.4))
	draw_circle(to, TILE_SIZE * 0.25, Color(0.5, 0.45, 0.4))

func draw_path_lines() -> void:
	pass

func draw_build_grid() -> void:
	queue_redraw()

func clear_path() -> void:
	path_lines.clear()
	queue_redraw()

func render_towers(towers: Array) -> void:
	var current_ids: Array = []
	
	for tower_data in towers:
		var tower_id: String = tower_data.get("id", "")
		current_ids.append(tower_id)
		
		var x: float = tower_data.get("x", 0)
		var y: float = tower_data.get("y", 0)
		var tower_type: String = tower_data.get("type", "raiz")
		var level: int = tower_data.get("level", 1)
		
		var pos := Vector2(x * TILE_SIZE + TILE_SIZE/2, y * TILE_SIZE + TILE_SIZE/2)
		
		if tower_id in tower_nodes:
			var node: Node2D = tower_nodes[tower_id]
			node.position = pos
			_update_tower_visual(node, tower_type, level)
		else:
			var node := _create_tower_node(tower_type, level)
			node.position = pos
			add_child(node)
			tower_nodes[tower_id] = node
	
	var to_remove: Array = []
	for id in tower_nodes:
		if not id in current_ids:
			to_remove.append(id)
	
	for id in to_remove:
		var node: Node2D = tower_nodes[id]
		node.queue_free()
		tower_nodes.erase(id)

func _create_tower_node(tower_type: String, level: int) -> Node2D:
	var node := Node2D.new()
	
	var color: Color = tower_colors.get(tower_type, Color.WHITE)
	var radius: float = TILE_SIZE * 0.35
	
	var circle := Polygon2D.new()
	circle.polygon = _create_circle_polygon(Vector2.ZERO, radius, 16)
	circle.color = color.darkened(0.2 + level * 0.1)
	circle.position = -radius * Vector2.ONE
	node.add_child(circle)
	
	var inner := Polygon2D.new()
	inner.polygon = _create_circle_polygon(Vector2.ZERO, radius * 0.5, 12)
	inner.color = color.lightened(0.3)
	inner.position = -radius * 0.5 * Vector2.ONE
	node.add_child(inner)
	
	if level > 1:
		var level_text := Label.new()
		level_text.text = "★" + str(level - 1)
		level_text.add_theme_font_size_override("font_size", 12)
		level_text.position = Vector2(-8, -20)
		node.add_child(level_text)
	
	return node

func _update_tower_visual(node: Node2D, tower_type: String, level: int) -> void:
	while node.get_child_count() > 3:
		node.get_child(3).queue_free()
	
	if level > 1:
		var stars := Label.new()
		stars.text = "★" + str(level - 1)
		stars.add_theme_font_size_override("font_size", 12)
		stars.position = Vector2(-8, -20)
		node.add_child(stars)

func render_enemies(enemies: Array) -> void:
	var current_ids: Array = []
	
	for enemy_data in enemies:
		var enemy_id: String = enemy_data.get("id", "")
		current_ids.append(enemy_id)
		
		var x: float = enemy_data.get("x", 0)
		var y: float = enemy_data.get("y", 0)
		var enemy_type: String = enemy_data.get("type", "boleto")
		var hp: int = enemy_data.get("hp", 0)
		var max_hp: int = enemy_data.get("maxHp", 100)
		
		var pos := Vector2(x * TILE_SIZE + TILE_SIZE/2, y * TILE_SIZE + TILE_SIZE/2)
		
		if enemy_id in enemy_nodes:
			var node: Node2D = enemy_nodes[enemy_id]
			node.position = node.position.lerp(pos, 0.3)
			_update_enemy_visual(node, enemy_type, hp, max_hp)
		else:
			var node := _create_enemy_node(enemy_type, hp, max_hp)
			node.position = pos
			add_child(node)
			enemy_nodes[enemy_id] = node
	
	var to_remove: Array = []
	for id in enemy_nodes:
		if not id in current_ids:
			to_remove.append(id)
	
	for id in to_remove:
		var node: Node2D = enemy_nodes[id]
		node.queue_free()
		enemy_nodes.erase(id)

func _create_enemy_node(enemy_type: String, hp: int, max_hp: int) -> Node2D:
	var node := Node2D.new()
	
	var color: Color = enemy_colors.get(enemy_type, Color.RED)
	var radius: float = TILE_SIZE * 0.25
	
	var circle := Polygon2D.new()
	circle.polygon = _create_circle_polygon(Vector2.ZERO, radius, 12)
	circle.color = color
	circle.position = -radius * Vector2.ONE
	node.add_child(circle)
	
	_update_enemy_visual(node, enemy_type, hp, max_hp)
	
	return node

func _update_enemy_visual(node: Node2D, enemy_type: String, hp: int, max_hp: int) -> void:
	while node.get_child_count() > 1:
		node.get_child(1).queue_free()
	
	if hp < max_hp:
		var hp_bar_bg := ColorRect.new()
		hp_bar_bg.color = Color.BLACK
		hp_bar_bg.size = Vector2(TILE_SIZE * 0.5, 4)
		hp_bar_bg.position = Vector2(-TILE_SIZE * 0.25, -TILE_SIZE * 0.4)
		node.add_child(hp_bar_bg)
		
		var hp_bar := ColorRect.new()
		var hp_percent: float = float(hp) / float(max_hp)
		hp_bar.color = Color.GREEN if hp_percent > 0.5 else Color.YELLOW if hp_percent > 0.25 else Color.RED
		hp_bar.size = Vector2(TILE_SIZE * 0.5 * hp_percent, 4)
		hp_bar.position = Vector2(-TILE_SIZE * 0.25, -TILE_SIZE * 0.4)
		node.add_child(hp_bar)

func _create_circle_polygon(center: Vector2, radius: float, segments: int) -> PackedVector2Array:
	var polygon := PackedVector2Array()
	for i in range(segments):
		var angle := (float(i) / float(segments)) * TAU
		var point := center + Vector2(cos(angle), sin(angle)) * radius
		polygon.append(point)
	return polygon
