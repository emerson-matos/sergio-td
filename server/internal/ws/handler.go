package ws

import (
	"bytes"
	"encoding/json"
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
	Payload map[string]interface{} `json:"payload,omitempty"`
}

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

	if err := writeMessage(conn, "HELLO_ACK", map[string]any{
		"message": "week2-loop-ready",
	}); err != nil {
		log.Printf("write ack failed: %v", err)
		return
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	rawMessages := make(chan []byte)
	readErr := make(chan error, 1)
	go readLoop(conn, rawMessages, readErr)

	sim := NewSimulation()
	running := false

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
				_ = writeMessage(conn, "ACK_HELLO", map[string]any{"accepted": true})
			case "START_MATCH":
				running = true
				_ = writeMessage(conn, "EVENT_WAVE_STARTED", map[string]any{
					"waveNumber": 1,
				})
			default:
				_ = writeMessage(conn, "ECHO", map[string]any{"raw": string(raw)})
			}

		case err := <-readErr:
			log.Printf("read failed: %v", err)
			return

		case <-ticker.C:
			if !running {
				continue
			}
			sim.Step()
			if err := writeMessage(conn, "SNAPSHOT_STATE", map[string]any{
				"tick":    sim.Tick(),
				"enemies": sim.Enemies(),
			}); err != nil {
				log.Printf("snapshot failed: %v", err)
				return
			}
		}
	}
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
