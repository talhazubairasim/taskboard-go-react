package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func init() { BcryptCost = bcrypt.MinCost }

func TestAccessTokenRoundTrip(t *testing.T) {
	m := NewJWTManager("secret", "refresh")
	tok, err := m.GenerateAccessToken("u1", "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	c, err := m.ValidateAccessToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.UserID != "u1" || c.Email != "a@b.com" {
		t.Fatalf("claims = %+v", c)
	}
}

func TestRefreshTokenRoundTripAndSeparation(t *testing.T) {
	m := NewJWTManager("secret", "refresh")
	rt, err := m.GenerateRefreshToken("u1", "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ValidateRefreshToken(rt); err != nil {
		t.Fatal(err)
	}
	// A refresh token must not work as an access token and vice versa.
	if _, err := m.ValidateAccessToken(rt); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("refresh token accepted as access token: %v", err)
	}
	at, _ := m.GenerateAccessToken("u1", "a@b.com")
	if _, err := m.ValidateRefreshToken(at); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("access token accepted as refresh token: %v", err)
	}
	rt2, _ := m.GenerateRefreshToken("u1", "a@b.com")
	if rt == rt2 {
		t.Fatal("refresh tokens should be unique")
	}
}

func TestExpiredToken(t *testing.T) {
	m := NewJWTManagerWithExpiry("s", "r", -time.Minute, -time.Minute)
	tok, _ := m.GenerateAccessToken("u1", "e")
	if _, err := m.ValidateAccessToken(tok); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("want expired, got %v", err)
	}
}

func TestRejectsOtherSigningMethods(t *testing.T) {
	m := NewJWTManager("s", "r")
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: "x"})
	s, _ := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := m.ValidateAccessToken(s); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("none alg accepted: %v", err)
	}
	if _, err := m.ValidateAccessToken("garbage"); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("garbage accepted")
	}
}

func TestValidatePassword(t *testing.T) {
	cases := map[string]error{
		"Sh0rt!":       ErrPasswordTooShort,
		"lowercase1!":  ErrPasswordNoUppercase,
		"UPPERCASE1!":  ErrPasswordNoLowercase,
		"NoNumbers!!":  ErrPasswordNoNumber,
		"NoSpecial123": ErrPasswordNoSpecial,
		"Good$Pass123": nil,
	}
	for pw, want := range cases {
		if got := ValidatePassword(pw); !errors.Is(got, want) {
			t.Errorf("%q: got %v want %v", pw, got, want)
		}
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	h, err := HashPassword("Good$Pass123")
	if err != nil {
		t.Fatal(err)
	}
	if CheckPassword("Good$Pass123", h) != nil {
		t.Fatal("correct password rejected")
	}
	if !errors.Is(CheckPassword("wrong", h), ErrInvalidPassword) {
		t.Fatal("wrong password accepted")
	}
	if _, err := HashPassword("weak"); err == nil {
		t.Fatal("weak password hashed")
	}
	BcryptCost = 99 // invalid cost makes bcrypt fail
	defer func() { BcryptCost = bcrypt.MinCost }()
	if _, err := HashPassword("Good$Pass123"); err == nil {
		t.Fatal("expected bcrypt error")
	}
}

func TestBearerToken(t *testing.T) {
	if tok, ok := BearerToken("Bearer abc"); !ok || tok != "abc" {
		t.Fatal("valid bearer rejected")
	}
	if tok, ok := BearerToken("bearer  abc "); !ok || tok != "abc" {
		t.Fatal("case-insensitive bearer rejected")
	}
	for _, h := range []string{"", "Bearer", "Basic abc", "Bearer  "} {
		if _, ok := BearerToken(h); ok {
			t.Errorf("%q accepted", h)
		}
	}
}

func TestMiddleware(t *testing.T) {
	m := NewJWTManager("s", "r")
	mw := NewAuthMiddleware(m)
	var got *Claims
	h := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusTeapot)
	}))

	// anonymous passes through
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusTeapot || got != nil {
		t.Fatalf("anonymous: code=%d claims=%v", rec.Code, got)
	}

	// valid token
	tok, _ := m.GenerateAccessToken("u1", "e")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot || got == nil || got.UserID != "u1" {
		t.Fatalf("valid token: code=%d claims=%v", rec.Code, got)
	}

	// malformed header and bad token are rejected
	for _, hv := range []string{"Token abc", "Bearer nope"} {
		req = httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", hv)
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%q: code=%d", hv, rec.Code)
		}
	}
}

func TestClaimsFromHeaderAndRequireAuth(t *testing.T) {
	m := NewJWTManager("s", "r")
	mw := NewAuthMiddleware(m)
	tok, _ := m.GenerateAccessToken("u9", "e")
	c, err := mw.ClaimsFromHeader("Bearer " + tok)
	if err != nil || c.UserID != "u9" {
		t.Fatalf("claims=%v err=%v", c, err)
	}
	if _, err := mw.ClaimsFromHeader("nope"); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("bad header accepted")
	}
	if _, err := RequireAuth(context.Background()); !errors.Is(err, ErrUnauthenticated) {
		t.Fatal("anonymous context passed RequireAuth")
	}
	ctx := ContextWithClaims(context.Background(), c)
	if got, err := RequireAuth(ctx); err != nil || got.UserID != "u9" {
		t.Fatal("RequireAuth failed with claims")
	}
}
