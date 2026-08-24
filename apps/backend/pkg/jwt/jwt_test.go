package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-32-characters-minimum!"
	userID := uuid.New()
	email := "test@lopor.ai"
	role := "admin"

	// 1. Test Valid Token
	token, err := GenerateAccessToken(userID, email, role, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Token validation failed for valid token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}

	// 2. Test Invalid Secret
	_, err = ValidateToken(token, "wrong-secret-key!")
	if err == nil {
		t.Error("Expected error when validating with wrong secret, got nil")
	}

	// 3. Test Expired Token
	expiredToken, err := GenerateAccessToken(userID, email, role, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	_, err = ValidateToken(expiredToken, secret)
	if err == nil {
		t.Error("Expected error when validating expired token, got nil")
	}
}
