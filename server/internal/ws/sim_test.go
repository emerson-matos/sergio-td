package ws

import "testing"

func TestSimulationSpawnsEnemyEveryTenTicks(t *testing.T) {
	sim := NewSimulation()

	for range 9 {
		sim.Step()
	}
	if got := len(sim.Enemies()); got != 0 {
		t.Fatalf("expected no enemies before tick 10, got %d", got)
	}

	sim.Step()
	if got := len(sim.Enemies()); got != 1 {
		t.Fatalf("expected 1 enemy at tick 10, got %d", got)
	}
}

func TestSimulationMovesEnemiesForward(t *testing.T) {
	sim := NewSimulation()

	for range 10 {
		sim.Step()
	}
	enemies := sim.Enemies()
	if len(enemies) != 1 {
		t.Fatalf("expected 1 enemy, got %d", len(enemies))
	}
	if enemies[0].X != 0.5 {
		t.Fatalf("expected first enemy X to be 0.5, got %.2f", enemies[0].X)
	}
}
