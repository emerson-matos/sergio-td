package ws

import "fmt"

type Enemy struct {
	ID string  `json:"id"`
	X  float64 `json:"x"`
	HP int     `json:"hp"`
}

type Simulation struct {
	tick           int
	nextEnemyID    int
	spawnEveryTick int
	enemies        []Enemy
}

func NewSimulation() *Simulation {
	return &Simulation{
		spawnEveryTick: 10,
		enemies:        make([]Enemy, 0),
	}
}

func (s *Simulation) Step() {
	s.tick++

	if s.tick%s.spawnEveryTick == 0 {
		s.nextEnemyID++
		s.enemies = append(s.enemies, Enemy{
			ID: fmt.Sprintf("e_%d", s.nextEnemyID),
			X:  0,
			HP: 100,
		})
	}

	alive := make([]Enemy, 0, len(s.enemies))
	for _, enemy := range s.enemies {
		enemy.X += 0.5
		if enemy.X <= 20 {
			alive = append(alive, enemy)
		}
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
