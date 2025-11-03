package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
)

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// IdentityKeyPair represents a long-term identity key pair
type IdentityKeyPair struct {
	SigningKey   KeyPair // Ed25519 for signatures
	ExchangeKey  KeyPair // X25519 for key exchange
	Fingerprint  string
}

// PreKey represents a signed prekey for key exchange
type PreKey struct {
	ID        uint32
	KeyPair   KeyPair
	Signature []byte
}

// OneTimeKey represents a one-time prekey
type OneTimeKey struct {
	ID      uint32
	KeyPair KeyPair
}

// GenerateIdentityKeyPair generates a new identity key pair
func GenerateIdentityKeyPair() (*IdentityKeyPair, error) {
	// Generate Ed25519 signing key
	signingPub, signingPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key: %w", err)
	}

	// Generate X25519 exchange key
	exchangePriv := make([]byte, 32)
	if _, err := rand.Read(exchangePriv); err != nil {
		return nil, fmt.Errorf("failed to generate exchange private key: %w", err)
	}

	exchangePub, err := curve25519.X25519(exchangePriv, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate exchange public key: %w", err)
	}

	identity := &IdentityKeyPair{
		SigningKey: KeyPair{
			PublicKey:  signingPub,
			PrivateKey: signingPriv,
		},
		ExchangeKey: KeyPair{
			PublicKey:  exchangePub,
			PrivateKey: exchangePriv,
		},
	}

	identity.Fingerprint = generateFingerprint(identity)

	return identity, nil
}

// GeneratePreKey generates a new signed prekey
func GeneratePreKey(id uint32, identityKey *IdentityKeyPair) (*PreKey, error) {
	// Generate X25519 key pair
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("failed to generate prekey private key: %w", err)
	}

	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prekey public key: %w", err)
	}

	keyPair := KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}

	// Sign the public key with identity signing key
	signature := ed25519.Sign(identityKey.SigningKey.PrivateKey, publicKey)

	return &PreKey{
		ID:        id,
		KeyPair:   keyPair,
		Signature: signature,
	}, nil
}

// GenerateOneTimeKeys generates multiple one-time prekeys
func GenerateOneTimeKeys(startID uint32, count int) ([]*OneTimeKey, error) {
	keys := make([]*OneTimeKey, count)

	for i := 0; i < count; i++ {
		// Generate X25519 key pair
		privateKey := make([]byte, 32)
		if _, err := rand.Read(privateKey); err != nil {
			return nil, fmt.Errorf("failed to generate one-time key private key: %w", err)
		}

		publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
		if err != nil {
			return nil, fmt.Errorf("failed to generate one-time key public key: %w", err)
		}

		keys[i] = &OneTimeKey{
			ID: startID + uint32(i),
			KeyPair: KeyPair{
				PublicKey:  publicKey,
				PrivateKey: privateKey,
			},
		}
	}

	return keys, nil
}

// VerifyPreKey verifies a signed prekey signature
func VerifyPreKey(prekey *PreKey, identityPublicKey []byte) bool {
	return ed25519.Verify(identityPublicKey, prekey.KeyPair.PublicKey, prekey.Signature)
}

// generateFingerprint generates a human-readable fingerprint for an identity
func generateFingerprint(identity *IdentityKeyPair) string {
	// Combine both public keys
	combined := append(identity.SigningKey.PublicKey, identity.ExchangeKey.PublicKey...)
	
	// Hash the combined keys
	hash := sha256.Sum256(combined)
	
	// Convert to hex string
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for shorter fingerprint
}

// GetSafetyNumber generates a 60-digit safety number for identity verification
func GetSafetyNumber(localIdentity, remoteIdentity *IdentityKeyPair) string {
	// Combine identity keys in a deterministic order
	var combined []byte
	
	localCombined := append(localIdentity.SigningKey.PublicKey, localIdentity.ExchangeKey.PublicKey...)
	remoteCombined := append(remoteIdentity.SigningKey.PublicKey, remoteIdentity.ExchangeKey.PublicKey...)
	
	// Ensure consistent ordering
	if string(localCombined) < string(remoteCombined) {
		combined = append(localCombined, remoteCombined...)
	} else {
		combined = append(remoteCombined, localCombined...)
	}
	
	// Generate hash
	hash := sha256.Sum256(combined)
	
	// Convert to 60-digit safety number
	safetyNumber := ""
	for i := 0; i < 12; i++ { // 12 groups of 5 digits
		if i > 0 {
			safetyNumber += " "
		}
		
		// Use 5 bytes of hash to generate 5 digits
		for j := 0; j < 5; j++ {
			byteIndex := (i*5 + j) % len(hash)
			digit := hash[byteIndex] % 10
			safetyNumber += fmt.Sprintf("%d", digit)
		}
	}
	
	return safetyNumber
}

// SecureCompare performs constant-time comparison of byte slices
func SecureCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	
	return result == 0
}
