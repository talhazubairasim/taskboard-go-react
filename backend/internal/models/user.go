package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	Avatar       *string   `json:"avatar" db:"avatar"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Task struct {
	ID           string     `json:"id" db:"id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description" db:"description"`
	Status       string     `json:"status" db:"status"`
	Priority     string     `json:"priority" db:"priority"`
	CreatedByID  string     `json:"created_by_id" db:"created_by_id"`
	AssignedToID *string    `json:"assigned_to_id" db:"assigned_to_id"`
	DueDate      *time.Time `json:"due_date" db:"due_date"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// For GraphQL relationships
type TaskWithRelations struct {
	Task
	CreatedBy  *User `json:"created_by"`
	AssignedTo *User `json:"assigned_to"`
}