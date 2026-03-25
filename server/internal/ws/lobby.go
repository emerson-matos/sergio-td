package ws

import (
	"sync"
	"time"
)

type Room struct {
	ID         string
	Name       string
	Players    map[string]*Player
	MaxPlayers int
	Status     string // "waiting", "playing"
	CreatedAt  time.Time
}

type LobbyServer struct {
	mu         sync.RWMutex
	rooms      map[string]*Room
	nextRoomID int
}

func NewLobbyServer() *LobbyServer {
	return &LobbyServer{
		rooms:      make(map[string]*Room),
		nextRoomID: 1,
	}
}

func (l *LobbyServer) CreateRoom(name string, maxPlayers int) *Room {
	l.mu.Lock()
	defer l.mu.Unlock()

	room := &Room{
		ID:         "room_" + string(rune('0'+l.nextRoomID%10)) + string(rune('0'+(l.nextRoomID/10)%10)),
		Name:       name,
		Players:    make(map[string]*Player),
		MaxPlayers: maxPlayers,
		Status:     "waiting",
		CreatedAt:  time.Now(),
	}
	l.nextRoomID++
	l.rooms[room.ID] = room
	return room
}

func (l *LobbyServer) GetRooms() []RoomInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	rooms := make([]RoomInfo, 0, len(l.rooms))
	for _, room := range l.rooms {
		rooms = append(rooms, RoomInfo{
			ID:          room.ID,
			Name:        room.Name,
			PlayerCount: len(room.Players),
			MaxPlayers:  room.MaxPlayers,
			Status:      room.Status,
		})
	}
	return rooms
}

func (l *LobbyServer) JoinRoom(roomID, playerID string) (*Room, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	room, ok := l.rooms[roomID]
	if !ok {
		return nil, false
	}

	if len(room.Players) >= room.MaxPlayers {
		return nil, false
	}

	room.Players[playerID] = &Player{
		ID:    playerID,
		Ready: false,
	}
	return room, true
}

func (l *LobbyServer) LeaveRoom(roomID, playerID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if room, ok := l.rooms[roomID]; ok {
		delete(room.Players, playerID)
		if len(room.Players) == 0 {
			delete(l.rooms, roomID)
		}
	}
}

func (l *LobbyServer) SetPlayerReady(roomID, playerID string, ready bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if room, ok := l.rooms[roomID]; ok {
		if player, ok := room.Players[playerID]; ok {
			player.Ready = ready
			return true
		}
	}
	return false
}

func (l *LobbyServer) AllPlayersReady(roomID string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if room, ok := l.rooms[roomID]; ok {
		if len(room.Players) == 0 {
			return false
		}
		for _, player := range room.Players {
			if !player.Ready {
				return false
			}
		}
		return true
	}
	return false
}

func (l *LobbyServer) GetRoomPlayers(roomID string) []PlayerInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if room, ok := l.rooms[roomID]; ok {
		players := make([]PlayerInfo, 0, len(room.Players))
		for _, player := range room.Players {
			players = append(players, PlayerInfo{
				ID:    player.ID,
				Ready: player.Ready,
			})
		}
		return players
	}
	return nil
}

func (l *LobbyServer) SetRoomStatus(roomID, status string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if room, ok := l.rooms[roomID]; ok {
		room.Status = status
	}
}

type RoomInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	Status      string `json:"status"`
}

type Player struct {
	ID    string
	Ready bool
}

type PlayerInfo struct {
	ID    string `json:"id"`
	Ready bool   `json:"ready"`
}
