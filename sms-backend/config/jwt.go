package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines what data lives inside the JWT access token.
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SSEClaims is a short-lived token used only for the EventSource URL.
// replaces passing the full access JWT as a query param (which gets logged).
type SSEClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a 1-hour JWT — used after login
func GenerateAccessToken(userID uint, role string, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := Claims{
		UserID: userID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sms-backend",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a 7-day token for refreshing access.
// The raw token is stored by the caller as a bcrypt hash in the DB.
func GenerateRefreshToken(userID uint) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "sms-backend",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
}

// GenerateSSEToken creates a 60-second token for use in EventSource URLs.
// replaces passing the long-lived access JWT in query strings (visible in logs).
// Signed with JWT_SSE_SECRET (separate secret — if not set, falls back to JWT_SECRET).
func GenerateSSEToken(userID uint, role string, email string) (string, error) {
	secret := os.Getenv("JWT_SSE_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET") // fallback for existing deployments
	}
	claims := SSEClaims{
		UserID: userID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sms-backend-sse",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseAccessToken validates and parses a JWT string into Claims
func ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			// Always verify the signing method — prevents algorithm switching attacks
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// ParseSSEToken validates a short-lived SSE token and returns claims.
func ParseSSEToken(tokenStr string) (*SSEClaims, error) {
	secret := os.Getenv("JWT_SSE_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&SSEClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*SSEClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid SSE token claims")
	}
	// Extra guard: reject if signed with the wrong issuer
	if claims.Issuer != "sms-backend-sse" {
		return nil, errors.New("token is not an SSE token")
	}
	return claims, nil
}
