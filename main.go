package main

import (
	"fmt"
	"github.com/ant00kuzn/Gnomus/config"
	"github.com/ant00kuzn/Gnomus/src/network"
	"github.com/ant00kuzn/Gnomus/src/utils/logger"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	// Initialize logger
	fmt.Println("Starting Gnomus v1.0.0...")
	if err := logger.SetupLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Creating channel for multithreading communication with main's thread and network's thread
	tx := make(chan bool)
	var wg sync.WaitGroup

	// Generate server's address and make it accessible with thread safe
	address := fmt.Sprintf("%s:%d", config.ADDRESS, config.ADDRESS_PORT)

	// Start network in another goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start network
		// If failed to start when return error
		if err := network.NetworkServerStart(address, tx); err != nil {
			log.Printf("Error: %v", err)
			tx <- false
		}
	}()

	// Wait for status from server's network
	if success := <-tx; success {
		// If Server successful started
		log.Printf("Server started at %s", address)
		// Showing about the full launch and showing the time to start
		elapsed := time.Since(start)
		log.Printf("The server was successfully started in %v", elapsed)
	} else {
		// If Failed to start Server
		log.Printf("Failed to start server on %s.", address)
		os.Exit(1)
	}

	// Wait for goroutine to finish
	wg.Wait()
	// Start console input handler(input commands)
	err := logger.StartInputHandler()
	if err != nil {
		return
	}
}

// Custom error
type SimpleError struct {
	Message string
	Err     error
}

func (e *SimpleError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Custom Result with custom Error
type SResult[T any] struct {
	Value T
	Error *SimpleError
}

var (
	NetServerWorks = &sync.Mutex{}
)

type Server struct {
	shutdownMutex sync.Mutex
	Shutdown      bool
}

const mutexLocked = 1

func MutexLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func startInputHandler() error {
	// Input buffer
	inp := ""
	// STDIN - os input
	stdin := os.Stdin

	// loop for infinity handling
	for {
		// Before write buffer we need to clear buffer
		inp = ""
		// Reading a line
		_, err := fmt.Fscanln(stdin, &inp)
		if err != nil {
			return err
		}
		// Clearing input's buffer
		inp = replace(inp, "n", "")
		// Simple realization of stop command
		if startsWith(inp, "stop") {
			// Sending status to shutdown network server
			server := Server{
				Shutdown: false,
			}
			server.shutdownMutex.Lock()
			server.Shutdown = true
			server.shutdownMutex.Unlock()
			fmt.Println("Stopping server...")
			// Running process killing in 6 secs if failed to common shutdown
			go func() {
				time.Sleep(6 * time.Second)
				exec.Command("kill", string(os.Getpid())).Run()
			}()
			// Waiting for shutdown network's server
			for {
				NetServerWorks.Lock()
				if MutexLocked(NetServerWorks) {
					NetServerWorks.Unlock()
					time.Sleep(25 * time.Millisecond)
				} else {
					NetServerWorks.Unlock()
					break
				}
			}
			// Disabling the input
			return nil
		}
		// If it's not stop command - when display buffer
		fmt.Printf("Entered: %s\n", inp)
	}
}

func replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
