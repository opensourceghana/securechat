package core

import (
	"fmt"
	"log"
	"time"

	"github.com/shelemiah/secure_chat/internal/config"
	"github.com/shelemiah/secure_chat/internal/models"
	"github.com/shelemiah/secure_chat/pkg/crypto"
	"github.com/shelemiah/secure_chat/pkg/network"
	"github.com/shelemiah/secure_chat/pkg/storage"
)

// App represents the core SecureChat application
type App struct {
	config   *config.Config
	storage  *storage.Storage
	client   *network.Client
	identity *crypto.IdentityKeyPair
	
	// Message handlers
	messageHandlers []MessageHandler
	
	// State
	contacts map[string]*models.Contact
	sessions map[string]*crypto.DoubleRatchet
}

// MessageHandler handles incoming messages
type MessageHandler func(*models.Message) error

// NewApp creates a new SecureChat application
func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		config:          cfg,
		contacts:        make(map[string]*models.Contact),
		sessions:        make(map[string]*crypto.DoubleRatchet),
		messageHandlers: make([]MessageHandler, 0),
	}
	
	// Initialize storage
	if err := app.initStorage(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	
	// Initialize or load identity
	if err := app.initIdentity(); err != nil {
		return nil, fmt.Errorf("failed to initialize identity: %w", err)
	}
	
	// Initialize network client
	if err := app.initNetworkClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize network client: %w", err)
	}
	
	// Load contacts
	if err := app.loadContacts(); err != nil {
		log.Printf("Warning: failed to load contacts: %v", err)
	}
	
	return app, nil
}

// initStorage initializes the storage layer
func (a *App) initStorage() error {
	dataDir := a.config.GetDataDir()
	
	storageOpts := storage.StorageOptions{
		DataDir: dataDir,
		UserID:  a.config.User.ID,
	}
	
	var err error
	a.storage, err = storage.NewStorage(storageOpts)
	if err != nil {
		return err
	}
	
	return nil
}

// initIdentity initializes or loads the user's identity
func (a *App) initIdentity() error {
	// Try to load existing identity from storage
	if identity, err := a.storage.GetIdentity(a.config.User.ID); err == nil {
		// Convert stored identity to crypto identity
		// This is simplified - in a real implementation, we'd properly deserialize the keys
		a.identity = &crypto.IdentityKeyPair{
			Fingerprint: identity.Fingerprint,
		}
		log.Printf("Loaded existing identity for user %s", a.config.User.ID)
		return nil
	}
	
	// Generate new identity
	identity, err := crypto.GenerateIdentityKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}
	
	a.identity = identity
	
	// Store identity
	storedIdentity := &models.Identity{
		UserID:      a.config.User.ID,
		IdentityKey: identity.SigningKey.PublicKey,
		Fingerprint: identity.Fingerprint,
	}
	
	if err := a.storage.SaveIdentity(storedIdentity); err != nil {
		log.Printf("Warning: failed to save identity: %v", err)
	}
	
	log.Printf("Generated new identity for user %s", a.config.User.ID)
	return nil
}

// initNetworkClient initializes the network client
func (a *App) initNetworkClient() error {
	if len(a.config.Network.RelayServers) == 0 {
		return fmt.Errorf("no relay servers configured")
	}
	
	// Use first relay server for now
	serverURL := "ws://" + a.config.Network.RelayServers[0]
	
	clientOpts := network.ClientOptions{
		ServerURL:         serverURL,
		UserID:           a.config.User.ID,
		MessageHandler:   a.handleNetworkMessage,
		ConnectionHandler: a.handleConnectionEvent,
	}
	
	a.client = network.NewClient(clientOpts)
	
	return nil
}

// loadContacts loads contacts from storage
func (a *App) loadContacts() error {
	contacts, err := a.storage.GetAllContacts()
	if err != nil {
		return err
	}
	
	for _, contact := range contacts {
		a.contacts[contact.UserID] = contact
	}
	
	log.Printf("Loaded %d contacts", len(a.contacts))
	return nil
}

// Connect connects to the network
func (a *App) Connect() error {
	return a.client.Connect()
}

// Disconnect disconnects from the network
func (a *App) Disconnect() error {
	return a.client.Disconnect()
}

