package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a raw password string using bcrypt
func HashPassword(rawPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a raw password against its bcrypt hash
func CheckPasswordHash(rawPassword, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawPassword))
	return err == nil
}
