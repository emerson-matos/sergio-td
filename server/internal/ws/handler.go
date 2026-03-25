package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type envelope struct {
	Version       int                    `json:"v"`
	Type          string                 `json:"type"`
	TS            float64                `json:"ts"`
	MatchID       string                 `json:"matchId,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
}

var sharedMatch = newMatchState("m_1")
var lobbyServer = NewLobbyServer()
var currentRoomID = ""
var currentPlayerID = ""

func Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("upgrade failed", "error", err)
		return
	}
	defer func() { _ = conn.Close() }()
	defer func() {
		if currentRoomID != "" && currentPlayerID != "" {
			lobbyServer.LeaveRoom(currentRoomID, currentPlayerID)
		}
		playerID, connected := sharedMatch.unregisterConnection(conn)
		if playerID != "" {
			broadcastLobbyState(connected)
		}
	}()

	// Send initial lobby state
	rooms := lobbyServer.GetRooms()
	_ = writeMessage(conn, "LOBBY_LIST", map[string]any{
		"rooms": rooms,
	})

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	rawMessages := make(chan []byte)
	readErr := make(chan error, 1)
	go readLoop(conn, rawMessages, readErr)

	for {
		select {
		case raw := <-rawMessages:
			raw = bytes.TrimSpace(raw)
			if len(raw) == 0 {
				continue
			}

			var msg envelope
			if err := json.Unmarshal(raw, &msg); err != nil {
				_ = writeMessage(conn, "ERROR_BAD_MESSAGE", map[string]any{
					"reason": "invalid json",
				})
				continue
			}

			switch msg.Type {
			case "HELLO":
				playerID := asString(msg.Payload["playerId"])
				if playerID == "" {
					playerID = fmt.Sprintf("p_%d", time.Now().UnixMilli()%10000)
				}
				reconnected, connected := sharedMatch.registerConnection(playerID, conn)
				_ = writeMessage(conn, "ACK_HELLO", map[string]any{
					"accepted":       true,
					"playerId":       playerID,
					"matchId":        sharedMatch.matchID,
					"isReconnected":  reconnected,
					"connectedCount": connected,
					"waypoints":      GetWaypoints(),
					"towerTypes":     GetTowerTypes(),
					"enemyTypes":     GetEnemyTypes(),
				})
				slog.Info("player connected", "playerId", playerID, "reconnected", reconnected, "connected", connected)
				broadcastLobbyState(connected)

			case "START_WAVE":
				var waveNum int
				var started bool
				sharedMatch.withSimulation(func(sim *Simulation) {
					started = sim.StartWave()
					if started {
						waveNum = sim.WaveNumber()
					} else {
						slog.Warn("cannot start wave", "waveNumber", sim.WaveNumber(), "totalWaves", len(sim.Waves()))
					}
				})
				if started {
					broadcastAll("EVENT_WAVE_STARTED", map[string]any{"waveNumber": waveNum})
				}

			case "PLAYER_READY":
				playerID := sharedMatch.connectionPlayer(conn)
				ready := sharedMatch.setPlayerReady(playerID, true)
				broadcastLobbyState(sharedMatch.getConnectedCount())

				// Auto-start when all players ready
				if ready && sharedMatch.allPlayersReady() && !sharedMatch.isSimulationRunning() {
					sharedMatch.withSimulation(func(sim *Simulation) {
						sim.StartWave()
						sharedMatch.setSimulationRunningLocked(true)
					})
					broadcastAll("EVENT_MATCH_STARTED", map[string]any{
						"matchId": sharedMatch.matchID,
					})
					broadcastAll("EVENT_WAVE_STARTED", map[string]any{
						"waveNumber": 1,
					})
				}

			case "COMMAND_PLACE_TOWER":
				playerID := sharedMatch.connectionPlayer(conn)
				towerType := asString(msg.Payload["towerType"])
				x, xOK := asFloat(msg.Payload["x"])
				y, yOK := asFloat(msg.Payload["y"])
				commandID := asString(msg.Payload["commandId"])
				if commandID == "" {
					commandID = fmt.Sprintf("cmd_%d", time.Now().UnixMilli())
				}

				if !xOK || !yOK {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    "invalid coordinates",
					})
					continue
				}

				var tower Tower
				var placeErr error
				sharedMatch.withSimulation(func(sim *Simulation) {
					tower, placeErr = sim.PlaceTower(playerID, towerType, x, y)
				})
				if placeErr != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    placeErr.Error(),
					})
					continue
				}

				_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
					"commandId": commandID,
					"accepted":  true,
					"towerId":   tower.ID,
				})

			case "COMMAND_UPGRADE_TOWER":
				playerID := sharedMatch.connectionPlayer(conn)
				towerID := asString(msg.Payload["towerId"])
				commandID := asString(msg.Payload["commandId"])
				if commandID == "" {
					commandID = fmt.Sprintf("cmd_%d", time.Now().UnixMilli())
				}

				var tower Tower
				var upgradeErr error
				sharedMatch.withSimulation(func(sim *Simulation) {
					tower, upgradeErr = sim.UpgradeTower(playerID, towerID)
				})
				if upgradeErr != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    upgradeErr.Error(),
					})
					continue
				}

				_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
					"commandId": commandID,
					"accepted":  true,
					"towerId":   tower.ID,
					"level":     tower.Level,
				})

			case "COMMAND_SET_TARGET":
				playerID := sharedMatch.connectionPlayer(conn)
				towerID := asString(msg.Payload["towerId"])
				targetMode := asString(msg.Payload["targetMode"])
				commandID := asString(msg.Payload["commandId"])
				if commandID == "" {
					commandID = fmt.Sprintf("cmd_%d", time.Now().UnixMilli())
				}

				var setErr error
				sharedMatch.withSimulation(func(sim *Simulation) {
					setErr = sim.SetTargetMode(playerID, towerID, targetMode)
				})
				if setErr != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    setErr.Error(),
					})
					continue
				}

				_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
					"commandId":  commandID,
					"accepted":   true,
					"towerId":    towerID,
					"targetMode": targetMode,
				})

			case "COMMAND_SELL_TOWER":
				playerID := sharedMatch.connectionPlayer(conn)
				towerID := asString(msg.Payload["towerId"])
				commandID := asString(msg.Payload["commandId"])
				if commandID == "" {
					commandID = fmt.Sprintf("cmd_%d", time.Now().UnixMilli())
				}

				var refund int
				var sellErr error
				sharedMatch.withSimulation(func(sim *Simulation) {
					refund, sellErr = sim.SellTower(playerID, towerID)
				})
				if sellErr != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    sellErr.Error(),
					})
					continue
				}

				_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
					"commandId": commandID,
					"accepted":  true,
					"towerId":   towerID,
					"refund":    refund,
				})

			case "GET_GAME_DATA":
				sharedMatch.withSimulation(func(sim *Simulation) {
					_ = writeMessage(conn, "GAME_DATA", map[string]any{
						"waypoints":  GetWaypoints(),
						"towerTypes": GetTowerTypes(),
						"enemyTypes": GetEnemyTypes(),
						"waveNumber": sim.WaveNumber(),
					})
				})

			default:
				_ = writeMessage(conn, "ECHO", map[string]any{"raw": string(raw)})
			}

		case err := <-readErr:
			slog.Error("read failed", "error", err)
			return

		case <-ticker.C:
			tickStart := time.Now()

			var snapPayload map[string]any
			var matchEndPayload map[string]any

			sharedMatch.withSimulation(func(sim *Simulation) {
				if sharedMatch.isSimulationRunningLocked() {
					sim.Step()
				}

				snapPayload = map[string]any{
					"tick":       sim.Tick(),
					"waveNumber": sim.WaveNumber(),
					"players":    sim.Players(),
					"towers":     sim.Towers(),
					"enemies":    sim.Enemies(),
				}

				if sim.Victory() {
					matchEndPayload = map[string]any{
						"result":     "victory",
						"waveNumber": sim.WaveNumber(),
						"victory":    true,
						"stats":      sim.MatchStats(),
					}
					sharedMatch.setSimulationRunningLocked(false)
				} else if sim.GameOver() {
					matchEndPayload = map[string]any{
						"result":     "defeat",
						"waveNumber": sim.WaveNumber(),
						"victory":    false,
						"stats":      sim.MatchStats(),
					}
					sharedMatch.setSimulationRunningLocked(false)
				}
			})

			broadcastAll("SNAPSHOT_STATE", snapPayload)
			if matchEndPayload != nil {
				broadcastAll("EVENT_MATCH_ENDED", matchEndPayload)
			}
			tickDuration := time.Since(tickStart)
			RecordTick(tickDuration)
			if tickDuration > 50*time.Millisecond {
				slog.Warn("tick drift", "duration", tickDuration)
			}
		}
	}
}

func broadcastLobbyState(connectedPlayers int) {
	broadcastAll("LOBBY_STATE", map[string]any{
		"matchId":        sharedMatch.matchID,
		"connectedCount": connectedPlayers,
		"minToStart":     1,
		"maxPlayers":     4,
	})
}

func broadcastAll(msgType string, payload map[string]any) {
	sharedMatch.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(sharedMatch.conns))
	for conn := range sharedMatch.conns {
		conns = append(conns, conn)
	}
	sharedMatch.mu.Unlock()

	for _, conn := range conns {
		if err := writeMessage(conn, msgType, payload); err != nil {
			slog.Error("broadcast failed", "msgType", msgType, "error", err)
		}
	}
}

func asString(value any) string {
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func asFloat(value any) (float64, bool) {
	number, ok := value.(float64)
	return number, ok
}

func readLoop(conn *websocket.Conn, messages chan<- []byte, readErr chan<- error) {
	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			readErr <- err
			return
		}
		if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
			continue
		}
		messages <- payload
	}
}

func writeMessage(conn *websocket.Conn, msgType string, payload map[string]any) error {
	return conn.WriteJSON(map[string]any{
		"v":       1,
		"type":    msgType,
		"ts":      time.Now().UnixMilli(),
		"payload": payload,
	})
}
