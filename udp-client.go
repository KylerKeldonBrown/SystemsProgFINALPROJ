package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const SERVER_ADDR = "localhost:8080"

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", SERVER_ADDR)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to chat server at", SERVER_ADDR)
	fmt.Println("Type messages and press Enter to send.")
	fmt.Println("Type ':ping' to measure latency.")

	// Goroutine to listen for incoming messages
	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error reading from server:", err)
				return
			}
			fmt.Print(string(buf[:n]))
		}
	}()

	// Main loop to read user input and send messages
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if text == ":ping" {
			ping(conn)
			continue
		}

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}

// ping sends a PING message and measures round-trip time
func ping(conn *net.UDPConn) {
	start := time.Now()

	_, err := conn.Write([]byte("PING"))
	if err != nil {
		fmt.Println("Error sending PING:", err)
		return
	}

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // 2s timeout

	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Timeout or error receiving PONG:", err)
		return
	}

	if string(buf[:n]) == "PONG" {
		latency := time.Since(start)
		fmt.Printf("Latency: %v\n", latency)
	} else {
		fmt.Println("Unexpected response:", string(buf[:n]))
	}

	conn.SetReadDeadline(time.Time{}) // Clear timeout
}
