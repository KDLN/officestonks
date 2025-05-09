package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// For simplicity, we're defining constants here.
// In a production app, these might be configuration values.
const (
	// Argon2id parameters
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLength   = 16
)

// HashPassword creates a new password hash using argon2id
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Hash the password with argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	// Format the hash string as argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	
	fullHash := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		encodedSalt,
		encodedHash,
	)

	return fullHash, nil
}

// VerifyPassword compares a password with a hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the hash string
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, errors.New("invalid hash format")
	}

	// Extract parameters from the hash
	var version int
	var memory, time, threads uint32
	fmt.Sscanf(vals[2], "v=%d", &version)
	fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	// Decode hash
	decodedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	// Compute hash of the password with same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		uint8(threads),
		uint32(len(decodedHash)),
	)

	// Constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(decodedHash, computedHash) == 1, nil
}