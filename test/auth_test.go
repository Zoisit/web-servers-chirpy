package test

import (
	"testing"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "mySuperSecretPassword"

	// Hash erstellen
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}

	// Check mit korrektem Passwort
	err = auth.CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash failed with correct password: %v", err)
	}

	// Check mit falschem Passwort
	err = auth.CheckPasswordHash("wrongPassword", hash)
	if err == nil {
		t.Error("CheckPasswordHash did not fail with wrong password")
	}
}
