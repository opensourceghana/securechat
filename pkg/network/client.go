package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a network client for SecureChat
type Client struct {
	// Connection management
	conn        *websocket.Conn
	connMutex   sync.RWMutex
	isConnected bool
	
	// Configuration
	serverURL   string
	userID      string
	
	// Channels
	incomingMessages chan *Message
	outgoingMessages chan *Message
	connectionEvents chan ConnectionEvent
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
	
	// Callbacks
	messageHandler     MessageHandler
	connectionHandler  ConnectionHandler
	
	// Reconnection
	reconnectAttempts int
	maxReconnectAttempts int
	reconnectDelay    time.Duration
}

// Message represents a network message
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Timestamp int64                  `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature,omitempty"`
}

// ConnectionEvent represents a connection state change
type ConnectionEvent struct {
	Type      ConnectionEventType
	Error     error
	Timestamp time.Time
}

// ConnectionEventType represents the type of connection event
type ConnectionEventType int

const (
	ConnectionEventConnected ConnectionEventType = iota
	ConnectionEventDisconnected
	ConnectionEventReconnecting
	ConnectionEventError
)

// MessageHandler is called when a message is received
type MessageHandler func(*Message) error

// ConnectionHandler is called when connection state changes
type ConnectionHandler func(ConnectionEvent)

// ClientOptions contains options for creating a client
type ClientOptions struct {
	ServerURL            string
	UserID               string
	MaxReconnectAttempts int
	ReconnectDelay       time.Duration
	MessageHandler       MessageHandler
	ConnectionHandler    ConnectionHandler
}

// NewClient creates a new network client
func NewClient(opts ClientOptions) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	
	if opts.MaxReconnectAttempts == 0 {
		opts.MaxReconnectAttempts = 10
	}
	if opts.ReconnectDelay == 0 {
		opts.ReconnectDelay = 5 * time.Second
	}
	
	return &Client{
		serverURL:            opts.ServerURL,
		userID:               opts.UserID,
		incomingMessages:     make(chan *Message, 100),
		outgoingMessages:     make(chan *Message, 100),
		connectionEvents:     make(chan ConnectionEvent, 10),
		ctx:                  ctx,
		cancel:               cancel,
		messageHandler:       opts.MessageHandler,
		connectionHandler:    opts.ConnectionHandler,
		maxReconnectAttempts: opts.MaxReconnectAttempts,
		reconnectDelay:       opts.ReconnectDelay,
	}
}

// Connect establishes a connection to the server
func (c *Client) Connect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	
	if c.isConnected {
		return fmt.Errorf("already connected")
	}
	
	// Parse server URL
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	
	// Convert HTTP(S) to WS(S)
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		// Already correct
	default:
		return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
	
	// Establish WebSocket connection
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second
	
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	
	c.conn = conn
	c.isConnected = true
	c.reconnectAttempts = 0
	
	// Start message handling goroutines
	go c.readMessages()
	go c.writeMessages()
	go c.handleEvents()
	
	// Send connection event
	c.sendConnectionEvent(ConnectionEvent{
		Type:      ConnectionEventConnected,
		Timestamp: time.Now(),
	})
	
	// Send client hello
	if err := c.sendClientHello(); err != nil {
		log.Printf("Failed to send client hello: %v", err)
	}
	
	return nil
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	
	if !c.isConnected {
		return nil
	}
	
	c.isConnected = false
	c.cancel() // Cancel context to stop goroutines
	
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	
	c.sendConnectionEvent(ConnectionEvent{
		Type:      ConnectionEventDisconnected,
		Timestamp: time.Now(),
	})
	
	return nil
}

// SendMessage sends a message to another user
func (c *Client) SendMessage(to string, content string, msgType string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}
	
	msg := &Message{
		ID:        generateMessageID(),
		Type:      msgType,
		From:      c.userID,
		To:        to,
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"content": content,
		},
	}
	
	select {
	case c.outgoingMessages <- msg:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client is shutting down")
	default:
		return fmt.Errorf("outgoing message queue is full")
	}
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.isConnected
}

// Close shuts down the client
func (c *Client) Close() error {
	return c.Disconnect()
}

