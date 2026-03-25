package ws

import (
	"fmt"
	"math"
)

type Waypoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

var MapWaypoints = []Waypoint{
	{X: 0, Y: 2},
	{X: 4, Y: 2},
	{X: 4, Y: 8},
	{X: 10, Y: 8},
	{X: 10, Y: 3},
	{X: 16, Y: 3},
	{X: 16, Y: 9},
	{X: 20, Y: 9},
}

type EnemyType struct {
	Name   string  `json:"name"`
	HP     int     `json:"hp"`
	Speed  float64 `json:"speed"`
	Reward int     `json:"reward"`
	Color  string  `json:"color"`
}

var EnemyTypes = map[string]EnemyType{
	"boleto":   {Name: "Boleto", HP: 100, Speed: 0.5, Reward: 10, Color: "red"},
	"imposto":  {Name: "Imposto", HP: 200, Speed: 0.4, Reward: 20, Color: "orange"},
	"taxa":     {Name: "Taxa", HP: 400, Speed: 0.3, Reward: 40, Color: "yellow"},
	"multa":    {Name: "Multa", HP: 800, Speed: 0.25, Reward: 80, Color: "purple"},
	"execucao": {Name: "Execução", HP: 2000, Speed: 0.15, Reward: 200, Color: "darkred"},
}

type TowerType struct {
	Name       string  `json:"name"`
	Cost       int     `json:"cost"`
	Damage     int     `json:"damage"`
	Range      float64 `json:"range"`
	FireRate   int     `json:"fireRate"`
	Projectile string  `json:"projectile"`
	Special    string  `json:"special"`
}

var TowerTypes = map[string]TowerType{
	"raiz":      {Name: "Careca Raiz", Cost: 100, Damage: 25, Range: 4, FireRate: 10, Projectile: "chinelo", Special: ""},
	"brilhante": {Name: "Careca Brilhante", Cost: 250, Damage: 50, Range: 6, FireRate: 15, Projectile: "laser", Special: "piercing"},
	"tank":      {Name: "Careca Tank", Cost: 150, Damage: 10, Range: 2.5, FireRate: 20, Projectile: "escudo", Special: "block"},
	"coach":     {Name: "Careca Coach", Cost: 300, Damage: 0, Range: 5, FireRate: 0, Projectile: "", Special: "buff"},
	"hacker":    {Name: "Careca Hacker", Cost: 200, Damage: 0, Range: 0, FireRate: 0, Projectile: "", Special: "income"},
}

type UpgradePath struct {
	Cost       int     `json:"cost"`
	DamageMult float64 `json:"damageMult"`
	RangeMult  float64 `json:"rangeMult"`
	Special    string  `json:"special"`
}

type Enemy struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	HP           int     `json:"hp"`
	MaxHP        int     `json:"maxHp"`
	Speed        float64 `json:"speed"`
	Reward       int     `json:"reward"`
	PathIndex    int     `json:"pathIndex"`
	PathProgress float64 `json:"pathProgress"`
	Frozen       int     `json:"frozen"`
}

