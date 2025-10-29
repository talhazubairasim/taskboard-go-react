package models

import (
	"time"
)

type Task struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null" validate:"required,min=3,max=255"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'todo'" validate:"oneof=todo in-progress review done"`
	Priority    string    `json:"priority" gorm:"type:varchar(20);default:'medium'" validate:"oneof=low medium high urgent"`
	DueDate     time.Time `json:"due_date" gorm:"type:timestamp"`
	AssigneeID  uint      `json:"assignee_id" gorm:"index"`
	Assignee    User      `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	CreatedByID uint      `json:"created_by_id" gorm:"not null"`
	CreatedBy   User      `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type TaskResponse struct {
	ID          uint           `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Priority    string         `json:"priority"`
	DueDate     time.Time      `json:"due_date"`
	Assignee    UserResponse   `json:"assignee"`
	CreatedBy   UserResponse   `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string    `json:"title" validate:"required,min=3,max=255"`
	Description string    `json:"description"`
	Status      string    `json:"status" validate:"oneof=todo in-progress review done"`
	Priority    string    `json:"priority" validate:"oneof=low medium high urgent"`
	DueDate     time.Time `json:"due_date"`
	AssigneeID  uint      `json:"assignee_id" validate:"required"`
}

type UpdateTaskRequest struct {
	Title       string    `json:"title" validate:"omitempty,min=3,max=255"`
	Description string    `json:"description"`
	Status      string    `json:"status" validate:"omitempty,oneof=todo in-progress review done"`
	Priority    string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	DueDate     time.Time `json:"due_date"`
	AssigneeID  uint      `json:"assignee_id"`
}