package session

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Session token duration (24 hours)
const TokenDuration = 24 * time.Hour

// Error types for token validation
var (
	ErrTokenRequired        = errors.New("session token required")
	ErrTokenInvalid         = errors.New("session token invalid")
	ErrTokenExpired         = errors.New("session token expired")
	ErrFingerprintMismatch  = errors.New("fingerprint mismatch")
)

// Claims represents the JWT claims for a session token
type Claims struct {
	jwt.RegisteredClaims
	Fingerprint string `json:"fingerprint"`
}

// TokenManager handles JWT session token operations
type TokenManager struct {
	secretKey []byte
	logger    *zap.Logger
}

// MinSecretKeyLength is the minimum required length for the secret key (32 bytes for HMAC-SHA256)
const MinSecretKeyLength = 32

// ErrSecretKeyTooShort is returned when the secret key is less than MinSecretKeyLength
var ErrSecretKeyTooShort = fmt.Errorf("secret key must be at least %d bytes for HMAC-SHA256", MinSecretKeyLength)

// NewTokenManager creates a new TokenManager with the given secret key
// Returns an error if the secret key is less than 32 bytes
func NewTokenManager(secretKey string, logger *zap.Logger) (*TokenManager, error) {
	if len(secretKey) < MinSecretKeyLength {
		return nil, ErrSecretKeyTooShort
	}
	return &TokenManager{
		secretKey: []byte(secretKey),
		logger:    logger,
	}, nil
}

// HashFingerprint creates a SHA-256 hash of the browser fingerprint headers
// Fingerprint = SHA256(User-Agent + "|" + Accept-Language + "|" + Accept-Encoding)
func HashFingerprint(userAgent, acceptLang, acceptEnc string) string {
	data := userAgent + "|" + acceptLang + "|" + acceptEnc
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateToken creates a new JWT session token with the given fingerprint
// Returns the token string and expiration time
func (m *TokenManager) GenerateToken(fingerprint string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(TokenDuration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Fingerprint: fingerprint,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		m.logger.Error("Failed to sign token", zap.Error(err))
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT session token and checks fingerprint match
// Returns nil if valid, or an appropriate error
func (m *TokenManager) ValidateToken(tokenString, fingerprint string) error {
	if tokenString == "" {
		return ErrTokenRequired
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secretKey, nil
	})

	if err != nil {
		// Check for specific error types
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrTokenExpired
		}
		m.logger.Debug("Token validation failed", zap.Error(err))
		return ErrTokenInvalid
	}

	if !token.Valid {
		return ErrTokenInvalid
	}

	// Extract claims and validate fingerprint
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return ErrTokenInvalid
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(claims.Fingerprint), []byte(fingerprint)) != 1 {
		m.logger.Warn("Fingerprint mismatch",
			zap.String("expected", fingerprint),
			zap.String("got", claims.Fingerprint),
		)
		return ErrFingerprintMismatch
	}

	return nil
}
