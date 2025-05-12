package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	maxMessageSize   = 1024
	inactivityPeriod = 60 * time.Second
	logDir           = "client_logs"
	packetTimeout    = 3 * time.Second // Timeout for packet acknowledgment
)

var (
	clientCount     int
	clientCountLock sync.Mutex
)

type Client struct {
	conn             net.Conn
	name             string
	logFile          *os.File
	lastSeen         time.Time
	sentMessages     int
	receivedMessages int
	totalBytes       int64 // Track total bytes sent
	LastPingLatency  time.Duration
}

type Server struct {
	clients         map[net.Conn]*Client
	register        chan *Client
	unregister      chan *Client
	broadcast       chan string
	mu              sync.Mutex
	packetLoss      int
	totalBytesSent  int64
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[net.Conn]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan string),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.conn] = client
			s.mu.Unlock()
			logEvent(fmt.Sprintf("Client connected: %s", client.name))
		case client := <-s.unregister:
			s.mu.Lock()
			if c, ok := s.clients[client.conn]; ok {
				c.logFile.Close()
				delete(s.clients, client.conn)
				client.conn.Close()
				logEvent(fmt.Sprintf("Client disconnected: %s", client.name))
			}
			s.mu.Unlock()
		case message := <-s.broadcast:
			s.mu.Lock()
			for _, client := range s.clients {
				_, err := fmt.Fprintln(client.conn, message)
				if err != nil {
					log.Printf("Error sending message to %s: %v", client.name, err)
				}
			}
			s.mu.Unlock()
		}
	}
}

func handleConnection(conn net.Conn, server *Server) {
	defer func() {
		server.unregister <- &Client{conn: conn}
	}()

	clientAddr := conn.RemoteAddr().String()
	os.MkdirAll(logDir, 0755)
	safeFileName := strings.ReplaceAll(clientAddr, ":", "_") + ".log"
	logFilePath := filepath.Join(logDir, safeFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(conn, "Server error: unable to open log file\n")
		conn.Close()
		return
	}

	client := &Client{
		conn:             conn,
		name:             clientAddr,
		logFile:          logFile,
		lastSeen:         time.Now(),
		sentMessages:     0,
		receivedMessages: 0,
		LastPingLatency:  0,
	}
	server.register <- client

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, maxMessageSize), maxMessageSize)

	timer := time.NewTimer(inactivityPeriod)
	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(inactivityPeriod)
	}
	done := make(chan bool)

	go func() {
		for scanner.Scan() {
			resetTimer()
			input := strings.TrimSpace(scanner.Text())

			if len(input) > maxMessageSize {
				conn.Write([]byte("Message too long.\n"))
				input = input[:maxMessageSize]
			}

			client.logFile.WriteString(fmt.Sprintf("%s: %s\n", time.Now().Format(time.RFC3339), input))

			switch input {
			case "":
				conn.Write([]byte("Wassup...\n"))
			case "GIMME 3":
				conn.Write([]byte("Brrrrrrrrrrrr!\n"))
			case "bye", "/quit":
				conn.Write([]byte("Later!\n"))
				done <- true
				return
			case "/time":
				conn.Write([]byte(time.Now().Format(time.RFC1123) + "\n"))
			case "/date":
				conn.Write([]byte(time.Now().Format("2006-01-02") + "\n"))
			case "/joke":
				conn.Write([]byte("If you wanted a joke you should have made one yourself\n"))
			case "/ping":
				start := time.Now()
				conn.Write([]byte("Pong!\n"))
				latency := time.Since(start)
				client.LastPingLatency = latency
				conn.Write([]byte(fmt.Sprintf("Latency: %s\n", latency)))
			case "/clients":
				clientCountLock.Lock()
				count := clientCount
				clientCountLock.Unlock()
				conn.Write([]byte(fmt.Sprintf("Connected clients: %d\n", count)))
			case "/help":
				conn.Write([]byte("Available commands:\n" +
					"/echo [message] - Echoes back your message\n" +
					"/time - Shows current server time\n" +
					"/date - Shows current server date\n" +
					"/joke - Tells a joke\n" +
					"/ping - Shows latency\n" +
					"/clients - Number of connected clients\n" +
					"/quit or bye - Disconnects you\n"))
			default:
				if strings.HasPrefix(input, "/echo ") {
					conn.Write([]byte(strings.TrimPrefix(input, "/echo ") + "\n"))
				} else {
					server.broadcast <- fmt.Sprintf("%s: %s", client.name, input)
				}
			}
			client.sentMessages++
			if strings.HasPrefix(input, "/echo ") || input == "/ping" {
				client.receivedMessages++ // only count it if it’s something we replied to
			}

		}
		done <- true
	}()

	clientCountLock.Lock()
	clientCount++
	clientCountLock.Unlock()

	select {
	case <-timer.C:
		conn.Write([]byte("Disconnected due to inactivity\n"))
		logEvent(fmt.Sprintf("Client disconnected (timeout): %s", clientAddr))
	case <-done:
		logEvent(fmt.Sprintf("Client disconnected: %s", clientAddr))
	}

	clientCountLock.Lock()
	clientCount--
	clientCountLock.Unlock()

	logMetrics(client)
}

func logMetrics(client *Client) {
	csvFile := "client_metrics.csv"
	fileExists := false
	if _, err := os.Stat(csvFile); err == nil {
		fileExists = true
	}

	f, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	if !fileExists {
		writer.WriteString("Client,SentMessages,ReceivedMessages,PacketLoss(%),Throughput(msg/sec),SessionDuration(seconds),Latency(ms),Timestamp\n")
	}

	duration := time.Since(client.lastSeen)
	throughput := float64(client.sentMessages) / duration.Seconds()

	var packetLoss float64
	if client.sentMessages > 0 {
		packetLoss = float64(client.sentMessages-client.receivedMessages) / float64(client.sentMessages) * 100
	}

	line := fmt.Sprintf("%s,%d,%d,%.2f,%.2f,%.2f,%.2f,%s\n",
		client.name,
		client.sentMessages,
		client.receivedMessages,
		packetLoss,
		throughput,
		duration.Seconds(),
		client.LastPingLatency.Seconds()*1000, // Latency in ms
		time.Now().Format(time.RFC3339),
	)
	writer.WriteString(line)
	writer.Flush()

	log.Printf("Metrics for %s written to CSV", client.name)
}

func logEvent(message string) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s\n", timestamp, message)
}

func main() {
	port := "9000"
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()
	logEvent(fmt.Sprintf("TCP chat server started on :%s", port))

	server := NewServer()
	go server.Run()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn, server)
	}
}
