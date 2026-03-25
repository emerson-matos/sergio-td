# Server (Go)

Servidor autoritativo inicial da Semana 2.

## Rodar

### Com Nix (na raiz do repo)

```bash
nix run .#server
```

ou:

```bash
nix run
```

### Sem Nix

```bash
cd server
go mod tidy
go run ./cmd/server
```

## Endpoints

- `GET /healthz` -> `200 ok`
- `GET /ws` -> WebSocket
  - envia `HELLO_ACK`
  - aceita `HELLO` e `START_MATCH`
  - transmite `SNAPSHOT_STATE` (tick + inimigos)