// readMessages reads messages from the WebSocket connection
func (c *Client) readMessages() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in readMessages: %v", r)
		}
	}()
	
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		
		c.connMutex.RLock()
		conn := c.conn
		connected := c.isConnected
		c.connMutex.RUnlock()
		
		if !connected || conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		
		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		// Read message
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			c.handleConnectionError(err)
			return
		}
		
		// Parse message
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}
		
		// Handle message
		if c.messageHandler != nil {
			if err := c.messageHandler(&msg); err != nil {
				log.Printf("Message handler error: %v", err)
			}
		}
	}
}

// writeMessages writes messages to the WebSocket connection
func (c *Client) writeMessages() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in writeMessages: %v", r)
		}
	}()
	
	ticker := time.NewTicker(54 * time.Second) // Ping every 54 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-c.ctx.Done():
			return
			
		case msg := <-c.outgoingMessages:
			if err := c.writeMessage(msg); err != nil {
				log.Printf("Failed to write message: %v", err)
				c.handleConnectionError(err)
				return
			}
			
		case <-ticker.C:
			if err := c.writePing(); err != nil {
				log.Printf("Failed to write ping: %v", err)
				c.handleConnectionError(err)
				return
			}
		}
	}
}

// writeMessage writes a single message to the connection
func (c *Client) writeMessage(msg *Message) error {
	c.connMutex.RLock()
	conn := c.conn
	connected := c.isConnected
	c.connMutex.RUnlock()
	
	if !connected || conn == nil {
		return fmt.Errorf("not connected")
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, data)
}

// writePing writes a ping message
func (c *Client) writePing() error {
	c.connMutex.RLock()
	conn := c.conn
	connected := c.isConnected
	c.connMutex.RUnlock()
	
	if !connected || conn == nil {
		return fmt.Errorf("not connected")
	}
	
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteMessage(websocket.PingMessage, nil)
}

// handleEvents processes connection events
func (c *Client) handleEvents() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case event := <-c.connectionEvents:
			if c.connectionHandler != nil {
				c.connectionHandler(event)
			}
		}
	}
}

// handleConnectionError handles connection errors and attempts reconnection
func (c *Client) handleConnectionError(err error) {
	c.connMutex.Lock()
	c.isConnected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connMutex.Unlock()
	
	c.sendConnectionEvent(ConnectionEvent{
		Type:      ConnectionEventError,
		Error:     err,
		Timestamp: time.Now(),
	})
	
	// Attempt reconnection
	go c.attemptReconnection()
}

// attemptReconnection attempts to reconnect to the server
func (c *Client) attemptReconnection() {
	for c.reconnectAttempts < c.maxReconnectAttempts {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		
		c.reconnectAttempts++
		
		c.sendConnectionEvent(ConnectionEvent{
			Type:      ConnectionEventReconnecting,
			Timestamp: time.Now(),
		})
		
		log.Printf("Attempting reconnection %d/%d", c.reconnectAttempts, c.maxReconnectAttempts)
		
		if err := c.Connect(); err != nil {
			log.Printf("Reconnection attempt %d failed: %v", c.reconnectAttempts, err)
			
			// Wait before next attempt with exponential backoff
			delay := c.reconnectDelay * time.Duration(c.reconnectAttempts)
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}
			
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		} else {
			log.Printf("Reconnection successful")
			return
		}
	}
	
	log.Printf("Max reconnection attempts reached, giving up")
}

// sendConnectionEvent sends a connection event
func (c *Client) sendConnectionEvent(event ConnectionEvent) {
	select {
	case c.connectionEvents <- event:
	default:
		// Channel is full, drop the event
	}
}

// sendClientHello sends the initial client hello message
func (c *Client) sendClientHello() error {
	msg := &Message{
		ID:        generateMessageID(),
		Type:      "client_hello",
		From:      c.userID,
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"version":      "1.0",
			"capabilities": []string{"e2e_encryption", "file_transfer"},
		},
	}
	
	return c.writeMessage(msg)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d_%d", time.Now().UnixNano(), time.Now().Unix()%1000)
}
