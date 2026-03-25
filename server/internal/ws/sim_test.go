package ws

import "testing"

func TestSimulationSpawnsEnemyEveryTenTicks(t *testing.T) {
	sim := NewSimulation()
	sim.StartWave()

	// Wave 1 has interval=15. Step 14 times → no enemies yet.
	for range 14 {
		sim.Step()
	}
	if got := len(sim.Enemies()); got != 0 {
		t.Fatalf("expected no enemies before tick 15, got %d", got)
	}

	// Tick 15 → first enemy spawns (15 %% 15 == 0)
	sim.Step()
	if got := len(sim.Enemies()); got != 1 {
		t.Fatalf("expected 1 enemy at tick 15, got %d", got)
	}
}

func TestSimulationMovesEnemiesForward(t *testing.T) {
	sim := NewSimulation()
	sim.StartWave()

	// Step 15 times: enemy spawns on tick 15 and moves on that same step.
	for range 15 {
		sim.Step()
	}
	enemies := sim.Enemies()
	if len(enemies) != 1 {
		t.Fatalf("expected 1 enemy, got %d", len(enemies))
	}
	if enemies[0].X < 0.1 {
		t.Fatalf("expected first enemy X to have moved forward, got %.2f", enemies[0].X)
	}
}

func TestPlaceTowerAcceptsValidCommand(t *testing.T) {
	sim := NewSimulation()

	tower, err := sim.PlaceTower("p_1", "raiz", 5, 3)
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

	// raiz costs 100 gold. Player starts with 650. Can place 6 towers.
	for i := range 6 {
		_, err := sim.PlaceTower("p_1", "raiz", float64(i*2), 1)
		if err != nil {
			t.Fatalf("expected placement %d to succeed, got %v", i, err)
		}
	}

	// 7th should fail — only 50 gold left
	_, err := sim.PlaceTower("p_1", "raiz", 14, 2)
	if err == nil {
		t.Fatalf("expected insufficient gold rejection, got nil")
	}
	if err.Error() != "insufficient gold" {
		t.Fatalf("expected insufficient gold error, got %q", err.Error())
	}
}

func TestUpgradeTowerConsumesGoldAndIncreasesLevel(t *testing.T) {
	sim := NewSimulation()

	tower, err := sim.PlaceTower("p_1", "raiz", 2, 2)
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

	tower, err := sim.PlaceTower("p_1", "raiz", 2, 2)
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

func TestTowerCombatKillsEnemyAndRewardsPlayer(t *testing.T) {
	sim := NewSimulation()

	// Use brilhante (50 dmg, range 6, fireRate 15) near the path start.
	_, err := sim.PlaceTower("p_1", "brilhante", 2, 5)
	if err != nil {
		t.Fatalf("place failed: %v", err)
	}

	// Inject a test enemy directly instead of relying on wave spawning.
	sim.nextEnemyID++
	sim.enemies = append(sim.enemies, Enemy{
		ID:           "e_test",
		Type:         "boleto",
		X:            MapWaypoints[0].X,
		Y:            MapWaypoints[0].Y,
		HP:           50,
		MaxHP:        50,
		Speed:        0.5,
		Reward:       10,
		PathIndex:    0,
		PathProgress: 0,
	})
	sim.waveNumber = 1
	sim.waves[0].TotalEnemies = 1
	// Clear wave configs so no additional enemies spawn from the wave.
	sim.waves[0].Configs = nil

	// brilhante does 50 dmg with fireRate=15. One hit kills 50 HP enemy.
	// First fire at tick 15 (tick - lastFire >= fireRate).
	for range 20 {
		sim.Step()
	}

	if got := len(sim.Enemies()); got != 0 {
		t.Fatalf("expected enemy to be killed by tower, got %d alive", got)
	}

	player := findPlayer(t, sim.Players(), "p_1")
	if player.Score < 1 {
		t.Fatalf("expected score to increase after kill, got %d", player.Score)
	}
}

func TestEnemyLeakReducesLives(t *testing.T) {
	sim := NewSimulation()

	// Register player by placing a tower
	_, err := sim.PlaceTower("p_1", "raiz", 19, 0)
	if err != nil {
		t.Fatalf("place failed: %v", err)
	}
	initialLives := findPlayer(t, sim.Players(), "p_1").Lives

	sim.StartWave()

	// Need enough steps for an enemy to traverse the full path (~89 ticks at speed 0.5).
	for range 100 {
		sim.Step()
	}

	player := findPlayer(t, sim.Players(), "p_1")
	if player.Lives >= initialLives {
		t.Fatalf("expected lives to be reduced by leaked enemies, initial=%d current=%d", initialLives, player.Lives)
	}
}

func findPlayer(t *testing.T, players []PlayerState, id string) PlayerState {
	t.Helper()
	for _, player := range players {
		if player.ID == id {
			return player
		}
	}
	t.Fatalf("player %s not found", id)
	return PlayerState{}
}
