package service

import (
	"context"
	"net/url"
	"strings"
	"time"

	"taskboard/internal/cache"
	"taskboard/internal/models"
)

// UserService reads and updates user profiles.
type UserService struct {
	users UserStore
	cache Cache
	ttl   time.Duration
}

// NewUserService builds a UserService. c may be nil to disable caching.
func NewUserService(users UserStore, c Cache, ttl time.Duration) *UserService {
	if c == nil {
		c = noopCache{}
	}
	return &UserService{users: users, cache: c, ttl: ttl}
}

// Get returns one user, served from cache when possible.
func (s *UserService) Get(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	if err := s.cache.Get(ctx, cache.UserKey(id), &u); err == nil {
		return &u, nil
	}
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fromRepo(err)
	}
	_ = s.cache.Set(ctx, cache.UserKey(id), user, s.ttl)
	return user, nil
}

// List returns every user.
func (s *UserService) List(ctx context.Context) ([]*models.User, error) {
	return s.users.List(ctx)
}

// UpdateProfile changes the caller's name and/or avatar URL.
func (s *UserService) UpdateProfile(ctx context.Context, userID string, upd models.UserUpdate) (*models.User, error) {
	if upd.Name != nil {
		name, err := validateName(*upd.Name)
		if err != nil {
			return nil, err
		}
		upd.Name = &name
	}
	if upd.Avatar != nil {
		a := strings.TrimSpace(*upd.Avatar)
		if a != "" {
			u, err := url.Parse(a)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(a) > 2048 {
				return nil, invalid("avatar must be an http(s) URL")
			}
		}
		upd.Avatar = &a
	}
	user, err := s.users.Update(ctx, userID, upd)
	if err != nil {
		return nil, fromRepo(err)
	}
	// Cached tasks embed user names and avatars, so drop them too.
	_ = s.cache.Delete(ctx, cache.UserKey(userID))
	_ = s.cache.DeletePattern(ctx, cache.TasksPrefix+"*")
	_ = s.cache.DeletePattern(ctx, "task:*")
	return user, nil
}
