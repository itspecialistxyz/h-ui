package util

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !strings.HasPrefix(hashedPassword, argon2idPrefix) {
		t.Errorf("Hashed password should have prefix %s, got %s", argon2idPrefix, hashedPassword)
	}

	parts := strings.Split(strings.TrimPrefix(hashedPassword, argon2idPrefix), "$")
	if len(parts) != 2 {
		t.Errorf("Hashed password format is incorrect, expected 2 parts, got %d", len(parts))
	}

	// Check salt length (hex encoded, so *2)
	if len(parts[0]) != int(argon2idSaltLen*2) {
		t.Errorf("Expected salt length %d, got %d", argon2idSaltLen*2, len(parts[0]))
	}

	// Check key length (hex encoded, so *2)
	if len(parts[1]) != int(argon2idKeyLen*2) {
		t.Errorf("Expected key length %d, got %d", argon2idKeyLen*2, len(parts[1]))
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Setup: HashPassword failed: %v", err)
	}

	// Test with correct password
	match, err := VerifyPassword(password, hashedPassword)
	if err != nil {
		t.Errorf("VerifyPassword with correct password failed: %v", err)
	}
	if !match {
		t.Errorf("VerifyPassword with correct password: expected true, got false")
	}

	// Test with incorrect password
	wrongPassword := "wrongpassword"
	match, err = VerifyPassword(wrongPassword, hashedPassword)
	if err != nil {
		t.Errorf("VerifyPassword with incorrect password failed: %v", err)
	}
	if match {
		t.Errorf("VerifyPassword with incorrect password: expected false, got true")
	}
}

func TestVerifyPassword_InvalidFormat(t *testing.T) {
	password := "testpassword"

	// Test with missing prefix
	invalidHash_noPrefix := "somesalt$somekey"
	_, err := VerifyPassword(password, invalidHash_noPrefix)
	if err == nil {
		t.Errorf("VerifyPassword with no prefix: expected error, got nil")
	} else if !strings.Contains(err.Error(), "missing prefix") {
		t.Errorf("VerifyPassword with no prefix: expected error containing 'missing prefix', got %v", err)
	}

	// Test with too few parts
	invalidHash_tooFewParts := "$argon2id$somesalt"
	_, err = VerifyPassword(password, invalidHash_tooFewParts)
	if err == nil {
		t.Errorf("VerifyPassword with too few parts: expected error, got nil")
	} else if !strings.Contains(err.Error(), "incorrect number of parts") {
		t.Errorf("VerifyPassword with too few parts: expected error containing 'incorrect number of parts', got %v", err)
	}

	// Test with non-hex salt
	invalidHash_badSalt := "$argon2id$nothexsalt$somekey"
	_, err = VerifyPassword(password, invalidHash_badSalt)
	if err == nil {
		t.Errorf("VerifyPassword with non-hex salt: expected error, got nil")
	} else if !strings.Contains(err.Error(), "decode salt from hex") {
		t.Errorf("VerifyPassword with non-hex salt: expected error containing 'decode salt from hex', got %v", err)
	}
	
	// Test with incorrect salt length (too short)
	validSaltButWrongLength := "0123456789abcdef" // 8 bytes, argon2idSaltLen is 16
	invalidHash_wrongSaltLength := "$argon2id$" + validSaltButWrongLength + "$somekey"
	_, err = VerifyPassword(password, invalidHash_wrongSaltLength)
	if err == nil {
		t.Errorf("VerifyPassword with wrong salt length: expected error, got nil")
	} else if !strings.Contains(err.Error(), "invalid salt length") {
		t.Errorf("VerifyPassword with wrong salt length: expected error containing 'invalid salt length', got %v", err)
	}


	// Test with non-hex key
	validSalt := strings.Repeat("a", int(argon2idSaltLen*2))
	invalidHash_badKey := "$argon2id$" + validSalt + "$nothexkey"
	_, err = VerifyPassword(password, invalidHash_badKey)
	if err == nil {
		t.Errorf("VerifyPassword with non-hex key: expected error, got nil")
	} else if !strings.Contains(err.Error(), "decode key from hex") {
		t.Errorf("VerifyPassword with non-hex key: expected error containing 'decode key from hex', got %v", err)
	}

	// Test with incorrect key length (too short)
	validKeyButWrongLength := "0123456789abcdef" // 8 bytes, argon2idKeyLen is 32
	invalidHash_wrongKeyLength := "$argon2id$" + validSalt + "$" + validKeyButWrongLength
	_, err = VerifyPassword(password, invalidHash_wrongKeyLength)
	if err == nil {
		t.Errorf("VerifyPassword with wrong key length: expected error, got nil")
	} else if !strings.Contains(err.Error(), "invalid key length") {
		t.Errorf("VerifyPassword with wrong key length: expected error containing 'invalid key length', got %v", err)
	}
}

func TestHashPassword_SaltUniqueness(t *testing.T) {
	password := "testpassword"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("HashPassword failed during salt uniqueness test: %v, %v", err1, err2)
	}

	if hash1 == hash2 {
		t.Errorf("Generated hashes for the same password should be different due to unique salts. Got identical hashes: %s", hash1)
	}

	parts1 := strings.Split(strings.TrimPrefix(hash1, argon2idPrefix), "$")
	parts2 := strings.Split(strings.TrimPrefix(hash2, argon2idPrefix), "$")

	salt1 := parts1[0]
	salt2 := parts2[0]

	if salt1 == salt2 {
		t.Errorf("Salts for two different hash operations should be unique. Got identical salts: %s", salt1)
	}
}
