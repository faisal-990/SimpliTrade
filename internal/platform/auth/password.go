// Package auth holds the pure, dependency-light primitives of authentication:
// password hashing and JWT/refresh-token management. It has no knowledge of
// HTTP, GORM, or the database, so it is trivially unit-testable.
package auth

import "golang.org/x/crypto/bcrypt"

// MaxPasswordBytes is bcrypt's hard input limit; inputs beyond this are silently
// truncated by the algorithm, so we reject them explicitly at the edge instead.
const MaxPasswordBytes = 72

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword reports whether plain matches the stored bcrypt hash. It is
// constant-time with respect to the hash (bcrypt guarantees this) and never
// reveals why a comparison failed.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
