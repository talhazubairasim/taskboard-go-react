package graph

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"taskboard/graph/model"
	"taskboard/internal/models"
	"taskboard/internal/rest"
	"taskboard/internal/service"
)

func toUser(u *models.User) *model.User {
	if u == nil {
		return nil
	}
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toTask(t *models.TaskDetails) *model.Task {
	return &model.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      model.TaskStatus(t.Status),
		Priority:    model.Priority(t.Priority),
		CreatedBy:   toUser(t.CreatedBy),
		AssignedTo:  toUser(t.AssignedTo),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		DueDate:     t.DueDate,
	}
}

func toTasks(ts []*models.TaskDetails) []*model.Task {
	out := make([]*model.Task, 0, len(ts))
	for _, t := range ts {
		out = append(out, toTask(t))
	}
	return out
}

func toAuth(r *service.AuthResult) *model.AuthPayload {
	return &model.AuthPayload{Token: r.Token, RefreshToken: r.RefreshToken, User: toUser(r.User)}
}

func enumPtr[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ErrorPresenter adds extensions.code to every error so clients can react to
// it (the frontend refreshes its token on UNAUTHENTICATED), and hides the
// details of unexpected internal errors.
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	if gqlErr.Extensions == nil {
		gqlErr.Extensions = map[string]interface{}{}
	}
	if _, ok := gqlErr.Extensions["code"]; ok {
		return gqlErr
	}
	status, code := rest.StatusFor(err)
	if status == http.StatusInternalServerError {
		var ge *gqlerror.Error
		if errors.As(err, &ge) && ge.Err == nil {
			// An error produced by gqlgen itself, e.g. bad input coercion.
			code = "BAD_USER_INPUT"
		} else {
			log.Printf("graphql: internal error: %v", err)
			gqlErr.Message = "internal server error"
		}
	}
	gqlErr.Extensions["code"] = code
	return gqlErr
}
