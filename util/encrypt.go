package util

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters
const (
	argon2idTime    uint32 = 1
	argon2idMemory  uint32 = 64 * 1024 // 64MB
	argon2idThreads uint8  = 4
	argon2idKeyLen  uint32 = 32
	argon2idSaltLen uint32 = 16
	argon2idPrefix  string = "$argon2id$"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2idTime, argon2idMemory, argon2idThreads, argon2idKeyLen)

	saltHex := hex.EncodeToString(salt)
	keyHex := hex.EncodeToString(key)

	return fmt.Sprintf("%s%s$%s", argon2idPrefix, saltHex, keyHex), nil
}

func VerifyPassword(password string, storedHash string) (bool, error) {
	if !strings.HasPrefix(storedHash, argon2idPrefix) {
		return false, errors.New("invalid stored hash format: missing prefix")
	}

	parts := strings.Split(strings.TrimPrefix(storedHash, argon2idPrefix), "$")
	if len(parts) != 2 {
		return false, errors.New("invalid stored hash format: incorrect number of parts")
	}

	saltHex := parts[0]
	keyHex := parts[1]

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt from hex: %w", err)
	}
	if uint32(len(salt)) != argon2idSaltLen {
		return false, errors.New("invalid salt length in stored hash")
	}

	storedKey, err := hex.DecodeString(keyHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode key from hex: %w", err)
	}
	if uint32(len(storedKey)) != argon2idKeyLen {
		return false, errors.New("invalid key length in stored hash")
	}

	derivedKey := argon2.IDKey([]byte(password), salt, argon2idTime, argon2idMemory, argon2idThreads, argon2idKeyLen)

	if subtle.ConstantTimeCompare(storedKey, derivedKey) == 1 {
		return true, nil
	}
	return false, nil
}

func SHA224String(password string) string {
	hash := sha256.New224()
	hash.Write([]byte(password))
	val := hash.Sum(nil)
	str := ""
	for _, v := range val {
		str += fmt.Sprintf("%02x", v)
	}
	return str
}
