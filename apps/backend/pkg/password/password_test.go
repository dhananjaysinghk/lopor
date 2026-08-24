package password

import "testing"

func TestPasswordHashing(t *testing.T) {
	rawPassword := "SuperSecretPass123!"

	hash, err := HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("Password hashing failed: %v", err)
	}

	if hash == "" {
		t.Fatal("Expected hashed password string, got empty")
	}

	if !CheckPasswordHash(rawPassword, hash) {
		t.Error("CheckPasswordHash failed for correct password")
	}

	if CheckPasswordHash("WrongPass123!", hash) {
		t.Error("CheckPasswordHash succeeded for wrong password")
	}
}
