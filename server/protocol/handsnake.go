package protocol

import (
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

// ReadHandSnake - чтение HandSnake пакета (https://wiki.vg/Protocol#Handshake)
// Теперь поддерживает как стандартный handshake, так и legacy ping (0xFE, 0xFA)
func ReadHandSnake(conn net.Conn) (protocol, intention int32, address string, port uint16, err error) {
	var (
		p                   packet.Packet
		Protocol, NextState packet.VarInt
		ServerAddress       packet.String
		ServerPort          packet.UnsignedShort
	)
	// Читаем первый байт для определения типа пакета
	peek := make([]byte, 1)
	_, err = conn.Read(peek)
	if err != nil {
		return
	}
	if peek[0] == 0xFE {
		// Legacy ping (1.6 и ниже)
		var legacy [2]byte
		conn.Read(legacy[:])
		if legacy[0] == 0x01 && legacy[1] == 0xFA {
			// Можно обработать legacy ping (MC|PingHost)
			intention = -1 // Специальное значение для legacy
			return
		}
		intention = -2 // Просто ping
		return
	}
	// Если не legacy, возвращаем байт обратно в поток
	conn = net.NewConnWithPrepend(conn, peek)
	// Читаем стандартный handshake
	if err = conn.ReadPacket(&p); err != nil {
		return
	}
	err = p.Scan(&Protocol, &ServerAddress, &ServerPort, &NextState)
	return int32(Protocol), int32(NextState), string(ServerAddress), uint16(ServerPort), err
}
