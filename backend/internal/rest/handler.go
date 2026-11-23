// Package rest exposes TaskBoard's REST API under /api. It calls the same
// services as the GraphQL API, so both behave identically.
package rest

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"taskboard/internal/auth"
	"taskboard/internal/models"
	"taskboard/internal/service"
)

// Handler serves the REST API.
type Handler struct {
	Auth  *service.AuthService
	Users *service.UserService
	Tasks *service.TaskService
}

// Register mounts the routes on mux. Requests must already have passed the
// auth middleware so that claims are on the context.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/refresh", h.refresh)

	mux.HandleFunc("GET /api/me", h.withUser(h.me))
	mux.HandleFunc("PATCH /api/me", h.withUser(h.updateMe))
	mux.HandleFunc("GET /api/users", h.withUser(h.listUsers))
	mux.HandleFunc("GET /api/users/{id}", h.withUser(h.getUser))

	mux.HandleFunc("GET /api/tasks", h.withUser(h.listTasks))
	mux.HandleFunc("POST /api/tasks", h.withUser(h.createTask))
	mux.HandleFunc("GET /api/tasks/mine", h.withUser(h.myTasks))
	mux.HandleFunc("GET /api/tasks/assigned", h.withUser(h.assignedTasks))
	mux.HandleFunc("GET /api/tasks/{id}", h.withUser(h.getTask))
	mux.HandleFunc("PATCH /api/tasks/{id}", h.withUser(h.updateTask))
	mux.HandleFunc("PUT /api/tasks/{id}", h.withUser(h.updateTask))
	mux.HandleFunc("DELETE /api/tasks/{id}", h.withUser(h.deleteTask))
	mux.HandleFunc("POST /api/tasks/{id}/assign", h.withUser(h.assignTask))
	mux.HandleFunc("POST /api/tasks/{id}/unassign", h.withUser(h.unassignTask))
}

type userHandler func(w http.ResponseWriter, r *http.Request, userID string)

// withUser rejects anonymous requests.
func (h *Handler) withUser(next userHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.RequireAuth(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		next(w, r, claims.UserID)
	}
}

// ---------- helpers ----------

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// StatusFor maps a service error to an HTTP status and machine-readable code.
func StatusFor(err error) (int, string) {
	var v *service.ValidationError
	switch {
	case errors.As(err, &v):
		return http.StatusBadRequest, "BAD_USER_INPUT"
	case errors.Is(err, service.ErrUnauthenticated), errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	case errors.Is(err, service.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, service.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	}
	return http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"
}

func writeError(w http.ResponseWriter, err error) {
	status, code := StatusFor(err)
	msg := err.Error()
	if status == http.StatusInternalServerError {
		log.Printf("rest: internal error: %v", err)
		msg = "internal server error"
	}
	writeJSON(w, status, errorBody{Error: msg, Code: code})
}

func decode(r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return &service.ValidationError{Message: "invalid JSON body"}
	}
	return nil
}

// ---------- auth ----------

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password, Name string }
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	res, err := h.Auth.Register(r.Context(), in.Email, in.Password, in.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	res, err := h.Auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	res, err := h.Auth.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// ---------- users ----------

func (h *Handler) me(w http.ResponseWriter, r *http.Request, userID string) {
	u, err := h.Users.Get(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request, userID string) {
	var in struct {
		Name   *string `json:"name"`
		Avatar *string `json:"avatar"`
	}
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.Users.UpdateProfile(r.Context(), userID, models.UserUpdate{Name: in.Name, Avatar: in.Avatar})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request, _ string) {
	users, err := h.Users.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request, _ string) {
	u, err := h.Users.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// ---------- tasks ----------

func (h *Handler) list(w http.ResponseWriter, r *http.Request, f models.TaskFilter) {
	tasks, err := h.Tasks.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request, _ string) {
	q := r.URL.Query()
	h.list(w, r, models.TaskFilter{
		Status:       q.Get("status"),
		Priority:     q.Get("priority"),
		AssignedToID: q.Get("assignedToId"),
		CreatedByID:  q.Get("createdById"),
	})
}

func (h *Handler) myTasks(w http.ResponseWriter, r *http.Request, userID string) {
	h.list(w, r, models.TaskFilter{CreatedByID: userID})
}

func (h *Handler) assignedTasks(w http.ResponseWriter, r *http.Request, userID string) {
	h.list(w, r, models.TaskFilter{AssignedToID: userID})
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request, userID string) {
	var in service.CreateTaskInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	t, err := h.Tasks.Create(r.Context(), userID, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, _ string) {
	t, err := h.Tasks.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// parseUpdate reads a partial update. A JSON null clears description,
// assignedToId or dueDate; an absent field leaves it unchanged.
func parseUpdate(r *http.Request) (models.TaskUpdate, error) {
	var raw map[string]json.RawMessage
	var u models.TaskUpdate
	if err := decode(r, &raw); err != nil {
		return u, err
	}
	isNull := func(m json.RawMessage) bool { return strings.TrimSpace(string(m)) == "null" }
	str := func(key string) (*string, bool, error) {
		m, ok := raw[key]
		if !ok {
			return nil, false, nil
		}
		if isNull(m) {
			return nil, true, nil
		}
		var s string
		if err := json.Unmarshal(m, &s); err != nil {
			return nil, false, &service.ValidationError{Message: key + " must be a string"}
		}
		return &s, false, nil
	}
	var err error
	var cleared bool
	if u.Title, _, err = str("title"); err != nil {
		return u, err
	}
	if u.Description, u.ClearDescription, err = str("description"); err != nil {
		return u, err
	}
	if u.Status, _, err = str("status"); err != nil {
		return u, err
	}
	if u.Priority, _, err = str("priority"); err != nil {
		return u, err
	}
	if u.AssignedToID, u.ClearAssignee, err = str("assignedToId"); err != nil {
		return u, err
	}
	var due *string
	if due, cleared, err = str("dueDate"); err != nil {
		return u, err
	}
	u.ClearDueDate = cleared
	if due != nil {
		t, perr := time.Parse(time.RFC3339, *due)
		if perr != nil {
			return u, &service.ValidationError{Message: "dueDate must be an RFC 3339 timestamp"}
		}
		u.DueDate = &t
	}
	return u, nil
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, userID string) {
	u, err := parseUpdate(r)
	if err != nil {
		writeError(w, err)
		return
	}
	t, err := h.Tasks.Update(r.Context(), userID, r.PathValue("id"), u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.Tasks.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) assignTask(w http.ResponseWriter, r *http.Request, userID string) {
	var in struct {
		UserID string `json:"userId"`
	}
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	t, err := h.Tasks.Assign(r.Context(), userID, r.PathValue("id"), in.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) unassignTask(w http.ResponseWriter, r *http.Request, userID string) {
	t, err := h.Tasks.Unassign(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}
