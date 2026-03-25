# Protocolo de Rede v1 (WebSocket + JSON)

## 1) Princípios

- Mensagens pequenas e versionadas.
- Campo `type` obrigatório em toda mensagem.
- Campo `ts` (timestamp cliente/servidor) para debug e latência.
- Servidor é fonte da verdade.

## 2) Envelope base

```json
{
  "v": 1,
  "type": "COMMAND_PLACE_TOWER",
  "ts": 1710000000000,
  "matchId": "m_123",
  "payload": {}
}
```

## 3) Mensagens cliente -> servidor

### `COMMAND_PLACE_TOWER`

```json
{
  "v": 1,
  "type": "COMMAND_PLACE_TOWER",
  "ts": 1710000000000,
  "matchId": "m_123",
  "payload": {
    "playerId": "p_1",
    "towerType": "dart",
    "x": 14.5,
    "y": 8.0
  }
}
```

### `COMMAND_UPGRADE_TOWER`

`payload`: `playerId`, `towerId`, `path`, `tier`.

### `COMMAND_SELL_TOWER`

`payload`: `playerId`, `towerId`.

## 4) Mensagens servidor -> cliente

### `SNAPSHOT_STATE`

```json
{
  "v": 1,
  "type": "SNAPSHOT_STATE",
  "ts": 1710000000100,
  "matchId": "m_123",
  "payload": {
    "tick": 240,
    "players": [],
    "towers": [],
    "enemies": [],
    "projectiles": []
  }
}
```

### `ACK_COMMAND`

`payload`: `commandId`, `accepted`, `reason?`.

### `EVENT_WAVE_STARTED`

`payload`: `waveNumber`, `enemySet`.

### `EVENT_MATCH_ENDED`

`payload`: `result`, `stats`.

## 5) Erros

### `ERROR_COMMAND_REJECTED`

Causas comuns:

- gold insuficiente,
- posição inválida,
- cooldown ativo,
- comando fora de contexto.

## 6) Evolução futura

- compressão (msgpack/protobuf),
- deltas binários,
- migração parcial para UDP/ENet em tráfego crítico.
