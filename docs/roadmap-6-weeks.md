# Roadmap (6 semanas) — MVP Multiplayer TD

## Semana 1 — Fundação

- Definir arquitetura v1 e protocolo v1.
- Criar projetos base:
  - `client/` Godot,
  - `server/` Go.
- Conectar WebSocket com handshake simples.

## Semana 2 — Loop mínimo jogável

- Mapa único e caminho fixo.
- Spawn de inimigos no servidor.
- Render + interpolação no cliente.
- Torre básica com targeting `first`.

## Semana 3 — Comandos de gameplay

- Colocar torre (`COMMAND_PLACE_TOWER`).
- Upgrade e venda.
- Validações server-side (gold/rules).
- ACK de comandos + tratamento de erro.

## Semana 4 — Multiplayer real

- Lobby mínimo (2 a 4 jogadores).
- Start de partida sincronizado.
- Estado de cada jogador (vida, ouro, score).
- Reconexão simples em sessão ativa.

## Semana 5 — Qualidade e estabilidade

- Balance inicial de ondas/torres.
- Métricas básicas (latência, tick drift, rejeições).
- Logs estruturados e IDs de correlação.
- Testes de carga leves (salas simultâneas).

## Semana 6 — Polimento de MVP

- UI mínima de partida e pós-partida.
- Efeitos visuais essenciais.
- Build desktop + web.
- Checklist de release do MVP interno.

## Critérios de pronto (MVP)

- Partida multiplayer estável de 10–15 minutos.
- Sem dessync crítico perceptível.
- Comandos principais (colocar, upgrade, vender) funcionais.
- 1 mapa + 1 modo de jogo end-to-end.