// SendMessage sends a message to another user
func (a *App) SendMessage(to, content string) error {
	// Check if we have this contact
	contact, exists := a.contacts[to]
	if !exists {
		return fmt.Errorf("contact not found: %s", to)
	}
	
	// For now, send unencrypted message
	// TODO: Implement proper encryption with Double Ratchet
	err := a.client.SendMessage(to, content, "chat")
	if err != nil {
		return err
	}
	
	// Save message to local storage
	msg := models.NewMessage(models.MessageTypeChat, a.config.User.ID, to, content)
	msg.ChatID = a.getChatID(a.config.User.ID, to)
	
	if err := a.storage.SaveMessage(msg); err != nil {
		log.Printf("Warning: failed to save sent message: %v", err)
	}
	
	// Notify handlers
	for _, handler := range a.messageHandlers {
		if err := handler(msg); err != nil {
			log.Printf("Message handler error: %v", err)
		}
	}
	
	log.Printf("Sent message to %s (%s)", contact.GetDisplayName(), to)
	return nil
}

// AddContact adds a new contact
func (a *App) AddContact(userID, displayName string) error {
	// Create contact
	contact := &models.Contact{
		UserID:      userID,
		DisplayName: displayName,
		Status:      models.UserStatusOffline,
	}
	
	// Save to storage
	if err := a.storage.SaveContact(contact); err != nil {
		return fmt.Errorf("failed to save contact: %w", err)
	}
	
	// Add to memory
	a.contacts[userID] = contact
	
	log.Printf("Added contact: %s (%s)", displayName, userID)
	return nil
}

// GetContacts returns all contacts
func (a *App) GetContacts() []*models.Contact {
	contacts := make([]*models.Contact, 0, len(a.contacts))
	for _, contact := range a.contacts {
		contacts = append(contacts, contact)
	}
	return contacts
}

// GetMessages returns messages for a chat
func (a *App) GetMessages(otherUserID string, limit int) ([]*models.Message, error) {
	chatID := a.getChatID(a.config.User.ID, otherUserID)
	return a.storage.GetMessages(chatID, limit, 0)
}

// AddMessageHandler adds a message handler
func (a *App) AddMessageHandler(handler MessageHandler) {
	a.messageHandlers = append(a.messageHandlers, handler)
}

// IsConnected returns true if connected to the network
func (a *App) IsConnected() bool {
	return a.client.IsConnected()
}

// GetUserID returns the current user's ID
func (a *App) GetUserID() string {
	return a.config.User.ID
}

// GetDisplayName returns the current user's display name
func (a *App) GetDisplayName() string {
	return a.config.User.DisplayName
}

// GetFingerprint returns the user's identity fingerprint
func (a *App) GetFingerprint() string {
	if a.identity == nil {
		return ""
	}
	return a.identity.Fingerprint
}

// Close closes the application and cleans up resources
func (a *App) Close() error {
	if a.client != nil {
		a.client.Close()
	}
	
	if a.storage != nil {
		a.storage.Close()
	}
	
	return nil
}

// handleNetworkMessage handles incoming network messages
func (a *App) handleNetworkMessage(netMsg *network.Message) error {
	// Convert network message to internal message
	msg := &models.Message{
		ID:        netMsg.ID,
		Type:      models.MessageType(netMsg.Type),
		From:      netMsg.From,
		To:        netMsg.To,
		ChatID:    a.getChatID(netMsg.From, netMsg.To),
		Timestamp: timeFromUnix(netMsg.Timestamp),
	}
	
	// Extract content from payload
	if content, ok := netMsg.Payload["content"].(string); ok {
		msg.Content = content
	}
	
	// Save message to storage
	if err := a.storage.SaveMessage(msg); err != nil {
		log.Printf("Warning: failed to save received message: %v", err)
	}
	
	// Notify handlers
	for _, handler := range a.messageHandlers {
		if err := handler(msg); err != nil {
			log.Printf("Message handler error: %v", err)
		}
	}
	
	log.Printf("Received message from %s: %s", msg.From, msg.Content)
	return nil
}

// handleConnectionEvent handles network connection events
func (a *App) handleConnectionEvent(event network.ConnectionEvent) {
	switch event.Type {
	case network.ConnectionEventConnected:
		log.Printf("Connected to relay server")
	case network.ConnectionEventDisconnected:
		log.Printf("Disconnected from relay server")
	case network.ConnectionEventReconnecting:
		log.Printf("Reconnecting to relay server...")
	case network.ConnectionEventError:
		log.Printf("Connection error: %v", event.Error)
	}
}

// getChatID generates a consistent chat ID for two users
func (a *App) getChatID(user1, user2 string) string {
	if user1 < user2 {
		return fmt.Sprintf("chat_%s_%s", user1, user2)
	}
	return fmt.Sprintf("chat_%s_%s", user2, user1)
}

// timeFromUnix converts Unix timestamp to time.Time
func timeFromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}
