package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MFASetupResult struct {
	Secret        string   `json:"secret"` // Base32 encoded secret
	QRCodeURL     string   `json:"qr_code_url"`
	RecoveryCodes []string `json:"recovery_codes"`
}

type MFAVerificationReq struct {
	Secret   string `json:"secret"`
	Passcode string `json:"passcode"`
}

type MFARecoveryReq struct {
	RecoveryCode string `json:"recovery_code"`
}

type MFAEngine struct{}

func NewMFAEngine() *MFAEngine {
	return &MFAEngine{}
}

// GenerateSetup initiates MFA enrollment with a base32 secret and recovery codes.
func (m *MFAEngine) GenerateSetup(issuer, userEmail string) (*MFASetupResult, error) {
	if userEmail == "" {
		return nil, fmt.Errorf("user email is required")
	}
	if issuer == "" {
		issuer = "Lopor AI Workspace"
	}

	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random secret: %w", err)
	}

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	secret := encoder.EncodeToString(secretBytes)

	qrURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, userEmail, secret, issuer)

	recoveryCodes := make([]string, 8)
	for i := 0; i < 8; i++ {
		codeBytes := make([]byte, 5)
		rand.Read(codeBytes)
		recoveryCodes[i] = fmt.Sprintf("%04x-%04x", codeBytes[:2], codeBytes[2:4])
	}

	return &MFASetupResult{
		Secret:        secret,
		QRCodeURL:     qrURL,
		RecoveryCodes: recoveryCodes,
	}, nil
}

// VerifyCode validates a 6-digit TOTP code against a base32 secret with ±1 time-step clock skew tolerance.
func (m *MFAEngine) VerifyCode(secret, passcode string) (bool, error) {
	passcode = strings.TrimSpace(passcode)
	if len(passcode) != 6 {
		return false, fmt.Errorf("passcode must be 6 digits")
	}

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	key, err := encoder.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return false, fmt.Errorf("invalid base32 secret: %w", err)
	}

	currentStep := time.Now().Unix() / 30

	// Check skew window: previous step (-1), current step (0), next step (+1)
	for skew := int64(-1); skew <= 1; skew++ {
		step := currentStep + skew
		generated := m.generateHOTP(key, step, 6)
		if generated == passcode {
			return true, nil
		}
	}

	return false, nil
}

// GenerateCurrentCode generates current 6-digit code for a secret (useful in testing).
func (m *MFAEngine) GenerateCurrentCode(secret string) (string, error) {
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	key, err := encoder.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", err
	}
	step := time.Now().Unix() / 30
	return m.generateHOTP(key, step, 6), nil
}

func (m *MFAEngine) generateHOTP(key []byte, counter int64, digits int) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	modulo := uint32(math.Pow10(digits))
	code := truncated % modulo

	return fmt.Sprintf("%0*d", digits, code)
}

// ValidateRecoveryCode checks single-use recovery code validity.
func (m *MFAEngine) ValidateRecoveryCode(providedCode string, validCodes []string) (bool, []string) {
	cleanProvided := strings.ToLower(strings.TrimSpace(providedCode))
	var remaining []string
	matched := false

	for _, code := range validCodes {
		if strings.ToLower(strings.TrimSpace(code)) == cleanProvided && !matched {
			matched = true
			continue // Consume this single-use code
		}
		remaining = append(remaining, code)
	}

	return matched, remaining
}

// GenerateUserMFARecord creates empty MFA profile.
func GenerateUserMFARecord(userID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"user_id":    userID,
		"enabled":    false,
		"created_at": time.Now().Format(time.RFC3339),
	}
}
