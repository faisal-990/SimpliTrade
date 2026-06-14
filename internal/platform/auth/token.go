package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when an access token fails validation.
var ErrInvalidToken = errors.New("auth: invalid token")

// Claims is the JWT payload. It carries the identity the API needs on every
// request — the user, their active account (so trade/portfolio handlers know
// which balance to touch), and role — so handlers never re-query just to learn
// who is calling.
type Claims struct {
	UserID    string `json:"uid"`
	AccountID string `json:"aid"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager issues and validates access tokens and mints opaque refresh
// tokens. It is constructed from config so the signing secret is never read
// from a package-level global.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager builds a TokenManager. secret must be non-empty (config
// guarantees this).
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccessToken signs a short-lived HS256 access token and returns it
// alongside its expiry.
func (m *TokenManager) GenerateAccessToken(userID, accountID, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTTL)
	claims := &Claims{
		UserID:    userID,
		AccountID: accountID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// ParseAccessToken verifies signature, algorithm, and expiry, returning the
// claims. It rejects any non-HMAC signing method to prevent algorithm
// confusion (e.g. "alg: none" or RS256 with the secret as a public key).
func (m *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// RefreshToken is a freshly minted refresh token: the opaque Plaintext is
// returned to the client exactly once; only Hash is persisted server-side.
type RefreshToken struct {
	Plaintext string
	Hash      string
	ExpiresAt time.Time
}

// GenerateRefreshToken creates a cryptographically-random opaque token and its
// storable hash. The plaintext is never stored, so a DB leak cannot be used to
// impersonate users.
func (m *TokenManager) GenerateRefreshToken() (RefreshToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return RefreshToken{}, err
	}
	plaintext := hex.EncodeToString(b)
	return RefreshToken{
		Plaintext: plaintext,
		Hash:      HashToken(plaintext),
		ExpiresAt: time.Now().Add(m.refreshTTL),
	}, nil
}

// NewOpaqueToken mints a cryptographically-random opaque token and its storable
// SHA-256 hash. Used for one-off secrets like password-reset links, where the
// plaintext is delivered out-of-band (email) and only the hash is persisted.
func NewOpaqueToken() (plaintext, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	plaintext = hex.EncodeToString(b)
	return plaintext, HashToken(plaintext), nil
}

// HashToken returns the SHA-256 hex digest used to store/look up refresh tokens.
// Refresh tokens are high-entropy random values, so a fast hash is appropriate
// (unlike passwords, which require bcrypt).
func HashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
