package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/shelemiah/secure_chat/internal/models"
)

// Storage provides persistent storage for SecureChat
type Storage struct {
	db       *badger.DB
	dataDir  string
	userID   string
}

// StorageOptions contains options for storage initialization
type StorageOptions struct {
	DataDir string
	UserID  string
}

// NewStorage creates a new storage instance
func NewStorage(opts StorageOptions) (*Storage, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(opts.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open BadgerDB
	dbPath := filepath.Join(opts.DataDir, "securechat.db")
	dbOpts := badger.DefaultOptions(dbPath).
		WithLogger(nil). // Disable logging for now
		WithSyncWrites(true)

	db, err := badger.Open(dbOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{
		db:      db,
		dataDir: opts.DataDir,
		userID:  opts.UserID,
	}

	return storage, nil
}

// Close closes the storage
func (s *Storage) Close() error {
	return s.db.Close()
}

// Message storage methods

// SaveMessage saves a message to storage
func (s *Storage) SaveMessage(msg *models.Message) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.messageKey(msg.ChatID, msg.ID)
		
		// Set timestamps
		now := time.Now()
		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = now
		}
		msg.UpdatedAt = now

		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		return txn.Set(key, data)
	})
}

// GetMessage retrieves a message by ID
func (s *Storage) GetMessage(chatID, messageID string) (*models.Message, error) {
	var msg models.Message

	err := s.db.View(func(txn *badger.Txn) error {
		key := s.messageKey(chatID, messageID)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &msg)
		})
	})

	if err != nil {
		return nil, err
	}

	return &msg, nil
}

// GetMessages retrieves messages for a chat
func (s *Storage) GetMessages(chatID string, limit int, offset int) ([]*models.Message, error) {
	var messages []*models.Message

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := s.messagePrefix(chatID)
		count := 0
		skipped := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix) && count < limit; it.Next() {
			if skipped < offset {
				skipped++
				continue
			}

			item := it.Item()
			err := item.Value(func(val []byte) error {
				var msg models.Message
				if err := json.Unmarshal(val, &msg); err != nil {
					return err
				}
				messages = append(messages, &msg)
				return nil
			})
			if err != nil {
				return err
			}

			count++
		}

		return nil
	})

	return messages, err
}

// DeleteMessage deletes a message
func (s *Storage) DeleteMessage(chatID, messageID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.messageKey(chatID, messageID)
		return txn.Delete(key)
	})
}

// Contact storage methods

// SaveContact saves a contact to storage
func (s *Storage) SaveContact(contact *models.Contact) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.contactKey(contact.UserID)
		
		// Set timestamps
		now := time.Now()
		if contact.AddedAt.IsZero() {
			contact.AddedAt = now
		}
		contact.UpdatedAt = now

		data, err := json.Marshal(contact)
		if err != nil {
			return fmt.Errorf("failed to marshal contact: %w", err)
		}

		return txn.Set(key, data)
	})
}

// GetContact retrieves a contact by user ID
func (s *Storage) GetContact(userID string) (*models.Contact, error) {
	var contact models.Contact

	err := s.db.View(func(txn *badger.Txn) error {
		key := s.contactKey(userID)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &contact)
		})
	})

	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// GetAllContacts retrieves all contacts
func (s *Storage) GetAllContacts() ([]*models.Contact, error) {
	var contacts []*models.Contact

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("contacts/")

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var contact models.Contact
				if err := json.Unmarshal(val, &contact); err != nil {
					return err
				}
				contacts = append(contacts, &contact)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return contacts, err
}

// DeleteContact deletes a contact
func (s *Storage) DeleteContact(userID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.contactKey(userID)
		return txn.Delete(key)
	})
}

// Session storage methods

// SaveSession saves a cryptographic session
func (s *Storage) SaveSession(session *models.Session) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.sessionKey(session.RemoteUserID)
		
		// Set timestamps
		now := time.Now()
		if session.CreatedAt.IsZero() {
			session.CreatedAt = now
		}
		session.UpdatedAt = now
		session.LastUsed = now

		data, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("failed to marshal session: %w", err)
		}

		return txn.Set(key, data)
	})
}

// GetSession retrieves a session by remote user ID
func (s *Storage) GetSession(remoteUserID string) (*models.Session, error) {
	var session models.Session

	err := s.db.View(func(txn *badger.Txn) error {
		key := s.sessionKey(remoteUserID)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &session)
		})
	})

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session
func (s *Storage) DeleteSession(remoteUserID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.sessionKey(remoteUserID)
		return txn.Delete(key)
	})
}

// Identity storage methods

// SaveIdentity saves an identity
func (s *Storage) SaveIdentity(identity *models.Identity) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.identityKey(identity.UserID)

		data, err := json.Marshal(identity)
		if err != nil {
			return fmt.Errorf("failed to marshal identity: %w", err)
		}

		return txn.Set(key, data)
	})
}

// GetIdentity retrieves an identity by user ID
func (s *Storage) GetIdentity(userID string) (*models.Identity, error) {
	var identity models.Identity

	err := s.db.View(func(txn *badger.Txn) error {
		key := s.identityKey(userID)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &identity)
		})
	})

	if err != nil {
		return nil, err
	}

	return &identity, nil
}

// Configuration storage methods

// SaveConfig saves a configuration value
func (s *Storage) SaveConfig(key string, value interface{}) error {
	return s.db.Update(func(txn *badger.Txn) error {
		dbKey := s.configKey(key)

		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal config value: %w", err)
		}

		return txn.Set(dbKey, data)
	})
}

// GetConfig retrieves a configuration value
func (s *Storage) GetConfig(key string, dest interface{}) error {
	return s.db.View(func(txn *badger.Txn) error {
		dbKey := s.configKey(key)
		item, err := txn.Get(dbKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, dest)
		})
	})
}

// Cleanup methods

// CleanupExpiredMessages removes messages older than the retention period
func (s *Storage) CleanupExpiredMessages(retentionDays int) error {
	if retentionDays <= 0 {
		return nil // No cleanup if retention is disabled
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	return s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("messages/")
		var keysToDelete [][]byte

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var msg models.Message
				if err := json.Unmarshal(val, &msg); err != nil {
					return err
				}

				if msg.Timestamp.Before(cutoff) {
					keysToDelete = append(keysToDelete, item.KeyCopy(nil))
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		// Delete expired messages
		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})
}

// Key generation methods

func (s *Storage) messageKey(chatID, messageID string) []byte {
	return []byte(fmt.Sprintf("messages/%s/%s", chatID, messageID))
}

func (s *Storage) messagePrefix(chatID string) []byte {
	return []byte(fmt.Sprintf("messages/%s/", chatID))
}

func (s *Storage) contactKey(userID string) []byte {
	return []byte(fmt.Sprintf("contacts/%s", userID))
}

func (s *Storage) sessionKey(remoteUserID string) []byte {
	return []byte(fmt.Sprintf("sessions/%s/%s", s.userID, remoteUserID))
}

func (s *Storage) identityKey(userID string) []byte {
	return []byte(fmt.Sprintf("identities/%s", userID))
}

func (s *Storage) configKey(key string) []byte {
	return []byte(fmt.Sprintf("config/%s", key))
}
