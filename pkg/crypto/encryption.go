package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"crypto/sha256"
	"io"
)

// EncryptedMessage represents an encrypted message
type EncryptedMessage struct {
	Ciphertext []byte
	Nonce      []byte
	Tag        []byte
}

// SessionKeys represents the keys for a messaging session
type SessionKeys struct {
	RootKey     []byte
	ChainKey    []byte
	MessageKey  []byte
	Counter     uint32
}

// DoubleRatchet implements a simplified version of the Double Ratchet algorithm
type DoubleRatchet struct {
	RootKey         []byte
	SendingChain    *ChainState
	ReceivingChain  *ChainState
	DHSelf          KeyPair
	DHRemote        []byte
	MessageNumber   uint32
	PreviousCounter uint32
}

// ChainState represents the state of a message chain
type ChainState struct {
	ChainKey      []byte
	MessageNumber uint32
}

// NewDoubleRatchet initializes a new Double Ratchet session
func NewDoubleRatchet(sharedSecret []byte, remotePublicKey []byte) (*DoubleRatchet, error) {
	// Generate initial DH key pair
	dhPrivate := make([]byte, 32)
	if _, err := rand.Read(dhPrivate); err != nil {
		return nil, fmt.Errorf("failed to generate DH private key: %w", err)
	}

	dhPublic, err := curve25519.X25519(dhPrivate, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate DH public key: %w", err)
	}

	dhSelf := KeyPair{
		PublicKey:  dhPublic,
		PrivateKey: dhPrivate,
	}

	// Initialize root key from shared secret
	rootKey := deriveRootKey(sharedSecret)

	// Initialize sending chain
	sendingChainKey, err := deriveChainKey(rootKey, dhSelf.PublicKey, remotePublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive sending chain key: %w", err)
	}

	return &DoubleRatchet{
		RootKey:      rootKey,
		DHSelf:       dhSelf,
		DHRemote:     remotePublicKey,
		SendingChain: &ChainState{
			ChainKey:      sendingChainKey,
			MessageNumber: 0,
		},
		ReceivingChain: &ChainState{
			ChainKey:      make([]byte, 32), // Will be derived when receiving
			MessageNumber: 0,
		},
		MessageNumber:   0,
		PreviousCounter: 0,
	}, nil
}

// Encrypt encrypts a message using the current session state
func (dr *DoubleRatchet) Encrypt(plaintext []byte) (*EncryptedMessage, error) {
	// Derive message key from chain key
	messageKey := deriveMessageKey(dr.SendingChain.ChainKey, dr.SendingChain.MessageNumber)
	
	// Advance chain key
	dr.SendingChain.ChainKey = advanceChainKey(dr.SendingChain.ChainKey)
	dr.SendingChain.MessageNumber++

	// Encrypt the message
	return encryptWithKey(plaintext, messageKey)
}

