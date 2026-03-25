# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Sergio TD — multiplayer tower defense (Bloons TD inspired, Brazilian humor theme). Monorepo with Go authoritative server and Godot 4.x GDScript client.

## Build & Run

```bash
# Preferred: Nix dev environment (includes Go, Godot, golangci-lint, gopls, websocat, jq)
nix develop
nix run .#dev      # starts both server and client
nix run .#server   # server only
nix run .#client   # client only

# Manual
cd server && go run ./cmd/server    # server on :8080
godot4 --path client                # client
```

## Test

```bash
cd server && go test -v ./...                        # all tests
cd server && go test -v ./internal/ws/... -run TestName  # single test
```

## Lint

```bash
cd server && golangci-lint run ./...
```

## Key Architecture

- Server tick rate: 200ms (5 TPS). All game state is server-authoritative.
- WebSocket JSON protocol on `/ws`, health check on `/healthz`.
- Map: 20×11 normalized tile coordinates. Client scales to viewport dynamically.
- Currently single shared match (`sharedMatch`), no matchmaking yet.
- For protocol details: @docs/network-protocol-v1.md
- For architecture: @docs/architecture-v1.md
- For roadmap: @docs/roadmap-6-weeks.md

## Code Style

Follows conventions in @AGENTS.md — Go uses `gofmt`, GDScript uses Godot auto-format. Key points:
- Go JSON tags use camelCase (e.g., `playerId`, `waveNumber`)
- GDScript requires type hints on all variables
- Go errors must always be checked with `if err != nil`

## Gotchas

- Client hard-codes `ws://127.0.0.1:8080/ws` — local dev only.
- `server/server` binary should not be committed (add to .gitignore).
- Player ID is client-generated from `OS.get_unique_id()`, not server-assigned.
