package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type playerSession struct {
	PlayerID  string
	Connected bool
	Ready     bool
	LastSeen  time.Time
}

type matchState struct {
	mu      sync.Mutex
	matchID string
	sim     *Simulation
	running bool
	players map[string]*playerSession
	conns   map[*websocket.Conn]string
}

func newMatchState(matchID string) *matchState {
	return &matchState{
		matchID: matchID,
		sim:     NewSimulation(),
		players: make(map[string]*playerSession),
		conns:   make(map[*websocket.Conn]string),
	}
}

func (m *matchState) registerConnection(playerID string, conn *websocket.Conn) (reconnected bool, connectedPlayers int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.players[playerID]
	reconnected = exists && !session.Connected
	if !exists {
		session = &playerSession{
			PlayerID: playerID,
		}
		m.players[playerID] = session
	}
	session.Connected = true
	session.LastSeen = time.Now()
	m.conns[conn] = playerID

	return reconnected, m.connectedPlayersLocked()
}

func (m *matchState) unregisterConnection(conn *websocket.Conn) (playerID string, connectedPlayers int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	playerID = m.conns[conn]
	delete(m.conns, conn)
	if session, ok := m.players[playerID]; ok {
		session.Connected = false
		session.LastSeen = time.Now()
	}
	return playerID, m.connectedPlayersLocked()
}

func (m *matchState) connectedPlayersLocked() int {
	count := 0
	for _, session := range m.players {
		if session.Connected {
			count++
		}
	}
	return count
}

func (m *matchState) tryStartMatch(minPlayers int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return false
	}
	if m.connectedPlayersLocked() < minPlayers {
		return false
	}
	m.running = true
	return true
}

func (m *matchState) withSimulation(fn func(*Simulation)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m.sim)
}

func (m *matchState) connectionPlayer(conn *websocket.Conn) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.conns[conn]
}

func (m *matchState) setPlayerReady(playerID string, ready bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, ok := m.players[playerID]; ok {
		session.Ready = ready
		return true
	}
	return false
}

func (m *matchState) getConnectedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connectedPlayersLocked()
}

func (m *matchState) allPlayersReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	connectedCount := 0
	readyCount := 0

	for _, session := range m.players {
		if session.Connected {
			connectedCount++
			if session.Ready {
				readyCount++
			}
		}
	}

	// Need at least 1 player and all must be ready
	return connectedCount > 0 && connectedCount == readyCount
}

func (m *matchState) isSimulationRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// Locked variants for use inside withSimulation callbacks (mu already held).
func (m *matchState) isSimulationRunningLocked() bool {
	return m.running
}

func (m *matchState) setSimulationRunningLocked(running bool) {
	m.running = running
}

func (m *matchState) getConnectedCountLocked() int {
	return m.connectedPlayersLocked()
}
