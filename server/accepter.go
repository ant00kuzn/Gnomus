package Gnomus

import (
	"Gnomus/config"
	"encoding/json"
	"log"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
)

// Получаем пинг-подкючение(PingList)
func acceptPing(conn net.Conn) {
	// Инициализируем пакет
	var p packet.Packet
	// Пинг или описание, будем принимать только 3 раза
	for i := 0; i < 3; i++ {
		// Читаем пакет
		err := conn.ReadPacket(&p)
		// Если ошибка - перестаём обрабатывать
		if err != nil {
			return
		}
		// Обрабатываем пакет по типу
		switch p.ID {
		case 0x00: // Описание
			// Отправляем пакет со списком
			err = conn.WritePacket(packet.Marshal(0x00, packet.String(listResp())))
		case 0x01: // Пинг
			// Отправляем полученный пакет
			err = conn.WritePacket(p)
		}
		// При ошибке - прекращаем обработку
		if err != nil {
			return
		}
	}
}

// Тип игрока для списка при пинге
type listRespPlayer struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
}

// Генерация JSON строки для ответа на описание
func listResp() string {
	var list struct {
		Version struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		} `json:"version"`
		Players struct {
			Max    int              `json:"max"`
			Online int              `json:"online"`
			Sample []listRespPlayer `json:"sample"`
		} `json:"players"`
		Description chat.Message `json:"description"`
		FavIcon     string       `json:"favicon,gnomus.png"`
	}

	// Устанавливаем дефолтные данные для ответа
	list.Version.Name = "Gnomus"
	list.Version.Protocol = int(config.ProtocolVersion)
	list.Players.Max = 25
	list.Players.Online = 1
	list.Players.Sample = []listRespPlayer{{
		Name: "pov228",
		ID:   uuid.UUID{},
	}}
	list.Description = config.MOTD

	// Превращаем структуру в JSON байты
	data, err := json.Marshal(list)
	if err != nil {
		log.Panic("Ошибка перевода в JSON из обьекта")
	}
	// Возращаем результат в виде строки, переведя из байтов
	return string(data)
}
