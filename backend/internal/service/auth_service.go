package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"taskboard/internal/auth"
	"taskboard/internal/models"
)

// AuthResult is returned by register, login and refresh.
type AuthResult struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refreshToken"`
	User         *models.User `json:"user"`
}

// AuthService registers users and issues tokens.
type AuthService struct {
	users UserStore
	jwt   *auth.JWTManager
}

// NewAuthService builds an AuthService.
func NewAuthService(users UserStore, jwt *auth.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

func validateEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 255 {
		return "", invalid("a valid email address is required")
	}
	return email, nil
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return "", invalid("name must be between 1 and 255 characters")
	}
	return name, nil
}

// Register creates an account and returns tokens for it.
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*AuthResult, error) {
	email, err := validateEmail(email)
	if err != nil {
		return nil, err
	}
	if name, err = validateName(name); err != nil {
		return nil, err
	}
	if err := auth.ValidatePassword(password); err != nil {
		return nil, invalid("%s", err.Error())
	}
	exists, err := s.users.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user, err := s.users.Create(ctx, &models.User{Email: email, PasswordHash: hash, Name: name})
	if err != nil {
		return nil, fromRepo(err)
	}
	return s.issue(user)
}

// Login checks credentials and returns tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(fromRepo(err), ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if auth.CheckPassword(password, user.PasswordHash) != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issue(user)
}

// Refresh exchanges a valid refresh token for a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrUnauthenticated
	}
	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(fromRepo(err), ErrNotFound) {
			return nil, ErrUnauthenticated
		}
		return nil, err
	}
	return s.issue(user)
}

func (s *AuthService) issue(user *models.User) (*AuthResult, error) {
	access, err := s.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	refresh, err := s.jwt.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: access, RefreshToken: refresh, User: user}, nil
}
