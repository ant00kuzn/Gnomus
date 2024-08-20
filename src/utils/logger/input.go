package logger

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	NET_SERVER_WORKS sync.Mutex
	SHUTDOWN_SERVER  sync.Mutex
)

// StartInputHandler handles input in a loop
func StartInputHandler() error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Read a line
		scanner.Scan()
		inp := strings.TrimSpace(scanner.Text())

		// Simple realization of stop command
		if strings.HasPrefix(inp, "stop") {
			SHUTDOWN_SERVER.Lock()
			// Sending status to shutdown network server
			SHUTDOWN_SERVER.Unlock()

			fmt.Println("Stopping server...")

			// Running process killing in 6 secs if failed to common shutdown
			go func() {
				time.Sleep(6 * time.Second)
				os.Exit(0)
			}()

			// Waiting for shutdown network's server
			for {
				NET_SERVER_WORKS.Lock()
				if true { // Replace with actual condition
					NET_SERVER_WORKS.Unlock()
					time.Sleep(25 * time.Millisecond)
				} else {
					NET_SERVER_WORKS.Unlock()
					break
				}
			}

			// Disabling the input
			return nil
		}

		// If it's not stop command - display buffer
		fmt.Printf("Entered: %s\n", inp)
	}
}

func main() {
	err := StartInputHandler()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
