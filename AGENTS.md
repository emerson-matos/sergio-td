# AGENTS.md - Sergio TD Development Guide

## Project Overview

**Sergio TD** is a multiplayer tower defense game (Bloons TD inspired with Brazilian humor theme).
- **Server**: Go 1.23 with Gorilla WebSocket
- **Client**: Godot 4.x (GDScript)

## Build Commands

### Server (Go)

```bash
# From project root with Nix
nix run .#server

# Or build and run manually
cd server
go mod tidy
go run ./cmd/server
go build -o server ./cmd/server   # Build binary
./server                         # Run binary

# Run single test
go test -v ./internal/ws/... -run TestName

# Run all tests
go test -v ./...
```

### Client (Godot)

```bash
# From project root with Nix
nix run .#client

# Or open in Godot editor
godot4 --path client
```

## Code Style - Go

### Formatting
- Use `gofmt` (automatic with save)
- 4-space indentation, no tabs
- One blank line between top-level declarations

### Naming
- **Variables/Functions**: `camelCase` (e.g., `playerID`, `sendMessage`)
- **Constants**: `PascalCase` or `camelCase` for scoped (e.g., `MaxPlayers`)
- **Types**: `PascalCase` (e.g., `PlayerState`, `Tower`)
- **Packages**: lowercase, short (e.g., `ws`, `sim`)

### Imports
- Group in order: standard lib, external, project
- Use aliases for conflicts (e.g., `ws "github.com/gorilla/websocket"`)

### Error Handling
- Always check errors with `if err != nil`
- Return meaningful errors: `return fmt.Errorf("failed to place tower: %w", err)`
- Use `_` to discard when appropriate

### Types & Structs
```go
type Tower struct {
    ID         string  `json:"id"`
    OwnerID    string  `json:"ownerId"`
    Type       string  `json:"type"`
    X, Y       float64 `json:"x, y"`
    Level      int     `json:"level"`
    TargetMode string  `json:"targetMode"`
}
```

### JSON
- Use snake_case in JSON tags (e.g., `playerId`, `waveNumber`)
- Float64 for coordinates (normalized 0-20, 0-11)

## Code Style - GDScript

### Formatting
- Godot auto-formats on save
- 4-space indentation
- One blank line between functions

### Naming
- **Variables/Functions**: `snake_case` (e.g., `current_towers`, `_on_button_pressed`)
- **Constants**: `SCREAMING_SNAKE_CASE` (e.g., `MAP_WIDTH`, `TILE_SIZE`)
- **Classes/Nodes**: `PascalCase` (e.g., `NetworkClient`, `Main`)

### Types
- Use type hints: `var x: float = 0.0`
- Vectors: `Vector2(x, y)`
- Arrays: `var arr: Array = []`
- Dictionaries: `var data: Dictionary = {}`

### Signals & Nodes
```gdscript
signal snapshot_received(data: Dictionary)

@onready var label: Label = $UI/StatusLabel
```

### Game Coordinates
- Server uses normalized coordinates: x in [0, 20], y in [0, 11]
- Client scales to viewport dynamically using `viewport_size.x / MAP_WIDTH`

## Architecture

### Server (`server/internal/ws/`)
- `handler.go`: WebSocket message handling, game loop (200ms tick)
- `sim.go`: Game simulation (waves, towers, enemies)
- `match.go`: Match state management
- `lobby.go`: Room/lobby management

### Client (`client/scripts/`)
- `Main.gd`: Main game scene, input handling, rendering
- `NetworkClient.gd`: WebSocket client, message handling
- `GameCanvas.gd`: (legacy, rendering in Main.gd)

## Common Patterns

### Server Message Flow
```
Client -> [JSON message] -> Server handler -> Simulation -> broadcast SNAPSHOT_STATE
```

### Client Render Loop
```
_on_snapshot_received() -> queue_redraw() -> _draw() (scales to viewport)
```

## Key Constants

- **Map**: 20x11 tiles (normalized coordinates)
- **Tick Rate**: 200ms (5 ticks/second)
- **Tower Types**: raiz, brillhante, tank, coach, hacker
- **Enemy Types**: boleto (bill), imposto (tax)

## Debug

- Press `D` in Godot to toggle debug logging
- Server broadcasts `SNAPSHOT_STATE` to all clients every tick
- Use `curl http://localhost:8080/health` to check server