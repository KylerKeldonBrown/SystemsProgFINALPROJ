package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	PORT       = ":8080"
	TIMEOUT    = 60 * time.Second // Timeout for client inactivity
	MAX_BUFFER = 1024
)

type Client struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
}

var (
	clients      = make(map[string]*Client)
	clientsMutex = sync.Mutex{}
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", PORT)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP chat server started on port", PORT)

	go monitorTimeouts()

	buf := make([]byte, MAX_BUFFER)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			continue
		}

		msg := string(buf[:n])
		handleMessage(msg, remoteAddr, conn)
	}
}

// Handles message: adds sender, updates activity, broadcasts to all
func handleMessage(msg string, sender *net.UDPAddr, conn *net.UDPConn) {
	clientsMutex.Lock()
	clientKey := sender.String()

	// Add/update client
	if _, exists := clients[clientKey]; !exists {
		fmt.Printf("New client connected: %s\n", clientKey)
	}
	clients[clientKey] = &Client{Addr: sender, LastSeen: time.Now()}
	clientsMutex.Unlock()

	// Broadcast the message
	broadcastMsg := fmt.Sprintf("[%s]: %s", sender.String(), msg)
	broadcast(broadcastMsg, sender, conn)

	if msg == "PING" {
		// Respond to sender with "PONG"
		_, err := conn.WriteToUDP([]byte("PONG"), sender)
		if err != nil {
			fmt.Println("Error responding to PING:", err)
		}
		return // Do not broadcast PINGs
	}
}

// Broadcast message to all clients (except sender)
func broadcast(message string, sender *net.UDPAddr, conn *net.UDPConn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range clients {
		if client.Addr.String() != sender.String() {
			_, err := conn.WriteToUDP([]byte(message), client.Addr)
			if err != nil {
				fmt.Printf("Error sending to %s: %v\n", client.Addr, err)
			}
		}
	}
}

// Monitor inactive clients and remove them
func monitorTimeouts() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		clientsMutex.Lock()
		now := time.Now()
		for key, client := range clients {
			if now.Sub(client.LastSeen) > TIMEOUT {
				fmt.Printf("Client %s timed out and was removed\n", key)
				delete(clients, key)
			}
		}
		clientsMutex.Unlock()
	}
}
