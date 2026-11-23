// Package service holds TaskBoard's business rules. Both the REST handlers and
// the GraphQL resolvers call into it, so the rules live in exactly one place.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/models"
	"taskboard/internal/pubsub"
	"taskboard/internal/repository"
)

// Errors returned by services. Transport layers map them to HTTP status codes
// or GraphQL error codes.
var (
	ErrUnauthenticated    = auth.ErrUnauthenticated
	ErrForbidden          = errors.New("you do not have permission to do that")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// ValidationError describes bad input.
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

func invalid(format string, args ...interface{}) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// UserStore is the persistence the services need for users.
type UserStore interface {
	Create(ctx context.Context, u *models.User) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]*models.User, error)
	List(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, id string, upd models.UserUpdate) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

// TaskStore is the persistence the services need for tasks.
type TaskStore interface {
	Create(ctx context.Context, t *models.Task) (*models.Task, error)
	GetByID(ctx context.Context, id string) (*models.Task, error)
	List(ctx context.Context, f models.TaskFilter) ([]*models.Task, error)
	Update(ctx context.Context, id string, u models.TaskUpdate) (*models.Task, error)
	Delete(ctx context.Context, id string) error
}

// Cache is an optional read-through cache.
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// Publisher receives task change events for real-time subscribers.
type Publisher interface {
	Publish(e pubsub.Event)
}

type noopCache struct{}

func (noopCache) Get(context.Context, string, interface{}) error {
	return cache.ErrCacheMiss
}
func (noopCache) Set(context.Context, string, interface{}, time.Duration) error { return nil }
func (noopCache) Delete(context.Context, ...string) error                       { return nil }
func (noopCache) DeletePattern(context.Context, string) error                   { return nil }

type noopPublisher struct{}

func (noopPublisher) Publish(pubsub.Event) {}

// fromRepo converts repository errors to service errors.
func fromRepo(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrDuplicate):
		return ErrConflict
	case errors.Is(err, repository.ErrInvalidReference):
		return invalid("referenced user does not exist")
	}
	return err
}
