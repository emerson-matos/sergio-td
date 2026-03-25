package ws

import "fmt"

type Enemy struct {
	ID string  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	HP int     `json:"hp"`
}

type Tower struct {
	ID       string  `json:"id"`
	OwnerID  string  `json:"ownerId"`
	Type     string  `json:"type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Level    int     `json:"level"`
	Target   string  `json:"targetMode"`
	BaseCost int     `json:"baseCost"`
}

type PlayerState struct {
	ID    string `json:"id"`
	Gold  int    `json:"gold"`
	Lives int    `json:"lives"`
	Score int    `json:"score"`
}

type playerRuntime struct {
	Gold  int
	Lives int
	Score int
}

type Simulation struct {
	tick           int
	nextEnemyID    int
	nextTowerID    int
	spawnEveryTick int
	enemies        []Enemy
	towers         []Tower
	players        map[string]*playerRuntime
}

func NewSimulation() *Simulation {
	return &Simulation{
		spawnEveryTick: 10,
		enemies:        make([]Enemy, 0),
		towers:         make([]Tower, 0),
		players: map[string]*playerRuntime{
			"p_1": {
				Gold:  650,
				Lives: 20,
				Score: 0,
			},
		},
	}
}

func (s *Simulation) Step() {
	s.tick++

	if s.tick%s.spawnEveryTick == 0 {
		s.nextEnemyID++
		s.enemies = append(s.enemies, Enemy{
			ID: fmt.Sprintf("e_%d", s.nextEnemyID),
			X:  0,
			Y:  3,
			HP: 100,
		})
	}

	s.applyTowerCombat()

	alive := make([]Enemy, 0, len(s.enemies))
	for _, enemy := range s.enemies {
		enemy.X += 0.5
		if enemy.X > 20 {
			s.adjustLives("p_1", -1)
			continue
		}
		alive = append(alive, enemy)
	}
	s.enemies = alive
}

func (s *Simulation) Tick() int {
	return s.tick
}

func (s *Simulation) Enemies() []Enemy {
	clone := make([]Enemy, len(s.enemies))
	copy(clone, s.enemies)
	return clone
}

func (s *Simulation) Towers() []Tower {
	clone := make([]Tower, len(s.towers))
	copy(clone, s.towers)
	return clone
}

func (s *Simulation) Players() []PlayerState {
	players := make([]PlayerState, 0, len(s.players))
	for id, player := range s.players {
		players = append(players, PlayerState{
			ID:    id,
			Gold:  player.Gold,
			Lives: player.Lives,
			Score: player.Score,
		})
	}
	return players
}

func (s *Simulation) PlaceTower(playerID, towerType string, x, y float64) (Tower, error) {
	if playerID == "" {
		return Tower{}, fmt.Errorf("missing playerId")
	}
	if towerType == "" {
		return Tower{}, fmt.Errorf("missing towerType")
	}
	if x < 0 || x > 20 || y < 0 || y > 12 {
		return Tower{}, fmt.Errorf("invalid position")
	}

	cost := towerCost(towerType)
	if cost <= 0 {
		return Tower{}, fmt.Errorf("unsupported towerType")
	}

	player := s.ensurePlayer(playerID)
	if player.Gold < cost {
		return Tower{}, fmt.Errorf("insufficient gold")
	}

	for _, existing := range s.towers {
		dx := existing.X - x
		dy := existing.Y - y
		if (dx*dx)+(dy*dy) < 1.0 {
			return Tower{}, fmt.Errorf("position occupied")
		}
	}

	s.nextTowerID++
	tower := Tower{
		ID:       fmt.Sprintf("t_%d", s.nextTowerID),
		OwnerID:  playerID,
		Type:     towerType,
		X:        x,
		Y:        y,
		Level:    1,
		Target:   "first",
		BaseCost: cost,
	}
	s.towers = append(s.towers, tower)
	player.Gold -= cost

	return tower, nil
}

func towerCost(towerType string) int {
	switch towerType {
	case "dart":
		return 100
	default:
		return 0
	}
}

func (s *Simulation) UpgradeTower(playerID, towerID string) (Tower, error) {
	if playerID == "" || towerID == "" {
		return Tower{}, fmt.Errorf("missing playerId or towerId")
	}

	idx := -1
	for i, tower := range s.towers {
		if tower.ID == towerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Tower{}, fmt.Errorf("tower not found")
	}

	tower := s.towers[idx]
	if tower.OwnerID != playerID {
		return Tower{}, fmt.Errorf("tower does not belong to player")
	}
	if tower.Level >= 3 {
		return Tower{}, fmt.Errorf("max level reached")
	}

	upgradeCost := tower.BaseCost * tower.Level
	player := s.ensurePlayer(playerID)
	if player.Gold < upgradeCost {
		return Tower{}, fmt.Errorf("insufficient gold")
	}

	tower.Level++
	s.towers[idx] = tower
	player.Gold -= upgradeCost

	return tower, nil
}

func (s *Simulation) SellTower(playerID, towerID string) (int, error) {
	if playerID == "" || towerID == "" {
		return 0, fmt.Errorf("missing playerId or towerId")
	}

	idx := -1
	for i, tower := range s.towers {
		if tower.ID == towerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return 0, fmt.Errorf("tower not found")
	}

	tower := s.towers[idx]
	if tower.OwnerID != playerID {
		return 0, fmt.Errorf("tower does not belong to player")
	}

	refund := (tower.BaseCost / 2) + ((tower.Level - 1) * tower.BaseCost / 2)
	player := s.ensurePlayer(playerID)
	player.Gold += refund

	s.towers = append(s.towers[:idx], s.towers[idx+1:]...)
	return refund, nil
}

func (s *Simulation) applyTowerCombat() {
	if len(s.towers) == 0 || len(s.enemies) == 0 {
		return
	}

	alive := make([]Enemy, 0, len(s.enemies))
	for _, enemy := range s.enemies {
		enemyAlive := true
		for _, tower := range s.towers {
			if inRange(tower, enemy) {
				damage := 25 * tower.Level
				enemy.HP -= damage
				if enemy.HP <= 0 {
					s.onEnemyKilled(tower.OwnerID)
					enemyAlive = false
					break
				}
			}
		}
		if enemyAlive {
			alive = append(alive, enemy)
		}
	}
	s.enemies = alive
}

func (s *Simulation) onEnemyKilled(playerID string) {
	player := s.ensurePlayer(playerID)
	player.Gold += 10
	player.Score++
}

func (s *Simulation) adjustLives(playerID string, delta int) {
	player := s.ensurePlayer(playerID)
	player.Lives += delta
	if player.Lives < 0 {
		player.Lives = 0
	}
}

func (s *Simulation) ensurePlayer(playerID string) *playerRuntime {
	player, ok := s.players[playerID]
	if ok {
		return player
	}

	player = &playerRuntime{
		Gold:  650,
		Lives: 20,
		Score: 0,
	}
	s.players[playerID] = player
	return player
}

func inRange(tower Tower, enemy Enemy) bool {
	dx := tower.X - enemy.X
	dy := tower.Y - enemy.Y
	return (dx*dx)+(dy*dy) <= 16
}
