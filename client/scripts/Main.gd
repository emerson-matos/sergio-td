extends Node2D

const NetworkClientScript = preload("res://scripts/NetworkClient.gd")

@onready var status_label: Label = $StatusLabel
@onready var simulation_label: Label = $SimulationLabel

var network_client: Node

func _ready() -> void:
	network_client = NetworkClientScript.new()
	add_child(network_client)
	network_client.connected.connect(_on_connected)
	network_client.connection_failed.connect(_on_connection_failed)
	network_client.snapshot_received.connect(_on_snapshot_received)
	network_client.connect_to_server("ws://127.0.0.1:8080/ws")

func _on_connected() -> void:
	status_label.text = "Conectado ao servidor (Semana 3 comandos OK)"

func _on_connection_failed(reason: String) -> void:
	status_label.text = "Falha de conexão: %s" % reason

func _on_snapshot_received(tick: int, enemies_count: int, towers_count: int) -> void:
	simulation_label.text = "Tick: %d | Inimigos: %d | Torres: %d" % [tick, enemies_count, towers_count]
