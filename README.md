# Sergio TD — Multiplayer Tower Defense (Godot + Go Authoritative Server)

Este repositório inicia o planejamento técnico para um jogo estilo Bloons TD com foco em:

- **Multiplataforma real**: Desktop (Linux/Windows/macOS), Mobile (Android/iOS) e Web.
- **Cliente único** em **Godot**.
- **Servidor autoritativo** em **Go** (stack única no backend).

## Objetivo

Construir um MVP multiplayer cooperativo/competitivo com:

- partida em tempo real,
- colocação e upgrade de torres,
- ondas de inimigos sincronizadas pelo servidor,
- protocolo de rede simples e evolutivo.

## Estrutura de documentação

- `docs/architecture-v1.md`: arquitetura alvo (cliente, servidor, comunicação, deploy).
- `docs/network-protocol-v1.md`: contrato inicial de mensagens cliente-servidor.
- `docs/roadmap-6-weeks.md`: plano de execução por semana para chegar em MVP.

## Decisões técnicas iniciais

1. **Cliente Godot**: render, input, UI, interpolação e predição leve.
2. **Servidor autoritativo**: estado oficial do jogo, validação de ações e anti-cheat básico.
3. **Comunicação WebSocket** no início (evolução futura para UDP/ENet quando necessário).
4. **Snapshots + deltas** para sincronização de estado.

## Status atual

Semana 3 iniciada com primeiro comando de gameplay autoritativo:

- `client/` conecta por WebSocket, envia `HELLO` + `START_MATCH` e dispara `COMMAND_PLACE_TOWER`.
- `server/` em Go valida `COMMAND_PLACE_TOWER` (gold, tipo, posição), responde `ACK_COMMAND`/`ERROR_COMMAND_REJECTED` e inclui `towers`/`players` no `SNAPSHOT_STATE`.

## Rodando o projeto

### Opção 1: com Nix (recomendado)

Pré-requisito: Nix com `flakes` habilitado.

```bash
nix develop
```

Ou sem entrar no shell, direto com `nix run`:

```bash
nix run .#server
```

```bash
nix run .#client
```

> `nix run` (sem sufixo) usa o app default e sobe o servidor (`.#server`).

### Opção 2: sem Nix (manual)

Pré-requisitos:

- Go 1.24+
- Godot 4.x

Servidor:

```bash
cd server
go mod tidy
go run ./cmd/server
```

Cliente:

```bash
cd client
godot4 --path .
```

> No editor do Godot, também é possível abrir `client/project.godot` e executar a cena principal.

## Verificação rápida

Com o servidor rodando:

```bash
curl -i http://127.0.0.1:8080/healthz
```

Resposta esperada: `HTTP/1.1 200 OK` e body `ok`.

## Próximos passos

1. Completar Semana 3:
   - `COMMAND_UPGRADE_TOWER`,
   - `COMMAND_SELL_TOWER`,
   - custo/economia por jogador com regras de upgrade/venda.
