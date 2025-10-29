package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"taskboard/internal/models"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) (*models.Task, error) {
	task.ID = uuid.New().String()
	
	query := `
		INSERT INTO tasks (id, title, description, status, priority, created_by_id, assigned_to_id, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`
	
	err := r.db.QueryRow(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.CreatedByID, task.AssignedToID, task.DueDate,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	
	return task, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	
	query := `
		SELECT id, title, description, status, priority, created_by_id, 
		       assigned_to_id, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.CreatedByID, &task.AssignedToID, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	
	return &task, nil
}

func (r *TaskRepository) List(ctx context.Context, filter map[string]interface{}) ([]*models.Task, error) {
	query := `
		SELECT id, title, description, status, priority, created_by_id,
		       assigned_to_id, due_date, created_at, updated_at
		FROM tasks
		WHERE 1=1
	`
	
	args := []interface{}{}
	argPos := 1
	
	if status, ok := filter["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}
	
	if priority, ok := filter["priority"].(string); ok && priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argPos)
		args = append(args, priority)
		argPos++
	}
	
	if assignedToID, ok := filter["assigned_to_id"].(string); ok && assignedToID != "" {
		query += fmt.Sprintf(" AND assigned_to_id = $%d", argPos)
		args = append(args, assignedToID)
		argPos++
	}
	
	if createdByID, ok := filter["created_by_id"].(string); ok && createdByID != "" {
		query += fmt.Sprintf(" AND created_by_id = $%d", argPos)
		args = append(args, createdByID)
		argPos++
	}
	
	query += " ORDER BY created_at DESC"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()
	
	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
			&task.CreatedByID, &task.AssignedToID, &task.DueDate,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}
	
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.Task, error) {
	query := "UPDATE tasks SET updated_at = NOW()"
	args := []interface{}{}
	argPos := 1
	
	if title, ok := updates["title"].(string); ok {
		query += fmt.Sprintf(", title = $%d", argPos)
		args = append(args, title)
		argPos++
	}
	
	if description, ok := updates["description"]; ok {
		query += fmt.Sprintf(", description = $%d", argPos)
		args = append(args, description)
		argPos++
	}
	
	if status, ok := updates["status"].(string); ok {
		query += fmt.Sprintf(", status = $%d", argPos)
		args = append(args, status)
		argPos++
	}
	
	if priority, ok := updates["priority"].(string); ok {
		query += fmt.Sprintf(", priority = $%d", argPos)
		args = append(args, priority)
		argPos++
	}
	
	if assignedToID, ok := updates["assigned_to_id"]; ok {
		query += fmt.Sprintf(", assigned_to_id = $%d", argPos)
		args = append(args, assignedToID)
		argPos++
	}
	
	if dueDate, ok := updates["due_date"]; ok {
		query += fmt.Sprintf(", due_date = $%d", argPos)
		args = append(args, dueDate)
		argPos++
	}
	
	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	
	return r.GetByID(ctx, id)
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM tasks WHERE id = $1"
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	
	return nil
}

func (r *TaskRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Task, error) {
	return r.List(ctx, map[string]interface{}{
		"created_by_id": userID,
	})
}

func (r *TaskRepository) GetAssignedToUser(ctx context.Context, userID string) ([]*models.Task, error) {
	return r.List(ctx, map[string]interface{}{
		"assigned_to_id": userID,
	})
}