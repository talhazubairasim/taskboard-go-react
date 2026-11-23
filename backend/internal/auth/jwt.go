package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	// ErrUnauthenticated means the request carried no valid access token.
	ErrUnauthenticated = errors.New("authentication required")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey        []byte
	refreshSecretKey []byte
	accessDuration   time.Duration
	refreshDuration  time.Duration
}

// NewJWTManager creates a manager with the default lifetimes
// (15 minute access tokens, 7 day refresh tokens).
func NewJWTManager(secretKey, refreshSecretKey string) *JWTManager {
	return NewJWTManagerWithExpiry(secretKey, refreshSecretKey, 15*time.Minute, 7*24*time.Hour)
}

// NewJWTManagerWithExpiry creates a manager with custom token lifetimes.
func NewJWTManagerWithExpiry(secretKey, refreshSecretKey string, access, refresh time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:        []byte(secretKey),
		refreshSecretKey: []byte(refreshSecretKey),
		accessDuration:   access,
		refreshDuration:  refresh,
	}
}

// GenerateAccessToken creates a new JWT access token
func (m *JWTManager) GenerateAccessToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken creates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			// A unique ID keeps two refresh tokens issued in the same second distinct.
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.refreshSecretKey)
}

// ValidateAccessToken verifies and parses an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.secretKey)
}

// ValidateRefreshToken verifies and parses a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.refreshSecretKey)
}

func (m *JWTManager) validateToken(tokenString string, key []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return key, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
