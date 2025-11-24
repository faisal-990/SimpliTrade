package dto

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Id uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}
