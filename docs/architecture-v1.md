# Arquitetura v1

## 1) Visão geral

Arquitetura cliente-servidor com servidor autoritativo:

- **Client (Godot)**
  - renderização 2D,
  - input (mouse/touch),
  - UI/UX,
  - interpolação visual.
- **Server (Go)**
  - estado oficial da partida,
  - simulação das ondas,
  - validação de comandos,
  - broadcast de snapshots.
- **Infra**
  - WebSocket para realtime,
  - banco relacional para conta/progressão (futuro),
  - Redis opcional para sessões/lobbies (futuro).

Decisão de stack:

- Backend 100% em **Go** para simplificar operação, contratação e manutenção.
- Evitar mistura de stacks nesta fase do projeto.

## 2) Responsabilidades

### Cliente (Godot)

- Capturar intenção do jogador:
  - construir torre,
  - vender torre,
  - upgrade,
  - habilidades ativas.
- Exibir estado recebido do servidor com suavização (lerp/interpolação).
- Aplicar feedback local não-autoritativo (efeitos, animações, SFX).

### Servidor (Go)

- Receber e validar comandos (custos, cooldown, posições válidas).
- Avançar simulação em ticks fixos.
- Resolver targeting, dano, mortes, renda e progressão de ondas.
- Enviar atualizações de estado em intervalos constantes.

## 3) Loop de simulação

- Tick lógico sugerido: **20 TPS** (50 ms).
- Broadcast de snapshot sugerido: **10–20 Hz**.
- Cliente interpola com buffer curto (ex.: 100–150 ms) para esconder jitter.

## 4) Modelo de dados (alto nível)

- `Match`
  - `matchId`, `seed`, `status`, `tick`
- `PlayerState`
  - `playerId`, `gold`, `life`, `loadout`
- `Tower`
  - `towerId`, `ownerId`, `type`, `position`, `level`, `targetMode`
- `Enemy`
  - `enemyId`, `type`, `hp`, `pathProgress`, `statusEffects`

## 5) Segurança e anti-cheat (v1)

- Cliente **nunca** define dano/vida/inimigos.
- Cliente envia só intenção de ação.
- Servidor valida:
  - recursos,
  - range,
  - timing/cooldown,
  - regras da partida.

## 6) Escalabilidade (evolução)

- v1: 1 instância servidor para desenvolvimento.
- v2: matchmaking + alocação de salas por processo.
- v3: gateway + múltiplos workers de partida + observabilidade completa.

## 7) Pastas recomendadas

```text
/client   # Godot
/server   # Go WebSocket authoritative
/docs     # arquitetura, protocolo, roadmap
```
