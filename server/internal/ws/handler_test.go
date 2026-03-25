package ws

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestBroadcastAllOutsideWithSimulation verifies that broadcastAll is never
// called while the match mutex is held (inside withSimulation). Go's
// sync.Mutex is not reentrant, so broadcastAll (which also locks mu) would
// deadlock if called from a withSimulation callback.
//
// This test reproduces the exact ticker pattern: gather data inside
// withSimulation, then broadcast outside. If someone accidentally moves a
// broadcastAll call back inside the lock, the test will deadlock and fail
// via timeout.
func TestBroadcastAllOutsideWithSimulation(t *testing.T) {
	m := newMatchState("m_deadlock_test")
	conn := &websocket.Conn{}
	m.registerConnection("p_1", conn)

	done := make(chan struct{})

	go func() {
		defer close(done)

		// Simulate the ticker pattern: capture data inside the lock,
		// broadcast outside.
		var snapPayload map[string]any
		var matchStarted bool
		var waveStartNum int

		m.withSimulation(func(sim *Simulation) {
			if !m.isSimulationRunningLocked() && m.getConnectedCountLocked() >= 1 {
				m.setSimulationRunningLocked(true)
				sim.StartWave()
				matchStarted = true
				waveStartNum = sim.WaveNumber()
			}

			if m.isSimulationRunningLocked() {
				sim.Step()
			}

			snapPayload = map[string]any{
				"tick":       sim.Tick(),
				"waveNumber": sim.WaveNumber(),
				"players":    sim.Players(),
				"towers":     sim.Towers(),
				"enemies":    sim.Enemies(),
			}
		})

		// These calls lock mu internally. If they were inside
		// withSimulation above, this goroutine would never reach here.
		if matchStarted {
			broadcastAll("EVENT_MATCH_STARTED", map[string]any{})
			broadcastAll("EVENT_WAVE_STARTED", map[string]any{"waveNumber": waveStartNum})
		}
		broadcastAll("SNAPSHOT_STATE", snapPayload)
	}()

	select {
	case <-done:
		// Success: no deadlock.
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock detected: broadcastAll is likely called inside withSimulation (mu held twice)")
	}
}

// TestStartWaveDoesNotDeadlock verifies the START_WAVE handler pattern:
// sim.StartWave() inside withSimulation, broadcastAll outside.
func TestStartWaveDoesNotDeadlock(t *testing.T) {
	m := newMatchState("m_startwave_test")
	conn := &websocket.Conn{}
	m.registerConnection("p_1", conn)

	done := make(chan struct{})

	go func() {
		defer close(done)

		var waveNum int
		var started bool
		m.withSimulation(func(sim *Simulation) {
			started = sim.StartWave()
			if started {
				waveNum = sim.WaveNumber()
			}
		})
		if started {
			broadcastAll("EVENT_WAVE_STARTED", map[string]any{"waveNumber": waveNum})
		}
	}()

	select {
	case <-done:
		// Success: no deadlock.
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock detected: broadcastAll is likely called inside withSimulation")
	}
}
