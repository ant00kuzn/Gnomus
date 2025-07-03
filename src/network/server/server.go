package network

import (
	"Gnomus/src/network/player"
	"Gnomus/src/network/world"
	"encoding/hex"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/cornelk/hashmap"
)

// Global variables
var (
	shutdownServer  sync.Mutex
	netServerWorks  sync.Mutex
	shutdownServerB bool
	netServerWorksB bool
	tickStop        = make(chan struct{})
)

const (
	serverToken  = 0
	tickInterval = 50 * time.Millisecond // 20 тиков в секунду
)

type NetworkClient struct {
	stream   net.Conn
	connType ConnectionType
}

type ConnectionType int

const (
	HANDSHAKING ConnectionType = iota
	STATUS
	LOGIN
	PLAY
)

func nextToken(current *int) int {
	next := *current
	*current++
	return next
}

func NetworkServerStart(address string, tx chan<- bool) error {
	// Converting String's address to net.Addr
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	// Starting a Network Listener
	listener, err := net.Listen("tcp", addr.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	// Creating a list of connections
	connections := hashmap.New[int, *NetworkClient]()

	// Creating a variable with latest token
	uniqueToken := serverToken + 1

	// Send over the channel that the server has been successfully started
	tx <- true

	// Network Events getting timeout
	timeout := 10 * time.Millisecond

	// Infinity loop to handle events
	for {
		// Checks whether it is necessary to shutdown the network server
		shutdownServer.Lock()
		if shutdownServerB {
			netServerWorks.Lock()
			netServerWorksB = false
			netServerWorks.Unlock()
			log.Println("Network Server Stopped!")
			shutdownServer.Unlock()
			return nil
		}
		shutdownServer.Unlock()

		// Set deadline for accepting new connections
		listener.(*net.TCPListener).SetDeadline(time.Now().Add(timeout))

		// Accept new connection
		conn, err := listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println("Error accepting connection:", err)
			continue
		}

		// Generating new token for this connection
		token := nextToken(&uniqueToken)

		// Pushing connection into connection's list
		connections.Set(token, &NetworkClient{
			stream:   conn,
			connType: HANDSHAKING,
		})

		// Handle the connection
		go handleConnection(connections, token, conn)
	}
}

func handleConnection(connections *hashmap.Map[int, *NetworkClient], token int, conn net.Conn) {
	defer func() {
		connections.Del(token)
		_ = conn.Close()
	}()

	for {
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if err != nil {
			if err != io.EOF {
				log.Println("read error:", err)
			}
			return
		}

		client, ok := connections.Get(token)
		if !ok {
			return
		}

		var handler func(*NetworkClient, []byte) (bool, error)
		switch client.connType {
		case HANDSHAKING:
			handler = handshaking
		case STATUS:
			handler = statusHandler
		case LOGIN:
			handler = loginHandler
		case PLAY:
			handler = playHandler
		default:
			handler = statusHandler
		}

		done, err := handler(client, data[:n])
		if err != nil {
			log.Println("handler error:", err)
			return
		}

		if done {
			return
		}
	}
}

// parseHandshakeNextState корректно парсит nextState из handshake-пакета по протоколу Minecraft
func parseHandshakeNextState(data []byte) (int, error) {
	idx := 0
	// Пропускаем packet length (VarInt)
	_, n := decodeVarInt(data)
	idx += n
	// Пропускаем packet id (VarInt)
	_, n = decodeVarInt(data[idx:])
	idx += n
	// Пропускаем protocol version (VarInt)
	_, n = decodeVarInt(data[idx:])
	idx += n
	// Пропускаем server address (VarInt + bytes)
	addrLen, n := decodeVarInt(data[idx:])
	idx += n + addrLen
	// Пропускаем port (2 байта)
	idx += 2
	// Теперь nextState (VarInt)
	nextState, _ := decodeVarInt(data[idx:])
	return nextState, nil
}

