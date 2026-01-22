// auth/token_store.go
package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// TokenStore manages OAuth2 tokens
type TokenStore interface {
	Save(ctx context.Context, providerName string, token *OAuth2Token) error
	Load(ctx context.Context, providerName string) (*OAuth2Token, error)
	Delete(ctx context.Context, providerName string) error
	Close() error
}

// FileTokenStore stores tokens in encrypted files
type FileTokenStore struct {
	baseDir       string
	encryptionKey string
	mu            sync.RWMutex
}

// NewFileTokenStore creates a new file-based token store
func NewFileTokenStore(baseDir, encryptionKey string) (*FileTokenStore, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create token store directory: %w", err)
	}

	return &FileTokenStore{
		baseDir:       baseDir,
		encryptionKey: encryptionKey,
	}, nil
}

// Save saves a token to disk (encrypted)
func (s *FileTokenStore) Save(ctx context.Context, providerName string, token *OAuth2Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize token
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Encrypt data
	encrypted, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Write to file
	filename := filepath.Join(s.baseDir, providerName+".token")
	if err := os.WriteFile(filename, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// Load loads a token from disk (decrypted)
func (s *FileTokenStore) Load(ctx context.Context, providerName string) (*OAuth2Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read file
	filename := filepath.Join(s.baseDir, providerName+".token")
	encrypted, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	// Decrypt data
	data, err := s.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Deserialize token
	var token OAuth2Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// Delete deletes a token
func (s *FileTokenStore) Delete(ctx context.Context, providerName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.baseDir, providerName+".token")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	return nil
}

// Close closes the token store
func (s *FileTokenStore) Close() error {
	return nil
}

// encrypt encrypts data using AES-GCM
func (s *FileTokenStore) encrypt(data []byte) ([]byte, error) {
	// Derive key from encryption key
	key := []byte(s.encryptionKey)
	if len(key) < 32 {
		// Pad key to 32 bytes
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// decrypt decrypts data using AES-GCM
func (s *FileTokenStore) decrypt(data []byte) ([]byte, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	// Derive key
	key := []byte(s.encryptionKey)
	if len(key) < 32 {
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateKey generates a random encryption key
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// MemoryTokenStore stores tokens in memory (for testing)
type MemoryTokenStore struct {
	tokens map[string]*OAuth2Token
	mu     sync.RWMutex
}

// NewMemoryTokenStore creates a new memory token store
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		tokens: make(map[string]*OAuth2Token),
	}
}

func (s *MemoryTokenStore) Save(ctx context.Context, providerName string, token *OAuth2Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[providerName] = token
	return nil
}

func (s *MemoryTokenStore) Load(ctx context.Context, providerName string) (*OAuth2Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, ok := s.tokens[providerName]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return token, nil
}

func (s *MemoryTokenStore) Delete(ctx context.Context, providerName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, providerName)
	return nil
}

func (s *MemoryTokenStore) Close() error {
	return nil
}
