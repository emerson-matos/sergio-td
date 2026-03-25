# Client (Godot)

Cliente inicial da Semana 2 com visualização do loop de simulação.

## Objetivo desta etapa

- Abrir cena principal.
- Tentar conectar em `ws://127.0.0.1:8080/ws`.
- Enviar `HELLO` e `START_MATCH` ao conectar.
- Mostrar na UI o tick e a quantidade de inimigos recebidos via `SNAPSHOT_STATE`.

## Como testar

1. Suba o servidor Go (`server/`).
2. Abra o projeto `client/` no Godot 4.
3. Rode a cena principal.
4. Verifique os labels:
   - status de conexão,
   - `Tick: ... | Inimigos ativos: ...` atualizando.

## Rodar por terminal

### Com Nix (na raiz do repo)

```bash
nix run .#client
```

### Sem Nix

```bash
cd client
godot4 --path .
```
