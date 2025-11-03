package models

import (
	"time"
)

// UserStatus represents a user's online status
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusAway    UserStatus = "away"
	UserStatusBusy    UserStatus = "busy"
	UserStatusOffline UserStatus = "offline"
)

// User represents a user in the system
type User struct {
	ID            string     `json:"id" db:"id"`
	DisplayName   string     `json:"display_name" db:"display_name"`
	PublicKey     []byte     `json:"public_key" db:"public_key"`
	Fingerprint   string     `json:"fingerprint" db:"fingerprint"`
	Status        UserStatus `json:"status" db:"status"`
	StatusMessage string     `json:"status_message" db:"status_message"`
	LastSeen      time.Time  `json:"last_seen" db:"last_seen"`
	Verified      bool       `json:"verified" db:"verified"`
	Blocked       bool       `json:"blocked" db:"blocked"`
	
	// Local fields
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// Contact represents a contact in the user's contact list
type Contact struct {
	UserID      string    `json:"user_id" db:"user_id"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Nickname    string    `json:"nickname" db:"nickname"`
	PublicKey   []byte    `json:"public_key" db:"public_key"`
	Fingerprint string    `json:"fingerprint" db:"fingerprint"`
	Verified    bool      `json:"verified" db:"verified"`
	Blocked     bool      `json:"blocked" db:"blocked"`
	Favorite    bool      `json:"favorite" db:"favorite"`
	Notes       string    `json:"notes" db:"notes"`
	
	// Cached status information
	Status        UserStatus `json:"status" db:"status"`
	StatusMessage string     `json:"status_message" db:"status_message"`
	LastSeen      time.Time  `json:"last_seen" db:"last_seen"`
	
	// Local fields
	AddedAt   time.Time `json:"-" db:"added_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// Identity represents cryptographic identity information
type Identity struct {
	UserID         string    `json:"user_id" db:"user_id"`
	IdentityKey    []byte    `json:"identity_key" db:"identity_key"`
	SignedPreKey   []byte    `json:"signed_pre_key" db:"signed_pre_key"`
	PreKeyID       uint32    `json:"pre_key_id" db:"pre_key_id"`
	PreKeySignature []byte   `json:"pre_key_signature" db:"pre_key_signature"`
	OneTimeKeys    [][]byte  `json:"one_time_keys" db:"one_time_keys"`
	Fingerprint    string    `json:"fingerprint" db:"fingerprint"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
}

// Session represents a cryptographic session with another user
type Session struct {
	ID              string    `json:"id" db:"id"`
	LocalUserID     string    `json:"local_user_id" db:"local_user_id"`
	RemoteUserID    string    `json:"remote_user_id" db:"remote_user_id"`
	SessionState    []byte    `json:"session_state" db:"session_state"`
	RootKey         []byte    `json:"root_key" db:"root_key"`
	ChainKey        []byte    `json:"chain_key" db:"chain_key"`
	MessageNumber   uint32    `json:"message_number" db:"message_number"`
	PreviousCounter uint32    `json:"previous_counter" db:"previous_counter"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	LastUsed  time.Time `json:"last_used" db:"last_used"`
}

// NewUser creates a new user with default values
func NewUser(id, displayName string, publicKey []byte) *User {
	now := time.Now()
	return &User{
		ID:          id,
		DisplayName: displayName,
		PublicKey:   publicKey,
		Fingerprint: generateFingerprint(publicKey),
		Status:      UserStatusOffline,
		LastSeen:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewContact creates a new contact from a user
func NewContact(user *User) *Contact {
	now := time.Now()
	return &Contact{
		UserID:        user.ID,
		DisplayName:   user.DisplayName,
		PublicKey:     user.PublicKey,
		Fingerprint:   user.Fingerprint,
		Status:        user.Status,
		StatusMessage: user.StatusMessage,
		LastSeen:      user.LastSeen,
		AddedAt:       now,
		UpdatedAt:     now,
	}
}

// IsOnline returns true if the user is currently online
func (u *User) IsOnline() bool {
	return u.Status == UserStatusOnline
}

// IsActive returns true if the user has been seen recently
func (u *User) IsActive() bool {
	return time.Since(u.LastSeen) < 5*time.Minute
}

// GetDisplayName returns the display name or nickname for a contact
func (c *Contact) GetDisplayName() string {
	if c.Nickname != "" {
		return c.Nickname
	}
	return c.DisplayName
}

// IsOnline returns true if the contact is currently online
func (c *Contact) IsOnline() bool {
	return c.Status == UserStatusOnline
}

// IsActive returns true if the contact has been seen recently
func (c *Contact) IsActive() bool {
	return time.Since(c.LastSeen) < 5*time.Minute
}

// GetSafetyNumber generates a human-readable safety number for verification
func (c *Contact) GetSafetyNumber() string {
	// In a real implementation, this would generate a proper safety number
	// based on both users' identity keys
	return formatSafetyNumber(c.Fingerprint)
}

// generateFingerprint generates a fingerprint from a public key
func generateFingerprint(publicKey []byte) string {
	// In a real implementation, this would use proper cryptographic hashing
	// For now, using a simple hash representation
	if len(publicKey) < 8 {
		return "invalid"
	}
	
	// Simple hex representation of first 8 bytes
	result := ""
	for i := 0; i < 8 && i < len(publicKey); i++ {
		result += string(rune('a' + (publicKey[i] % 26)))
	}
	return result
}

// formatSafetyNumber formats a fingerprint as a human-readable safety number
func formatSafetyNumber(fingerprint string) string {
	// Convert fingerprint to 60-digit safety number (12 groups of 5 digits)
	// This is a simplified version - real implementation would use proper algorithm
	safetyNumber := ""
	for i := 0; i < 60; i++ {
		if i > 0 && i%5 == 0 {
			safetyNumber += " "
		}
		// Simple conversion from fingerprint to digits
		digit := (int(fingerprint[i%len(fingerprint)]) % 10)
		safetyNumber += string(rune('0' + digit))
	}
	return safetyNumber
}
