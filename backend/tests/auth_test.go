package tests

import (
	"testing"
	"time"

	"taskboard/internal/auth"
)

func TestJWTManager_GenerateAndValidateAccessToken(t *testing.T) {
	manager := auth.NewJWTManager("test-secret-key", "test-refresh-key")
	
	userID := "test-user-123"
	email := "test@example.com"
	
	token, err := manager.GenerateAccessToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}
	
	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}
	
	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
	
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager := auth.NewJWTManager("test-secret-key", "test-refresh-key")
	
	// This test would need to modify the expiration time or use time mocking
	// For now, we just test that an invalid token fails
	_, err := manager.ValidateAccessToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestJWTManager_RefreshToken(t *testing.T) {
	manager := auth.NewJWTManager("test-secret-key", "test-refresh-key")
	
	userID := "test-user-123"
	email := "test@example.com"
	
	refreshToken, err := manager.GenerateRefreshToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}
	
	claims, err := manager.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}
	
	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:     "Too short",
			password: "Short1!",
			wantErr:  true,
		},
		{
			name:     "No uppercase",
			password: "securepass123!",
			wantErr:  true,
		},
		{
			name:     "No lowercase",
			password: "SECUREPASS123!",
			wantErr:  true,
		},
		{
			name:     "No number",
			password: "SecurePass!",
			wantErr:  true,
		},
		{
			name:     "No special character",
			password: "SecurePass123",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "SecurePass123!"
	
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if hash == password {
		t.Error("Hash should not equal original password")
	}
	
	// Test password verification
	err = auth.CheckPassword(password, hash)
	if err != nil {
		t.Errorf("Password verification failed: %v", err)
	}
	
	// Test wrong password
	err = auth.CheckPassword("WrongPassword123!", hash)
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}

func TestHashPassword_InvalidPassword(t *testing.T) {
	invalidPasswords := []string{
		"short",
		"nouppercase1!",
		"NOLOWERCASE1!",
		"NoNumbers!",
		"NoSpecialChar123",
	}
	
	for _, pwd := range invalidPasswords {
		_, err := auth.HashPassword(pwd)
		if err == nil {
			t.Errorf("Expected error for invalid password '%s', got nil", pwd)
		}
	}
}