package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
	Version int                    `json:"v"`
	Type    string                 `json:"type"`
	TS      int64                  `json:"ts"`
	MatchID string                 `json:"matchId,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

var sharedMatch = newMatchState("m_1")

func Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	defer func() {
		playerID, connected := sharedMatch.unregisterConnection(conn)
		if playerID != "" {
			broadcastLobbyState(connected)
		}
	}()

	if err := writeMessage(conn, "HELLO_ACK", map[string]any{
		"message": "week7-multiplayer-ready",
		"matchId": sharedMatch.matchID,
	}); err != nil {
		log.Printf("write ack failed: %v", err)
		return
	}

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
					_ = writeMessage(conn, "ACK_HELLO", map[string]any{
						"accepted": false,
						"reason":   "missing playerId",
					})
					continue
				}
				reconnected, connected := sharedMatch.registerConnection(playerID, conn)
				_ = writeMessage(conn, "ACK_HELLO", map[string]any{
					"accepted":       true,
					"playerId":       playerID,
					"matchId":        sharedMatch.matchID,
					"isReconnected":  reconnected,
					"connectedCount": connected,
				})
				broadcastLobbyState(connected)
			case "START_MATCH":
				if !sharedMatch.tryStartMatch(2) {
					_ = writeMessage(conn, "ERROR_COMMAND_REJECTED", map[string]any{
						"reason": "need at least 2 connected players to start",
					})
					continue
				}
				broadcastAll("EVENT_MATCH_STARTED", map[string]any{
					"matchId": sharedMatch.matchID,
				})
				broadcastAll("EVENT_WAVE_STARTED", map[string]any{
					"waveNumber": 1,
				})
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
					_ = writeMessage(conn, "ERROR_COMMAND_REJECTED", map[string]any{
						"commandId": commandID,
						"reason":    "invalid coordinates",
					})
					continue
				}

				var tower Tower
				var placeErr error
				sharedMatch.withSimulation(func(sim *Simulation) {
					tower, placeErr = sim.PlaceTower(playerID, towerType, x, y)
				})
				err := placeErr
				if err != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    err.Error(),
					})
					_ = writeMessage(conn, "ERROR_COMMAND_REJECTED", map[string]any{
						"commandId": commandID,
						"reason":    err.Error(),
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
				err := upgradeErr
				if err != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    err.Error(),
					})
					_ = writeMessage(conn, "ERROR_COMMAND_REJECTED", map[string]any{
						"commandId": commandID,
						"reason":    err.Error(),
					})
					continue
				}

				_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
					"commandId": commandID,
					"accepted":  true,
					"towerId":   tower.ID,
					"level":     tower.Level,
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
				err := sellErr
				if err != nil {
					_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
						"commandId": commandID,
						"accepted":  false,
						"reason":    err.Error(),
					})
					_ = writeMessage(conn, "ERROR_COMMAND_REJECTED", map[string]any{
						"commandId": commandID,
						"reason":    err.Error(),
					})
					continue
				}

				_ = writeMessage(conn, "ACK_COMMAND", map[string]any{
					"commandId": commandID,
					"accepted":  true,
					"towerId":   towerID,
					"refund":    refund,
				})
			default:
				_ = writeMessage(conn, "ECHO", map[string]any{"raw": string(raw)})
			}

		case err := <-readErr:
			log.Printf("read failed: %v", err)
			return

		case <-ticker.C:
			if !sharedMatch.isRunning() {
				continue
			}
			sharedMatch.withSimulation(func(sim *Simulation) {
				sim.Step()
			})
			tick, players, towers, enemies := sharedMatch.snapshot()
			if err := writeMessage(conn, "SNAPSHOT_STATE", map[string]any{
				"tick":    tick,
				"players": players,
				"towers":  towers,
				"enemies": enemies,
			}); err != nil {
				log.Printf("snapshot failed: %v", err)
				return
			}
		}
	}
}

func broadcastLobbyState(connectedPlayers int) {
	broadcastAll("LOBBY_STATE", map[string]any{
		"matchId":        sharedMatch.matchID,
		"connectedCount": connectedPlayers,
		"minToStart":     2,
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
			log.Printf("broadcast %s failed: %v", msgType, err)
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
