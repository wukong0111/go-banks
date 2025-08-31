package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

// GenerateSecureSecret generates a cryptographically secure secret using crypto/rand.
// Returns a base64-encoded string of the specified byte length.
func GenerateSecureSecret(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("byte length must be positive, got %d", byteLength)
	}

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// EncryptSecret encrypts a new secret using the old secret as the key.
// Returns base64-encoded encrypted data.
func EncryptSecret(newSecret, oldSecret string) (string, error) {
	// Create a key from the old secret using SHA256
	key := sha256.Sum256([]byte(oldSecret))

	// Create cipher block
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the new secret
	ciphertext := gcm.Seal(nonce, nonce, []byte(newSecret), nil)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts an encrypted secret using the key.
func DecryptSecret(encryptedSecret, key string) (string, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create key hash
	keyHash := sha256.Sum256([]byte(key))

	// Create cipher block
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and encrypted data
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
