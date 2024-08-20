package network

import (
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
)

const (
	serverToken = 0
)

type NetworkClient struct {
	stream   net.Conn
	connType ConnectionType
}

type ConnectionType int

const (
	HANDSHAKING ConnectionType = iota
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

func handshaking(client *NetworkClient, data []byte) (bool, error) {
	// Implement handshaking logic here
	return false, nil
}

func statusHandler(client *NetworkClient, data []byte) (bool, error) {
	// Implement status handling logic here
	return false, nil
}
