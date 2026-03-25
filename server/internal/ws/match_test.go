package ws

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestMatchStateReconnectFlow(t *testing.T) {
	m := newMatchState("m_test")

	conn1 := &websocket.Conn{}
	reconnected, connected := m.registerConnection("p_1", conn1)
	if reconnected {
		t.Fatalf("expected first connection to not be reconnect")
	}
	if connected != 1 {
		t.Fatalf("expected 1 connected player, got %d", connected)
	}

	playerID, connected := m.unregisterConnection(conn1)
	if playerID != "p_1" {
		t.Fatalf("expected player p_1 to disconnect, got %s", playerID)
	}
	if connected != 0 {
		t.Fatalf("expected 0 connected players, got %d", connected)
	}

	conn2 := &websocket.Conn{}
	reconnected, connected = m.registerConnection("p_1", conn2)
	if !reconnected {
		t.Fatalf("expected reconnect to be true")
	}
	if connected != 1 {
		t.Fatalf("expected 1 connected player after reconnect, got %d", connected)
	}
}

func TestMatchStateStartNeedsMinimumPlayers(t *testing.T) {
	m := newMatchState("m_test")
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	m.registerConnection("p_1", conn1)
	if started := m.tryStartMatch(2); started {
		t.Fatalf("expected start to fail with only 1 player")
	}

	m.registerConnection("p_2", conn2)
	if started := m.tryStartMatch(2); !started {
		t.Fatalf("expected start to succeed with 2 players")
	}
}
