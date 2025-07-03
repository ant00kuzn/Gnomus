package player

import (
	"net"
	"sync"
)

type Player struct {
	UUID     string
	Name     string
	Conn     net.Conn
	Position [3]float64 // x, y, z
	Yaw      float32
	Pitch    float32
	// Можно добавить инвентарь, здоровье и т.д.
}

var (
	players      = make(map[string]*Player) // UUID -> Player
	playersMutex sync.RWMutex
)

// AddPlayer добавляет игрока в список
func AddPlayer(p *Player) {
	playersMutex.Lock()
	defer playersMutex.Unlock()
	players[p.UUID] = p
}

// RemovePlayer удаляет игрока по UUID
func RemovePlayer(uuid string) {
	playersMutex.Lock()
	defer playersMutex.Unlock()
	delete(players, uuid)
}

// GetPlayer возвращает игрока по UUID
func GetPlayer(uuid string) (*Player, bool) {
	playersMutex.RLock()
	defer playersMutex.RUnlock()
	p, ok := players[uuid]
	return p, ok
}

// GetAllPlayers возвращает срез всех игроков
func GetAllPlayers() []*Player {
	playersMutex.RLock()
	defer playersMutex.RUnlock()
	result := make([]*Player, 0, len(players))
	for _, p := range players {
		result = append(result, p)
	}
	return result
}
