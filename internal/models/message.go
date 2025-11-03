package models

import (
	"time"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeChat     MessageType = "chat"
	MessageTypeSystem   MessageType = "system"
	MessageTypePresence MessageType = "presence"
	MessageTypeTyping   MessageType = "typing"
	MessageTypeAck      MessageType = "ack"
	MessageTypeError    MessageType = "error"
)

// Message represents a chat message
type Message struct {
	ID        string      `json:"id" db:"id"`
	Type      MessageType `json:"type" db:"type"`
	From      string      `json:"from" db:"from_user"`
	To        string      `json:"to" db:"to_user"`
	ChatID    string      `json:"chat_id" db:"chat_id"`
	Content   string      `json:"content" db:"content"`
	Timestamp time.Time   `json:"timestamp" db:"timestamp"`
	Encrypted bool        `json:"encrypted" db:"encrypted"`
	Signature string      `json:"signature,omitempty" db:"signature"`
	Metadata  *Metadata   `json:"metadata,omitempty" db:"metadata"`
	
	// Local fields (not transmitted)
	Status    MessageStatus `json:"-" db:"status"`
	CreatedAt time.Time     `json:"-" db:"created_at"`
	UpdatedAt time.Time     `json:"-" db:"updated_at"`
}

// MessageStatus represents the delivery status of a message
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// Metadata contains additional message information
type Metadata struct {
	ReplyTo   string    `json:"reply_to,omitempty"`
	ThreadID  string    `json:"thread_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	EditedAt  time.Time `json:"edited_at,omitempty"`
	
	// File attachment metadata
	Attachment *Attachment `json:"attachment,omitempty"`
	
	// Rich content metadata
	Formatting string                 `json:"formatting,omitempty"`
	Entities   []Entity              `json:"entities,omitempty"`
	Preview    *LinkPreview          `json:"preview,omitempty"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
	URL      string `json:"url,omitempty"`
}

// Entity represents a rich text entity (mention, link, etc.)
type Entity struct {
	Type   string `json:"type"`   // "mention", "link", "bold", "italic", "code"
	Offset int    `json:"offset"` // Character offset in text
	Length int    `json:"length"` // Length of entity
	Data   string `json:"data,omitempty"` // Additional data (URL, user ID, etc.)
}

// LinkPreview represents a preview of a shared link
type LinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
}

// NewMessage creates a new message with default values
func NewMessage(msgType MessageType, from, to, content string) *Message {
	now := time.Now()
	return &Message{
		ID:        generateMessageID(),
		Type:      msgType,
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: now,
		Status:    MessageStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsFromUser returns true if the message is from the specified user
func (m *Message) IsFromUser(userID string) bool {
	return m.From == userID
}

// IsExpired returns true if the message has expired
func (m *Message) IsExpired() bool {
	if m.Metadata == nil || m.Metadata.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(m.Metadata.ExpiresAt)
}

// IsEdited returns true if the message has been edited
func (m *Message) IsEdited() bool {
	return m.Metadata != nil && !m.Metadata.EditedAt.IsZero()
}

// HasAttachment returns true if the message has a file attachment
func (m *Message) HasAttachment() bool {
	return m.Metadata != nil && m.Metadata.Attachment != nil
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	// In a real implementation, this would generate a proper unique ID
	// For now, using timestamp + random component
	return time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
