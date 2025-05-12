# SystemsProgFINALPROJ
Chat Server Performance Comparison: TCP vs UDP
This project involves the implementation of two types of chat servers: a TCP server and a UDP server, both of which allow users to connect, chat with one another, and track various metrics like message latency, packet loss, and throughput.

The primary objective of this project is to evaluate the performance and reliability of TCP versus UDP in handling chat communication. By testing both server types under various conditions, this project aims to highlight the strengths and weaknesses of each protocol, providing insights into which is more suited for different networking environments.

TCP Server
The TCP server is designed to facilitate reliable, connection-oriented communication between clients. TCP (Transmission Control Protocol) ensures that data is sent in order, and guarantees that all messages are delivered without loss. The server implements several features to manage client connections and enhance the user experience.

Key Features of the TCP Server:
Connection Management:

Each client establishes a persistent, reliable connection to the server.

The server allows multiple clients to connect simultaneously.

Connections are handled asynchronously, allowing for concurrent messaging between clients.

Message Handling:

Clients can send and receive messages to and from other connected clients.

Messages are logged and stored in a file for each client, providing a record of chat history.

User Commands:

/time: Displays the current server time.

/date: Displays the current server date.

/joke: Tells a simple joke.

/ping: Tests and displays the latency between the client and the server.

/clients: Displays the number of clients currently connected to the server.

/help: Provides a list of available commands.

/echo [message]: Echos back the user’s message.

bye or /quit: Disconnects the user from the server.

Metrics Tracking:

Latency: Measures the round-trip time between a client’s request and the server’s response.

Message Throughput: Tracks the number of messages sent and received by each client per second.

Packet Loss: Measures the percentage of sent messages that were not acknowledged by the client.

Session Duration: Records how long each client stays connected to the server.

Inactivity Timeout:

If a client remains inactive for a specified period, the server will automatically disconnect the client to preserve resources.

Logging:

A log file is created for each client that records all interactions and server responses, which can be reviewed later for debugging or analysis.

UDP Server
The UDP server is designed for lightweight, connectionless communication, allowing clients to send and receive messages with minimal overhead. UDP (User Datagram Protocol) does not guarantee message delivery, order, or integrity, which makes it faster but less reliable than TCP.

Key Features of the UDP Server:
Connectionless Communication:

UDP does not require clients to establish a persistent connection with the server, making it more suitable for scenarios where speed is prioritized over reliability.

Message Handling:

Similar to the TCP server, users can send and receive messages, but without any guarantee of delivery.

Latency and Packet Loss Metrics:

The server records the latency of each message (round-trip time).

Tracks packet loss, which is important when using UDP as it does not ensure message delivery.

Chat Functionality:

Users can chat with each other in a similar way to the TCP server.

Performance Comparison
The project’s goal is to evaluate both server types and compare the following factors:

Reliability: How well each protocol ensures that messages are delivered to the client.

Performance: How each server handles multiple concurrent clients and message throughput.

Latency: The time it takes for a message to travel from the client to the server and back.

Packet Loss: The percentage of messages that are lost when using UDP compared to TCP’s reliability.

Scalability: The ability of each server to handle more users without significant degradation in performance.

By analyzing these factors, the project will determine which protocol (TCP or UDP) is more suitable for chat-based communication applications, based on different use cases and requirements.

