package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	serverAddress = "127.0.0.1:8080" // Server address
	numClients    = 100              // Number of concurrent clients to simulate
)

//func main() {
//	var wg sync.WaitGroup
//
//	for i := 0; i < numClients; i++ {
//		wg.Add(1)
//		go func(clientID int) {
//			defer wg.Done()
//			err := runClient(clientID)
//			if err != nil {
//				log.Printf("Client %d error: %v", clientID, err)
//			}
//		}(i)
//		// Optional: Sleep to stagger client connections
//		time.Sleep(100 * time.Millisecond)
//	}
//
//	wg.Wait()
//}

func runClient(clientID int) error {
	// Connect to the server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	log.Printf("Client %d connected to %s", clientID, serverAddress)

	// Send a message to the server
	message := fmt.Sprintf("Hello from client %d", clientID)
	_, err = fmt.Fprintf(conn, message+"\n")
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Receive a response from the server
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	log.Printf("Client %d received: %s", clientID, reply)

	return nil
}