type Tower struct {
	ID         string  `json:"id"`
	OwnerID    string  `json:"ownerId"`
	Type       string  `json:"type"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Level      int     `json:"level"`
	TargetMode string  `json:"targetMode"`
	BaseCost   int     `json:"baseCost"`
	Damage     int     `json:"damage"`
	Range      float64 `json:"range"`
	LastFire   int     `json:"lastFire"`
	Buffed     bool    `json:"buffed"`
}

type PlayerState struct {
	ID    string `json:"id"`
	Gold  int    `json:"gold"`
	Lives int    `json:"lives"`
	Score int    `json:"score"`
}

type WaveConfig struct {
	EnemyType string `json:"enemyType"`
	Count     int    `json:"count"`
	Interval  int    `json:"interval"`
	Spawned   int    `json:"spawned"`
}

type Wave struct {
	Number         int          `json:"number"`
	Configs        []WaveConfig `json:"configs"`
	TotalEnemies   int          `json:"totalEnemies"`
	EnemiesSpawned int          `json:"enemiesSpawned"`
	EnemiesKilled  int          `json:"enemiesKilled"`
	Completed      bool         `json:"completed"`
}

type playerRuntime struct {
	Gold            int
	Lives           int
	Score           int
	TowerGold       int
	TowersBuilt     int
	TotalGoldEarned int
}

type Simulation struct {
	tick        int
	nextEnemyID int
	nextTowerID int
	waveNumber  int
	waves       []Wave
	enemies     []Enemy
	towers      []Tower
	players     map[string]*playerRuntime
	incomeTick  int
	gameOver    bool
}

func NewSimulation() *Simulation {
	return &Simulation{
		enemies: make([]Enemy, 0),
		towers:  make([]Tower, 0),
		waves:   generateWaves(10),
		players: make(map[string]*playerRuntime),
	}
}

func generateWaves(numWaves int) []Wave {
	waves := make([]Wave, numWaves)
	for i := 0; i < numWaves; i++ {
		wave := Wave{Number: i + 1}

		if i < 3 {
			wave.Configs = []WaveConfig{{EnemyType: "boleto", Count: 5 + i*3, Interval: 15}}
		} else if i < 6 {
			wave.Configs = []WaveConfig{
				{EnemyType: "boleto", Count: 8 + i, Interval: 12},
				{EnemyType: "imposto", Count: 3 + (i - 3), Interval: 20},
			}
		} else if i < 9 {
			wave.Configs = []WaveConfig{
				{EnemyType: "boleto", Count: 10 + i, Interval: 10},
				{EnemyType: "imposto", Count: 5 + (i - 6), Interval: 15},
				{EnemyType: "taxa", Count: 2 + (i - 6), Interval: 25},
			}
		} else {
			wave.Configs = []WaveConfig{
				{EnemyType: "boleto", Count: 15 + i, Interval: 8},
				{EnemyType: "imposto", Count: 8 + i, Interval: 12},
				{EnemyType: "taxa", Count: 5 + i, Interval: 18},
				{EnemyType: "multa", Count: 2 + (i - 9), Interval: 30},
			}
		}
		waves[i] = wave
	}
	return waves
}

func (s *Simulation) StartWave() bool {
	if s.waveNumber >= len(s.waves) {
		return false
	}

	currentWave := &s.waves[s.waveNumber]
	currentWave.EnemiesSpawned = 0
	currentWave.EnemiesKilled = 0
	currentWave.Completed = false

	totalEnemies := 0
	for i := range currentWave.Configs {
		totalEnemies += currentWave.Configs[i].Count
		currentWave.Configs[i].Spawned = 0
	}
	currentWave.TotalEnemies = totalEnemies

	s.waveNumber++
	return true
}

func (s *Simulation) GetWaveNumber() int {
	return s.waveNumber
}

func (s *Simulation) IsWaveComplete() bool {
	if s.waveNumber == 0 || s.waveNumber > len(s.waves) {
		return false
	}

	currentWave := &s.waves[s.waveNumber-1]
	return currentWave.EnemiesKilled >= currentWave.TotalEnemies && len(s.enemies) == 0
}

func (s *Simulation) Step() {
	s.tick++
	s.incomeTick++

	s.spawnEnemies()
	s.applyTowerCombat()
	s.moveEnemies()
	s.applyPassiveEffects()
	s.processWaves()
}

func (s *Simulation) spawnEnemies() {
	if s.waveNumber == 0 || s.waveNumber > len(s.waves) {
		return
	}

	currentWave := &s.waves[s.waveNumber-1]

	for i, cfg := range currentWave.Configs {
		spawnTick := s.tick % cfg.Interval
		if spawnTick == 0 && currentWave.Configs[i].Spawned < cfg.Count {
			enemyType, ok := EnemyTypes[cfg.EnemyType]
			if !ok {
				continue
			}

			s.nextEnemyID++
			enemy := Enemy{
				ID:           fmt.Sprintf("e_%d", s.nextEnemyID),
				Type:         cfg.EnemyType,
				X:            MapWaypoints[0].X,
				Y:            MapWaypoints[0].Y,
				HP:           enemyType.HP,
				MaxHP:        enemyType.HP,
				Speed:        enemyType.Speed,
				Reward:       enemyType.Reward,
				PathIndex:    0,
				PathProgress: 0,
			}
			s.enemies = append(s.enemies, enemy)
			currentWave.Configs[i].Spawned++
			currentWave.EnemiesSpawned++
		}
	}
}

func (s *Simulation) moveEnemies() {
	alive := make([]Enemy, 0, len(s.enemies))

	for _, enemy := range s.enemies {
		speed := enemy.Speed
		if enemy.Frozen > 0 {
			speed *= 0.5
			enemy.Frozen--
		}

		enemy.PathProgress += speed

		if enemy.PathIndex < len(MapWaypoints)-1 {
			currentWP := MapWaypoints[enemy.PathIndex]
			nextWP := MapWaypoints[enemy.PathIndex+1]

			dx := nextWP.X - currentWP.X
			dy := nextWP.Y - currentWP.Y
			segmentLen := math.Sqrt(dx*dx + dy*dy)

			if enemy.PathProgress >= segmentLen {
				enemy.PathIndex++
				enemy.PathProgress = 0
			}
		}

		if enemy.PathIndex < len(MapWaypoints)-1 {
			currentWP := MapWaypoints[enemy.PathIndex]
			nextWP := MapWaypoints[enemy.PathIndex+1]

			dx := nextWP.X - currentWP.X
			dy := nextWP.Y - currentWP.Y
			segmentLen := math.Sqrt(dx*dx + dy*dy)

			if segmentLen > 0 {
				t := enemy.PathProgress / segmentLen
				enemy.X = currentWP.X + dx*t
				enemy.Y = currentWP.Y + dy*t
			}
		} else {
			s.leakEnemy(enemy)
			continue
		}

		alive = append(alive, enemy)
	}

	s.enemies = alive
}

func (s *Simulation) leakEnemy(enemy Enemy) {
	for _, player := range s.players {
		player.Lives--
		if player.Lives < 0 {
			player.Lives = 0
		}
	}

	if s.TotalLives() <= 0 {
		s.gameOver = true
	}
}

func (s *Simulation) applyTowerCombat() {
	if len(s.towers) == 0 || len(s.enemies) == 0 {
		return
	}

	for i := range s.towers {
		tower := &s.towers[i]

		towerType, ok := TowerTypes[tower.Type]
		if !ok {
			continue
		}

		if towerType.Special == "buff" || towerType.Special == "income" {
			continue
		}

		if s.tick-tower.LastFire < towerType.FireRate {
			continue
		}

		target := s.findTarget(tower)
		if target == nil {
			continue
		}

		damage := tower.Damage
		if tower.Buffed {
			damage = int(float64(damage) * 1.5)
		}

		target.HP -= damage
		tower.LastFire = s.tick

		if target.HP <= 0 {
			s.killEnemy(target.ID, tower.OwnerID)
		}
	}
}

func (s *Simulation) findTarget(tower *Tower) *Enemy {
	if len(s.enemies) == 0 {
		return nil
	}

	var candidates []Enemy
	for _, e := range s.enemies {
		dx := tower.X - e.X
		dy := tower.Y - e.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		effectiveRange := tower.Range
		if tower.Buffed {
			effectiveRange *= 1.25
		}

		if dist <= effectiveRange {
			candidates = append(candidates, e)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	switch tower.TargetMode {
	case "first":
		best := candidates[0]
		bestProgress := 0.0
		for _, c := range candidates {
			progress := float64(c.PathIndex) + c.PathProgress/10.0
			if progress > bestProgress {
				bestProgress = progress
				best = c
			}
		}
		for i := range s.enemies {
			if s.enemies[i].ID == best.ID {
				return &s.enemies[i]
			}
		}
	case "last":
		best := candidates[0]
		bestProgress := 999.0
		for _, c := range candidates {
			progress := float64(c.PathIndex) + c.PathProgress/10.0
			if progress < bestProgress {
				bestProgress = progress
				best = c
			}
		}
		for i := range s.enemies {
			if s.enemies[i].ID == best.ID {
				return &s.enemies[i]
			}
		}
	case "strong":
		best := candidates[0]
		bestHP := candidates[0].HP
		for _, c := range candidates {
			if c.HP > bestHP {
				bestHP = c.HP
				best = c
			}
		}
		for i := range s.enemies {
			if s.enemies[i].ID == best.ID {
				return &s.enemies[i]
			}
		}
	case "close":
		best := candidates[0]
		bestDist := 999.0
		for _, c := range candidates {
			dx := tower.X - c.X
			dy := tower.Y - c.Y
			dist := dx*dx + dy*dy
			if dist < bestDist {
				bestDist = dist
				best = c
			}
		}
		for i := range s.enemies {
			if s.enemies[i].ID == best.ID {
				return &s.enemies[i]
			}
		}
	}

	return nil
}

func (s *Simulation) killEnemy(enemyID, killerID string) {
	for i, e := range s.enemies {
		if e.ID == enemyID {
			enemy := e
			s.enemies = append(s.enemies[:i], s.enemies[i+1:]...)

			if s.waveNumber > 0 && s.waveNumber <= len(s.waves) {
				s.waves[s.waveNumber-1].EnemiesKilled++
			}

			player := s.ensurePlayer(killerID)
			player.Gold += enemy.Reward
			player.Score++
			player.TotalGoldEarned += enemy.Reward
			return
		}
	}
}

func (s *Simulation) applyPassiveEffects() {
	for i := range s.towers {
		tower := &s.towers[i]

		towerType, ok := TowerTypes[tower.Type]
		if !ok {
			continue
		}

		if towerType.Special == "buff" {
			for j := range s.towers {
				if i == j {
					continue
				}
				other := &s.towers[j]
				dx := tower.X - other.X
				dy := tower.Y - other.Y
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist <= tower.Range {
					other.Buffed = true
				}
			}
		}

		if towerType.Special == "income" {
			if s.incomeTick%50 == 0 {
				player := s.ensurePlayer(tower.OwnerID)
				player.Gold += 25 + (tower.Level * 10)
			}
		}
	}
}

func (s *Simulation) processWaves() {
	if s.waveNumber == 0 {
		return
	}

	if s.waveNumber > len(s.waves) && len(s.enemies) == 0 {
		s.gameOver = true
		return
	}

	if s.waveNumber <= len(s.waves) {
		currentWave := &s.waves[s.waveNumber-1]

		if currentWave.EnemiesKilled >= currentWave.TotalEnemies && len(s.enemies) == 0 && !currentWave.Completed {
			currentWave.Completed = true
			// Auto-start next wave or advance past all waves for victory
			if s.waveNumber < len(s.waves) {
				s.StartWave()
			} else {
				s.waveNumber++
			}
		}
	}
}

func (s *Simulation) Tick() int {
	return s.tick
}

func (s *Simulation) WaveNumber() int {
	return s.waveNumber
}

func (s *Simulation) Waves() []Wave {
	return s.waves
}

func (s *Simulation) GameOver() bool {
	return s.gameOver
}

func (s *Simulation) Victory() bool {
	return s.gameOver && s.waveNumber > len(s.waves) && len(s.enemies) == 0
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

func (s *Simulation) TotalLives() int {
	total := 0
	for _, player := range s.players {
		total += player.Lives
	}
	return total
}

func (s *Simulation) PlaceTower(playerID, towerType string, x, y float64) (Tower, error) {
	if playerID == "" {
		return Tower{}, fmt.Errorf("missing playerId")
	}
	if towerType == "" {
		return Tower{}, fmt.Errorf("missing towerType")
	}

	towerDef, ok := TowerTypes[towerType]
	if !ok {
		return Tower{}, fmt.Errorf("unsupported towerType: %s", towerType)
	}

	if x < 0 || x > 20 || y < 0 || y > 11 {
		return Tower{}, fmt.Errorf("invalid position")
	}

	player := s.ensurePlayer(playerID)
	if player.Gold < towerDef.Cost {
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
		ID:         fmt.Sprintf("t_%d", s.nextTowerID),
		OwnerID:    playerID,
		Type:       towerType,
		X:          x,
		Y:          y,
		Level:      1,
		TargetMode: "first",
		BaseCost:   towerDef.Cost,
		Damage:     towerDef.Damage,
		Range:      towerDef.Range,
	}
	s.towers = append(s.towers, tower)
	player.Gold -= towerDef.Cost
	player.TowersBuilt++

	return tower, nil
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

	towerDef := TowerTypes[tower.Type]
	upgradeCost := tower.BaseCost * tower.Level
	player := s.ensurePlayer(playerID)
	if player.Gold < upgradeCost {
		return Tower{}, fmt.Errorf("insufficient gold")
	}

	tower.Level++
	tower.Damage = int(float64(towerDef.Damage) * (1 + float64(tower.Level-1)*0.5))
	tower.Range = towerDef.Range * (1 + float64(tower.Level-1)*0.1)

	s.towers[idx] = tower
	player.Gold -= upgradeCost

	return tower, nil
}

func (s *Simulation) SetTargetMode(playerID, towerID, mode string) error {
	if playerID == "" || towerID == "" {
		return fmt.Errorf("missing playerId or towerId")
	}

	validModes := map[string]bool{
		"first": true, "last": true, "strong": true, "close": true,
	}
	if !validModes[mode] {
		return fmt.Errorf("invalid target mode")
	}

	for i, tower := range s.towers {
		if tower.ID == towerID && tower.OwnerID == playerID {
			s.towers[i].TargetMode = mode
			return nil
		}
	}

	return fmt.Errorf("tower not found")
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

func (s *Simulation) Reset() {
	s.tick = 0
	s.nextEnemyID = 0
	s.nextTowerID = 0
	s.waveNumber = 0
	s.waves = generateWaves(10)
	s.enemies = make([]Enemy, 0)
	s.towers = make([]Tower, 0)
	s.players = make(map[string]*playerRuntime)
	s.incomeTick = 0
	s.gameOver = false
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

type PlayerMatchStats struct {
	ID              string `json:"id"`
	Score           int    `json:"score"`
	TowersBuilt     int    `json:"towersBuilt"`
	TotalGoldEarned int    `json:"totalGoldEarned"`
}

func (s *Simulation) MatchStats() []PlayerMatchStats {
	stats := make([]PlayerMatchStats, 0, len(s.players))
	for id, p := range s.players {
		stats = append(stats, PlayerMatchStats{
			ID:              id,
			Score:           p.Score,
			TowersBuilt:     p.TowersBuilt,
			TotalGoldEarned: p.TotalGoldEarned,
		})
	}
	return stats
}

func GetWaypoints() []Waypoint {
	return MapWaypoints
}

func GetTowerTypes() map[string]TowerType {
	return TowerTypes
}

func GetEnemyTypes() map[string]EnemyType {
	return EnemyTypes
}
