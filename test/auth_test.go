package test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Zoisit/web-servers-chirpy/internal/auth"
	"github.com/google/uuid"
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

func TestCreateJWTAndValidateJWT_Success(t *testing.T) {
	secret := "supersecret"
	userId := uuid.New()
	expiresIn := time.Minute

	token, err := auth.MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	returnedID, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT error: %v", err)
	}

	if returnedID != userId {
		t.Errorf("Returned UUID mismatch: got %v, want %v", returnedID, userId)
	}
}

func TestValidateJWT_Expired(t *testing.T) {
	secret := "supersecret"
	userId := uuid.New()
	expiresIn := -time.Minute // already expired

	token, err := auth.MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	_, err = auth.ValidateJWT(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	secret := "supersecret"
	wrongSecret := "wrongsecret"
	userId := uuid.New()
	expiresIn := time.Minute

	token, err := auth.MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	_, err = auth.ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("Expected error for wrong secret, got nil")
	}
}

// from solution
func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "Valid Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer valid_token"},
			},
			wantToken: "valid_token",
			wantErr:   false,
		},
		{
			name:      "Missing Authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Malformed Authorization header",
			headers: http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := auth.GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
