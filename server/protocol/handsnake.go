package protocol

import (
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

// ReadHandSnake - чтение HandSnake пакета( https://wiki.vg/Protocol#Handshake )
func ReadHandSnake(conn net.Conn) (protocol, intention int32, address string, port uint16, err error) {
	// Переменные пакета
	var (
		p                   packet.Packet
		Protocol, NextState packet.VarInt
		ServerAddress       packet.String
		ServerPort          packet.UnsignedShort
	)
	// Читаем входящий пакет и при ошибке ничего не возращаем
	if err = conn.ReadPacket(&p); err != nil {
		return
	}
	// Читаем содержимое пакета
	err = p.Scan(&Protocol, &ServerAddress, &ServerPort, &NextState)
	// Возращаем результат чтения в привычной форме для работы(примитивные типы)
	return int32(Protocol), int32(NextState), string(ServerAddress), uint16(ServerPort), err
}
