package graph

import (
	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/repository"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	userRepo   *repository.UserRepository
	taskRepo   *repository.TaskRepository
	cache      *cache.RedisCache
	jwtManager *auth.JWTManager
}

func NewResolver(
	userRepo *repository.UserRepository,
	taskRepo *repository.TaskRepository,
	cache *cache.RedisCache,
	jwtManager *auth.JWTManager,
) *Resolver {
	return &Resolver{
		userRepo:   userRepo,
		taskRepo:   taskRepo,
		cache:      cache,
		jwtManager: jwtManager,
	}
}