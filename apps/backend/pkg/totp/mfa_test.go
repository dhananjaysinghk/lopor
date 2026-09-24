package totp_test

import (
	"strings"
	"testing"

	"github.com/lopor-ai/lopor/pkg/totp"
)

func TestMFAEngine_SetupAndVerification(t *testing.T) {
	engine := totp.NewMFAEngine()

	// 1. Setup MFA
	setup, err := engine.GenerateSetup("Lopor AI", "developer@lopor.ai")
	if err != nil {
		t.Fatalf("unexpected error setting up MFA: %v", err)
	}

	if len(setup.Secret) == 0 {
		t.Errorf("expected non-empty secret")
	}

	if !strings.HasPrefix(setup.QRCodeURL, "otpauth://totp/") {
		t.Errorf("expected valid otpauth URI prefix, got %s", setup.QRCodeURL)
	}

	if len(setup.RecoveryCodes) != 8 {
		t.Errorf("expected 8 recovery codes, got %d", len(setup.RecoveryCodes))
	}

	// 2. Generate current valid 6-digit passcode
	code, err := engine.GenerateCurrentCode(setup.Secret)
	if err != nil {
		t.Fatalf("failed to generate current code: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("expected 6 digit code, got %s", code)
	}

	// 3. Verify valid passcode
	valid, err := engine.VerifyCode(setup.Secret, code)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if !valid {
		t.Errorf("expected valid code to be accepted")
	}

	// 4. Verify invalid passcode
	invalid, err := engine.VerifyCode(setup.Secret, "000000")
	if err == nil && invalid {
		t.Errorf("expected invalid code to be rejected")
	}

	// 5. Test Recovery code redemption
	testRecovery := setup.RecoveryCodes[0]
	matched, remaining := engine.ValidateRecoveryCode(testRecovery, setup.RecoveryCodes)
	if !matched {
		t.Errorf("expected recovery code match")
	}
	if len(remaining) != 7 {
		t.Errorf("expected 7 remaining recovery codes, got %d", len(remaining))
	}
}
