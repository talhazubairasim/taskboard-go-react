// Package models holds the domain types shared by every layer of the backend.
package models

import "time"

// User is a registered TaskBoard user.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Avatar       *string   `json:"avatar,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Task statuses, in Kanban column order.
const (
	StatusTodo       = "TODO"
	StatusInProgress = "IN_PROGRESS"
	StatusReview     = "REVIEW"
	StatusDone       = "DONE"
)

// Task priorities.
const (
	PriorityLow    = "LOW"
	PriorityMedium = "MEDIUM"
	PriorityHigh   = "HIGH"
	PriorityUrgent = "URGENT"
)

// ValidStatus reports whether s is a known task status.
func ValidStatus(s string) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusReview, StatusDone:
		return true
	}
	return false
}

// ValidPriority reports whether p is a known task priority.
func ValidPriority(p string) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

// Task is a unit of work on the board.
type Task struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	CreatedByID  string     `json:"createdById"`
	AssignedToID *string    `json:"assignedToId,omitempty"`
	DueDate      *time.Time `json:"dueDate,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// TaskDetails is a task with its related users resolved.
type TaskDetails struct {
	Task
	CreatedBy  *User `json:"createdBy"`
	AssignedTo *User `json:"assignedTo,omitempty"`
}

// TaskFilter narrows a task listing. Empty fields are ignored.
type TaskFilter struct {
	Status       string `json:"status,omitempty"`
	Priority     string `json:"priority,omitempty"`
	AssignedToID string `json:"assignedToId,omitempty"`
	CreatedByID  string `json:"createdById,omitempty"`
}

// TaskUpdate is a partial update. Nil fields are left unchanged.
// ClearAssignee / ClearDueDate / ClearDescription set the column to NULL.
type TaskUpdate struct {
	Title            *string
	Description      *string
	ClearDescription bool
	Status           *string
	Priority         *string
	AssignedToID     *string
	ClearAssignee    bool
	DueDate          *time.Time
	ClearDueDate     bool
}

// Empty reports whether the update changes nothing.
func (u TaskUpdate) Empty() bool {
	return u.Title == nil && u.Description == nil && !u.ClearDescription && u.Status == nil &&
		u.Priority == nil && u.AssignedToID == nil && !u.ClearAssignee && u.DueDate == nil && !u.ClearDueDate
}

// UserUpdate is a partial profile update.
type UserUpdate struct {
	Name   *string
	Avatar *string
}
