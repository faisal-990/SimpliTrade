package utils

import (
	"os"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/golang-jwt/jwt/v5"
)

// Use HS256 signing (symmetric key)
var jwtkey = []byte(os.Getenv("JWT_KEY"))

// GenerateJwt creates a JWT token using HS256
func GenerateJwt(username string) (string, error) {
	claims := &dto.Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}

	// Use HS256 since our key is a simple byte slice
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtkey)
	if err != nil {
		LogError("failed to generate jwt token", err)
		return "", err // return actual error
	}

	return signedToken, nil
}

// ValidateJwt verifies the token and returns claims
func ValidateJwt(tokenString string) (*dto.Claims, error) {
	claims := &dto.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HS256
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return jwtkey, nil
	})
	if err != nil {
		LogError("invalid token", err)
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidId
	}

	return claims, nil
}
