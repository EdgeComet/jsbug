package session

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func TestHashFingerprint(t *testing.T) {
	tests := []struct {
		name       string
		userAgent  string
		acceptLang string
		acceptEnc  string
	}{
		{
			name:       "typical browser headers",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
			acceptLang: "en-US,en;q=0.9",
			acceptEnc:  "gzip, deflate, br",
		},
		{
			name:       "empty headers",
			userAgent:  "",
			acceptLang: "",
			acceptEnc:  "",
		},
		{
			name:       "mobile browser",
			userAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X)",
			acceptLang: "en-GB",
			acceptEnc:  "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := HashFingerprint(tt.userAgent, tt.acceptLang, tt.acceptEnc)
			hash2 := HashFingerprint(tt.userAgent, tt.acceptLang, tt.acceptEnc)

			// Hash should be deterministic
			if hash1 != hash2 {
				t.Errorf("HashFingerprint not deterministic: %s != %s", hash1, hash2)
			}

			// Hash should be 64 characters (SHA-256 in hex)
			if len(hash1) != 64 {
				t.Errorf("Expected hash length 64, got %d", len(hash1))
			}
		})
	}

	// Different inputs should produce different hashes
	hash1 := HashFingerprint("UA1", "en", "gzip")
	hash2 := HashFingerprint("UA2", "en", "gzip")
	if hash1 == hash2 {
		t.Error("Different inputs produced same hash")
	}
}

func TestNewTokenManager_SecretKeyTooShort(t *testing.T) {
	logger := zap.NewNop()

	// Test with key that's too short
	_, err := NewTokenManager("short-key", logger)
	if err != ErrSecretKeyTooShort {
		t.Errorf("Expected ErrSecretKeyTooShort, got %v", err)
	}

	// Test with key that's exactly 31 bytes (one less than minimum)
	_, err = NewTokenManager("1234567890123456789012345678901", logger)
	if err != ErrSecretKeyTooShort {
		t.Errorf("Expected ErrSecretKeyTooShort for 31-byte key, got %v", err)
	}

	// Test with key that's exactly 32 bytes (minimum)
	tm, err := NewTokenManager("12345678901234567890123456789012", logger)
	if err != nil {
		t.Errorf("Expected no error for 32-byte key, got %v", err)
	}
	if tm == nil {
		t.Error("Expected non-nil TokenManager for valid key")
	}
}

func TestTokenManager_GenerateToken(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	token, expiresAt, err := tm.GenerateToken(fingerprint)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token is empty")
	}

	// Expiry should be ~24 hours from now
	expectedExpiry := time.Now().Add(TokenDuration)
	if expiresAt.Sub(expectedExpiry).Abs() > time.Second {
		t.Errorf("Expiry time mismatch: got %v, expected ~%v", expiresAt, expectedExpiry)
	}
}

func TestTokenManager_ValidateToken_Success(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	token, _, err := tm.GenerateToken(fingerprint)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	err = tm.ValidateToken(token, fingerprint)
	if err != nil {
		t.Errorf("ValidateToken failed for valid token: %v", err)
	}
}

func TestTokenManager_ValidateToken_EmptyToken(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	err = tm.ValidateToken("", fingerprint)
	if err != ErrTokenRequired {
		t.Errorf("Expected ErrTokenRequired, got %v", err)
	}
}

func TestTokenManager_ValidateToken_InvalidToken(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	err = tm.ValidateToken("not-a-valid-jwt", fingerprint)
	if err != ErrTokenInvalid {
		t.Errorf("Expected ErrTokenInvalid, got %v", err)
	}
}

func TestTokenManager_ValidateToken_WrongSecret(t *testing.T) {
	logger := zap.NewNop()
	tm1, err := NewTokenManager("secret-key-one-32-bytes-long!!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	tm2, err := NewTokenManager("secret-key-two-32-bytes-long!!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	token, _, err := tm1.GenerateToken(fingerprint)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate with different secret should fail
	err = tm2.ValidateToken(token, fingerprint)
	if err != ErrTokenInvalid {
		t.Errorf("Expected ErrTokenInvalid for wrong secret, got %v", err)
	}
}

func TestTokenManager_ValidateToken_FingerprintMismatch(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint1 := HashFingerprint("Mozilla/5.0", "en-US", "gzip")
	fingerprint2 := HashFingerprint("Safari/17.0", "fr-FR", "br")

	token, _, err := tm.GenerateToken(fingerprint1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate with different fingerprint should fail
	err = tm.ValidateToken(token, fingerprint2)
	if err != ErrFingerprintMismatch {
		t.Errorf("Expected ErrFingerprintMismatch, got %v", err)
	}
}

func TestTokenManager_ValidateToken_Expired(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	// Create an already-expired token manually
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-25 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expired 1 hour ago
		},
		Fingerprint: fingerprint,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key-32-bytes-long!!!"))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	err = tm.ValidateToken(tokenString, fingerprint)
	if err != ErrTokenExpired {
		t.Errorf("Expected ErrTokenExpired, got %v", err)
	}
}

func TestTokenManager_ValidateToken_WrongSigningMethod(t *testing.T) {
	logger := zap.NewNop()
	tm, err := NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}
	fingerprint := HashFingerprint("Mozilla/5.0", "en-US", "gzip")

	// Create a token with a different signing method (none)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenDuration)),
		},
		Fingerprint: fingerprint,
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	err = tm.ValidateToken(tokenString, fingerprint)
	if err != ErrTokenInvalid {
		t.Errorf("Expected ErrTokenInvalid for wrong signing method, got %v", err)
	}
}
