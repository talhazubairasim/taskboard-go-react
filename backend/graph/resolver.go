package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"taskboard/internal/pubsub"
	"taskboard/internal/service"
)

// Resolver is the root GraphQL resolver. It delegates all work to the service
// layer, which the REST API shares.
type Resolver struct {
	AuthSvc *service.AuthService
	UserSvc *service.UserService
	TaskSvc *service.TaskService
	Broker  *pubsub.Broker
}