// Decrypt decrypts a message using the current session state
func (dr *DoubleRatchet) Decrypt(encrypted *EncryptedMessage, messageNumber uint32) ([]byte, error) {
	// For simplicity, we'll use the current receiving chain key
	// In a full implementation, we'd need to handle out-of-order messages
	messageKey := deriveMessageKey(dr.ReceivingChain.ChainKey, messageNumber)
	
	// Decrypt the message
	plaintext, err := decryptWithKey(encrypted, messageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	// Advance receiving chain
	dr.ReceivingChain.ChainKey = advanceChainKey(dr.ReceivingChain.ChainKey)
	dr.ReceivingChain.MessageNumber++

	return plaintext, nil
}

// PerformDHRatchet performs a Diffie-Hellman ratchet step
func (dr *DoubleRatchet) PerformDHRatchet(remotePublicKey []byte) error {
	// Generate new DH key pair
	newPrivate := make([]byte, 32)
	if _, err := rand.Read(newPrivate); err != nil {
		return fmt.Errorf("failed to generate new DH private key: %w", err)
	}

	newPublic, err := curve25519.X25519(newPrivate, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("failed to generate new DH public key: %w", err)
	}

	// Perform DH exchange
	sharedSecret, err := curve25519.X25519(newPrivate, remotePublicKey)
	if err != nil {
		return fmt.Errorf("failed to perform DH exchange: %w", err)
	}

	// Derive new root key and chain key
	newRootKey, newChainKey := deriveRootAndChainKeys(dr.RootKey, sharedSecret)

	// Update state
	dr.RootKey = newRootKey
	dr.DHSelf = KeyPair{
		PublicKey:  newPublic,
		PrivateKey: newPrivate,
	}
	dr.DHRemote = remotePublicKey
	dr.SendingChain = &ChainState{
		ChainKey:      newChainKey,
		MessageNumber: 0,
	}
	dr.PreviousCounter = dr.MessageNumber
	dr.MessageNumber = 0

	return nil
}

// encryptWithKey encrypts data using ChaCha20-Poly1305
func encryptWithKey(plaintext, key []byte) (*EncryptedMessage, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD cipher: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	return &EncryptedMessage{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// decryptWithKey decrypts data using ChaCha20-Poly1305
func decryptWithKey(encrypted *EncryptedMessage, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD cipher: %w", err)
	}

	// Decrypt
	plaintext, err := aead.Open(nil, encrypted.Nonce, encrypted.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// deriveRootKey derives the initial root key from shared secret
func deriveRootKey(sharedSecret []byte) []byte {
	hash := sha256.Sum256(sharedSecret)
	return hash[:]
}

// deriveChainKey derives a chain key using HKDF
func deriveChainKey(rootKey, localPublic, remotePublic []byte) ([]byte, error) {
	// Combine public keys
	combined := append(localPublic, remotePublic...)
	
	// Use HKDF to derive chain key
	hkdf := hkdf.New(sha256.New, rootKey, nil, combined)
	chainKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdf, chainKey); err != nil {
		return nil, fmt.Errorf("failed to derive chain key: %w", err)
	}
	
	return chainKey, nil
}

// deriveMessageKey derives a message key from chain key and message number
func deriveMessageKey(chainKey []byte, messageNumber uint32) []byte {
	// Simple derivation: hash(chainKey || messageNumber)
	h := sha256.New()
	h.Write(chainKey)
	h.Write([]byte{byte(messageNumber), byte(messageNumber >> 8), byte(messageNumber >> 16), byte(messageNumber >> 24)})
	hash := h.Sum(nil)
	return hash[:32] // Use first 32 bytes for ChaCha20 key
}

// advanceChainKey advances the chain key using a hash function
func advanceChainKey(chainKey []byte) []byte {
	h := sha256.New()
	h.Write(chainKey)
	h.Write([]byte{0x01}) // Chain key advancement constant
	return h.Sum(nil)[:32]
}

// deriveRootAndChainKeys derives new root and chain keys from DH output
func deriveRootAndChainKeys(rootKey, dhOutput []byte) ([]byte, []byte) {
	// Use HKDF to derive both keys
	hkdf := hkdf.New(sha256.New, dhOutput, rootKey, []byte("SecureChat-RootChain"))
	
	newRootKey := make([]byte, 32)
	newChainKey := make([]byte, 32)
	
	io.ReadFull(hkdf, newRootKey)
	io.ReadFull(hkdf, newChainKey)
	
	return newRootKey, newChainKey
}

// SimpleEncrypt provides a simple encryption interface for basic use cases
func SimpleEncrypt(plaintext []byte, key []byte) (*EncryptedMessage, error) {
	return encryptWithKey(plaintext, key)
}

// SimpleDecrypt provides a simple decryption interface for basic use cases
func SimpleDecrypt(encrypted *EncryptedMessage, key []byte) ([]byte, error) {
	return decryptWithKey(encrypted, key)
}

// GenerateSharedSecret performs X25519 key exchange
func GenerateSharedSecret(privateKey, publicKey []byte) ([]byte, error) {
	return curve25519.X25519(privateKey, publicKey)
}
