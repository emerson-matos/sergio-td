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

func TestPlaceTowerAcceptsValidCommand(t *testing.T) {
	sim := NewSimulation()

	tower, err := sim.PlaceTower("p_1", "dart", 5, 3)
	if err != nil {
		t.Fatalf("expected placement to succeed, got error: %v", err)
	}

	if tower.ID == "" {
		t.Fatalf("expected tower id to be assigned")
	}
	if got := len(sim.Towers()); got != 1 {
		t.Fatalf("expected 1 tower, got %d", got)
	}
}

func TestPlaceTowerRejectsInsufficientGold(t *testing.T) {
	sim := NewSimulation()

	for i := range 6 {
		_, err := sim.PlaceTower("p_1", "dart", float64(i*2), 1)
		if err != nil {
			t.Fatalf("expected placement %d to succeed, got %v", i, err)
		}
	}

	_, err := sim.PlaceTower("p_1", "dart", 14, 2)
	if err == nil {
		t.Fatalf("expected insufficient gold rejection, got nil")
	}
	if err.Error() != "insufficient gold" {
		t.Fatalf("expected insufficient gold error, got %q", err.Error())
	}
}

func TestUpgradeTowerConsumesGoldAndIncreasesLevel(t *testing.T) {
	sim := NewSimulation()

	tower, err := sim.PlaceTower("p_1", "dart", 2, 2)
	if err != nil {
		t.Fatalf("place failed: %v", err)
	}

	updatedTower, err := sim.UpgradeTower("p_1", tower.ID)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}

	if updatedTower.Level != 2 {
		t.Fatalf("expected level 2, got %d", updatedTower.Level)
	}
}

func TestSellTowerRemovesTowerAndReturnsRefund(t *testing.T) {
	sim := NewSimulation()

	tower, err := sim.PlaceTower("p_1", "dart", 2, 2)
	if err != nil {
		t.Fatalf("place failed: %v", err)
	}

	refund, err := sim.SellTower("p_1", tower.ID)
	if err != nil {
		t.Fatalf("sell failed: %v", err)
	}

	if refund != 50 {
		t.Fatalf("expected refund 50, got %d", refund)
	}
	if got := len(sim.Towers()); got != 0 {
		t.Fatalf("expected no towers after sell, got %d", got)
	}
}