// decodeVarInt декодирует VarInt из байтов
func decodeVarInt(data []byte) (int, int) {
	var num int
	var shift uint
	for i, b := range data {
		num |= int(b&0x7F) << shift
		if b&0x80 == 0 {
			return num, i + 1
		}
		shift += 7
	}
	return 0, 0
}

func handshaking(client *NetworkClient, data []byte) (bool, error) {
	log.Printf("[HANDSHAKING | INFO] Received from %v: %x", client.stream.RemoteAddr(), data)
	if len(data) < 1 {
		return true, nil
	}
	nextState, _ := parseHandshakeNextState(data)
	switch nextState {
	case 1:
		client.connType = STATUS
		log.Printf("[HANDSHAKING | INFO] Switch to STATUS for %v", client.stream.RemoteAddr())
	case 2:
		client.connType = LOGIN
		log.Printf("[HANDSHAKING | INFO] Switch to LOGIN for %v", client.stream.RemoteAddr())
	default:
		log.Printf("[HANDSHAKING | ERROR] Unknown next state: %d", nextState)
	}
	return false, nil
}

func statusHandler(client *NetworkClient, data []byte) (bool, error) {
	log.Printf("[STATUS] Received from %v: %x", client.stream.RemoteAddr(), data)
	if len(data) < 1 {
		return true, nil
	}
	packetID := data[0]
	if packetID == 0x00 {
		// Status request
		motd := `{"version":{"name":"1.21.1","protocol":763},"players":{"max":100,"online":0},"description":{"text":"§aGnomus Minecraft Server"}}`
		resp := buildStatusResponse(motd)
		_, err := client.stream.Write(resp)
		return true, err
	} else if packetID == 0x01 {
		// Ping request
		// Отправить обратно полученный payload (long)
		resp := buildPongResponse(data[1:])
		_, err := client.stream.Write(resp)
		return true, err
	}
	return true, nil
}

func loginHandler(client *NetworkClient, data []byte) (bool, error) {
	log.Printf("[LOGIN] Received from %v: %x", client.stream.RemoteAddr(), data)
	if len(data) < 1 {
		return true, nil
	}
	packetID := data[0]
	if packetID == 0x00 {
		// Login Start (обычно содержит имя игрока)
		uuid := "00000000-0000-0000-0000-000000000000"
		name := "Player"
		resp := buildLoginSuccess(uuid, name)
		_, err := client.stream.Write(resp)
		if err != nil {
			return true, err
		}
		client.connType = PLAY
		log.Printf("[LOGIN] Switch to PLAY for %v", client.stream.RemoteAddr())
		// Добавляем игрока в глобальный список
		p := &player.Player{
			UUID:     uuid,
			Name:     name,
			Conn:     client.stream,
			Position: [3]float64{0, 64, 0},
		}
		player.AddPlayer(p)
		// Генерируем и отправляем чанк игроку (0,0)
		chunk := world.GetChunk(0, 0)
		packet := world.SerializeChunkToPacket(0, 0, chunk)
		_, err = client.stream.Write(packet)
		if err != nil {
			log.Printf("[LOGIN] Failed to send chunk: %v", err)
		} else {
			log.Printf("[LOGIN] Sent chunk (0,0) to %s", name)
		}
		return false, nil
	}
	return true, nil
}

func playHandler(client *NetworkClient, data []byte) (bool, error) {
	log.Printf("[PLAY] Received from %v: %x", client.stream.RemoteAddr(), data)
	if len(data) < 1 {
		return false, nil
	}
	packetID := data[0]
	switch packetID {
	case 0x0F: // Example: Client Settings (vanilla)
		log.Printf("[PLAY] Client Settings packet from %v", client.stream.RemoteAddr())
		// Можно разобрать настройки клиента
	case 0x14: // Example: Chat Message (vanilla)
		log.Printf("[PLAY] Chat Message packet from %v", client.stream.RemoteAddr())
		// Можно разобрать и отправить чат другим игрокам
	case 0x1A: // Example: Player Position (vanilla)
		log.Printf("[PLAY] Player Position packet from %v", client.stream.RemoteAddr())
		// Можно обновить позицию игрока
	case 0x21: // Keep Alive Response
		log.Printf("[PLAY] Keep Alive Response from %v", client.stream.RemoteAddr())
		// Можно обновить время активности игрока
	default:
		log.Printf("[PLAY] Unknown packet id 0x%X from %v", packetID, client.stream.RemoteAddr())
	}
	return false, nil
}

