package dtos

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// JWT Claims
type JWTClaims struct {
	UserSession UserSession `json:"user_session"`
	Type        string      `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}
