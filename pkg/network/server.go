package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server represents a relay server for SecureChat
type Server struct {
	// Server configuration
	addr     string
	port     int
	
	// WebSocket upgrader
	upgrader websocket.Upgrader
	
	// Client management
	clients    map[string]*ServerClient
	clientsMux sync.RWMutex
	
	// Message routing
	messageQueue chan *RoutedMessage
	
	// Server control
	ctx    context.Context
	cancel context.CancelFunc
	server *http.Server
	
	// Statistics
	stats ServerStats
}

// ServerClient represents a connected client
type ServerClient struct {
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Send     chan *Message
	Server   *Server
	LastSeen time.Time
}

// RoutedMessage represents a message to be routed
type RoutedMessage struct {
	From    string
	To      string
	Message *Message
}

// ServerStats contains server statistics
type ServerStats struct {
	ConnectedClients int
	MessagesRouted   int64
	Uptime          time.Time
}

// ServerOptions contains options for creating a server
type ServerOptions struct {
	Addr string
	Port int
}

// NewServer creates a new relay server
func NewServer(opts ServerOptions) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	
	if opts.Addr == "" {
		opts.Addr = "0.0.0.0"
	}
	if opts.Port == 0 {
		opts.Port = 8080
	}
	
	return &Server{
		addr: opts.Addr,
		port: opts.Port,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:      make(map[string]*ServerClient),
		messageQueue: make(chan *RoutedMessage, 1000),
		ctx:          ctx,
		cancel:       cancel,
		stats: ServerStats{
			Uptime: time.Now(),
		},
	}
}

// Start starts the relay server
func (s *Server) Start() error {
	// Start message router
	go s.messageRouter()
	
	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/stats", s.handleStats)
	
	// Create HTTP server
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.addr, s.port),
		Handler: mux,
	}
	
	log.Printf("Starting SecureChat relay server on %s:%d", s.addr, s.port)
	
	// Start server
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}
	
	return nil
}

// Stop stops the relay server
func (s *Server) Stop() error {
	log.Printf("Stopping SecureChat relay server")
	
	s.cancel()
	
	// Close all client connections
	s.clientsMux.Lock()
	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.clientsMux.Unlock()
	
	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return s.server.Shutdown(ctx)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	
	// Create client
	client := &ServerClient{
		ID:       generateClientID(),
		Conn:     conn,
		Send:     make(chan *Message, 256),
		Server:   s,
		LastSeen: time.Now(),
	}
	
	log.Printf("New client connected: %s", client.ID)
	
	// Start client handlers
	go client.readMessages()
	go client.writeMessages()
	
	// Wait for client hello to get user ID
	// For now, we'll add the client immediately
	s.addClient(client)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(s.stats.Uptime).Seconds(),
	}
	
	json.NewEncoder(w).Encode(health)
}

// handleStats handles statistics requests
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	s.clientsMux.RLock()
	s.stats.ConnectedClients = len(s.clients)
	s.clientsMux.RUnlock()
	
	json.NewEncoder(w).Encode(s.stats)
}

// addClient adds a client to the server
func (s *Server) addClient(client *ServerClient) {
	s.clientsMux.Lock()
	s.clients[client.ID] = client
	s.clientsMux.Unlock()
	
	log.Printf("Client added: %s (total: %d)", client.ID, len(s.clients))
}

// removeClient removes a client from the server
func (s *Server) removeClient(client *ServerClient) {
	s.clientsMux.Lock()
	delete(s.clients, client.ID)
	s.clientsMux.Unlock()
	
	close(client.Send)
	log.Printf("Client removed: %s (total: %d)", client.ID, len(s.clients))
}

// findClientByUserID finds a client by user ID
func (s *Server) findClientByUserID(userID string) *ServerClient {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	
	for _, client := range s.clients {
		if client.UserID == userID {
			return client
		}
	}
	
	return nil
}

// messageRouter routes messages between clients
func (s *Server) messageRouter() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case routedMsg := <-s.messageQueue:
			s.routeMessage(routedMsg)
		}
	}
}

// routeMessage routes a message to its destination
func (s *Server) routeMessage(routedMsg *RoutedMessage) {
	// Find destination client
	destClient := s.findClientByUserID(routedMsg.To)
	if destClient == nil {
		log.Printf("Destination client not found: %s", routedMsg.To)
		// TODO: Store message for offline delivery
		return
	}
	
	// Send message to destination
	select {
	case destClient.Send <- routedMsg.Message:
		s.stats.MessagesRouted++
		log.Printf("Message routed from %s to %s", routedMsg.From, routedMsg.To)
	default:
		log.Printf("Failed to route message: destination client queue full")
	}
}

// ServerClient methods

// readMessages reads messages from the client connection
func (c *ServerClient) readMessages() {
	defer func() {
		c.Server.removeClient(c)
		c.Conn.Close()
	}()
	
	// Set read limits and deadline
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		// Read message
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for client %s: %v", c.ID, err)
			}
			break
		}
		
		// Parse message
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Failed to parse message from client %s: %v", c.ID, err)
			continue
		}
		
		// Update last seen
		c.LastSeen = time.Now()
		
		// Handle message based on type
		c.handleMessage(&msg)
	}
}

// writeMessages writes messages to the client connection
func (c *ServerClient) writeMessages() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// Marshal and send message
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to marshal message for client %s: %v", c.ID, err)
				continue
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to write message to client %s: %v", c.ID, err)
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles different types of messages from clients
func (c *ServerClient) handleMessage(msg *Message) {
	switch msg.Type {
	case "client_hello":
		c.handleClientHello(msg)
	case "chat":
		c.handleChatMessage(msg)
	case "presence":
		c.handlePresenceMessage(msg)
	default:
		log.Printf("Unknown message type from client %s: %s", c.ID, msg.Type)
	}
}

// handleClientHello handles client hello messages
func (c *ServerClient) handleClientHello(msg *Message) {
	// Extract user ID from message
	c.UserID = msg.From
	
	log.Printf("Client %s identified as user %s", c.ID, c.UserID)
	
	// Send server hello response
	response := &Message{
		ID:        generateMessageID(),
		Type:      "server_hello",
		From:      "server",
		To:        c.UserID,
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"version":    "1.0",
			"session_id": c.ID,
			"capabilities": []string{"message_relay", "offline_storage"},
		},
	}
	
	select {
	case c.Send <- response:
	default:
		log.Printf("Failed to send server hello to client %s", c.ID)
	}
}

// handleChatMessage handles chat messages
func (c *ServerClient) handleChatMessage(msg *Message) {
	// Validate message
	if msg.To == "" {
		log.Printf("Chat message from %s missing destination", c.UserID)
		return
	}
	
	// Queue message for routing
	routedMsg := &RoutedMessage{
		From:    c.UserID,
		To:      msg.To,
		Message: msg,
	}
	
	select {
	case c.Server.messageQueue <- routedMsg:
	default:
		log.Printf("Message queue full, dropping message from %s to %s", c.UserID, msg.To)
	}
}

// handlePresenceMessage handles presence messages
func (c *ServerClient) handlePresenceMessage(msg *Message) {
	// TODO: Implement presence handling
	log.Printf("Presence message from %s: %v", c.UserID, msg.Payload)
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d_%d", time.Now().UnixNano(), time.Now().Unix()%1000)
}