// buildStatusResponse формирует пакет ответа на status request
func buildStatusResponse(json string) []byte {
	// [packet length][packet id][json length][json]
	jsonBytes := []byte(json)
	var packet []byte
	packet = append(packet, encodeVarInt(1+lenVarInt(len(jsonBytes))+len(jsonBytes))...)
	packet = append(packet, 0x00) // packet id
	packet = append(packet, encodeVarInt(len(jsonBytes))...)
	packet = append(packet, jsonBytes...)
	return packet
}

// buildPongResponse формирует пакет ответа на ping
func buildPongResponse(payload []byte) []byte {
	// [packet length][packet id][payload]
	var packet []byte
	packet = append(packet, encodeVarInt(1+len(payload))...)
	packet = append(packet, 0x01) // packet id
	packet = append(packet, payload...)
	return packet
}

// buildLoginSuccess формирует пакет Login Success
func buildLoginSuccess(uuid, name string) []byte {
	// UUID должен быть 16 байт (raw), а не строкой
	uuidRaw := uuidStringToBytes(uuid)
	nameBytes := []byte(name)
	var packet []byte
	// packet id (0x02) + uuid (16) + name (varint+bytes)
	packetData := []byte{0x02}
	packetData = append(packetData, uuidRaw...)
	packetData = append(packetData, encodeVarInt(len(nameBytes))...)
	packetData = append(packetData, nameBytes...)
	packet = append(packet, encodeVarInt(len(packetData))...)
	packet = append(packet, packetData...)
	return packet
}

// uuidStringToBytes конвертирует UUID-строку в 16 raw байт
func uuidStringToBytes(uuid string) []byte {
	uuid = removeDashes(uuid)
	b, err := hex.DecodeString(uuid)
	if err != nil || len(b) != 16 {
		return make([]byte, 16)
	}
	return b
}

func removeDashes(s string) string {
	res := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			res = append(res, s[i])
		}
	}
	return string(res)
}

// buildKeepAlive формирует пакет Keep Alive
func buildKeepAlive(id int64) []byte {
	// [packet length][packet id][keep alive id (long)]
	var packet []byte
	packet = append(packet, encodeVarInt(9+1)...) // 1 (id) + 8 (long)
	packet = append(packet, 0x21)                 // packet id
	for i := 7; i >= 0; i-- {
		packet = append(packet, byte(id>>(8*i)))
	}
	return packet
}

// --- Вспомогательные функции для VarInt ---
func encodeVarInt(value int) []byte {
	var buf []byte
	for {
		b := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if value == 0 {
			break
		}
	}
	return buf
}

func lenVarInt(value int) int {
	l := 0
	for {
		l++
		value >>= 7
		if value == 0 {
			break
		}
	}
	return l
}

// StartTickLoop запускает игровой цикл (tick loop)
func StartTickLoop() {
	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()
		tick := 0
		for {
			select {
			case <-tickStop:
				log.Println("Tick loop stopped")
				return
			case <-ticker.C:
				tick++
				// Пример: логируем всех игроков
				players := player.GetAllPlayers()
				log.Printf("Tick %d: online %d players", tick, len(players))
				// Здесь можно обновлять позиции, отправлять пакеты и т.д.
			}
		}
	}()
}

// StopTickLoop останавливает игровой цикл
func StopTickLoop() {
	close(tickStop)
}
