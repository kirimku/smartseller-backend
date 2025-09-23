package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// PasswordConfig contains the parameters for Argon2 password hashing
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordConfig returns a default configuration for password hashing
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// GenerateSalt generates a random salt for password hashing
func GenerateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// HashPassword hashes a password using the Argon2id algorithm
func HashPassword(password string, salt []byte, config *PasswordConfig) (string, error) {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// Base64 encode the hash for storage
	encodedHash := base64.StdEncoding.EncodeToString(hash)
	return encodedHash, nil
}

// VerifyPassword verifies if a password matches the stored hash
func VerifyPassword(password string, encodedHash string, salt []byte, config *PasswordConfig) (bool, error) {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	// Decode the stored hash
	storedHash, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode stored hash: %w", err)
	}

	// Hash the provided password
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// Compare the computed hash with the stored hash
	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1, nil
}

// EncodeSalt encodes a salt to base64 for storage
func EncodeSalt(salt []byte) string {
	return base64.StdEncoding.EncodeToString(salt)
}

// DecodeSalt decodes a base64-encoded salt
func DecodeSalt(encodedSalt string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedSalt)
}
