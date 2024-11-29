package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	keyFile     = ".encryption_key"
	keySize     = 32 // AES-256
	nonceSize   = 12
	prefix      = "ENC["
	suffix      = "]"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Manager handles encryption and decryption operations
type Manager struct {
	key []byte
}

// NewManager creates a new encryption manager
func NewManager() (*Manager, error) {
	key, err := loadOrGenerateKey()
	if err != nil {
		return nil, err
	}
	return &Manager{key: key}, nil
}

// loadOrGenerateKey loads the encryption key from file or generates a new one
func loadOrGenerateKey() ([]byte, error) {
	// Try to load existing key
	key, err := os.ReadFile(keyFile)
	if err == nil {
		if len(key) != keySize {
			return nil, ErrInvalidKey
		}
		log.Debug().Msg("Loaded existing encryption key")
		return key, nil
	}

	// Generate new key if file doesn't exist
	if os.IsNotExist(err) {
		key := make([]byte, keySize)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, err
		}

		// Save the key
		if err := os.WriteFile(keyFile, key, 0600); err != nil {
			return nil, err
		}

		log.Info().Msg("Generated new encryption key")
		return key, nil
	}

	return nil, err
}

// Encrypt encrypts the plaintext and returns a base64-encoded string
func (m *Manager) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create cipher
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode with base64 and add prefix/suffix
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return prefix + encoded + suffix, nil
}

// Decrypt decrypts the ciphertext and returns the plaintext
func (m *Manager) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Check if the string is encrypted
	if !IsEncrypted(ciphertext) {
		return ciphertext, nil
	}

	// Remove prefix and suffix
	encoded := strings.TrimPrefix(strings.TrimSuffix(ciphertext, suffix), prefix)

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Ensure the ciphertext is long enough
	if len(decoded) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce := decoded[:nonceSize]
	ciphertextBytes := decoded[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string is encrypted
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, prefix) && strings.HasSuffix(s, suffix)
}
