package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

// UserContextKey is the context key under which validated claims are stored.
const UserContextKey contextKey = "user"

// AuthMiddleware validates bearer tokens and puts the claims on the request context.
type AuthMiddleware struct {
	jwtManager *JWTManager
}

// NewAuthMiddleware builds the middleware.
func NewAuthMiddleware(jwtManager *JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

// Middleware lets anonymous requests through (resolvers and handlers decide
// what needs auth) but rejects requests that carry a malformed or invalid token.
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}
		token, ok := BearerToken(header)
		if !ok {
			writeJSONError(w, "invalid authorization header format")
			return
		}
		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			writeJSONError(w, "invalid or expired token")
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithClaims(r.Context(), claims)))
	})
}

// ClaimsFromHeader validates an "Authorization: Bearer <token>" value.
// It is used for WebSocket connections, where the token arrives in the init payload.
func (m *AuthMiddleware) ClaimsFromHeader(header string) (*Claims, error) {
	token, ok := BearerToken(header)
	if !ok {
		return nil, ErrInvalidToken
	}
	return m.jwtManager.ValidateAccessToken(token)
}

// BearerToken extracts the token from an Authorization header value.
func BearerToken(header string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// ContextWithClaims returns a copy of ctx carrying the claims.
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}

// GetUserFromContext extracts user claims from context.
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok && claims != nil
}

// RequireAuth returns the caller's claims or ErrUnauthenticated.
func RequireAuth(ctx context.Context) (*Claims, error) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	return claims, nil
}

func writeJSONError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + msg + `","code":"UNAUTHENTICATED"}`))
}
