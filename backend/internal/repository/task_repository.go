package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskboard/internal/models"
)

const taskColumns = `id, title, description, status, priority, created_by_id,
	assigned_to_id, due_date, created_at, updated_at`

// TaskRepository stores tasks.
type TaskRepository struct {
	db *pgxpool.Pool
}

// NewTaskRepository builds a TaskRepository.
func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func scanTask(row pgx.Row) (*models.Task, error) {
	var t models.Task
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.CreatedByID, &t.AssignedToID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, translate(err)
	}
	return &t, nil
}

// Create inserts a task, assigning a new ID.
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) (*models.Task, error) {
	task.ID = uuid.NewString()
	err := r.db.QueryRow(ctx,
		`INSERT INTO tasks (id, title, description, status, priority, created_by_id, assigned_to_id, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.CreatedByID, task.AssignedToID, task.DueDate,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", translate(err))
	}
	return task, nil
}

// GetByID loads one task.
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	return scanTask(r.db.QueryRow(ctx, `SELECT `+taskColumns+` FROM tasks WHERE id = $1`, id))
}

// List returns tasks matching the filter, newest first.
func (r *TaskRepository) List(ctx context.Context, f models.TaskFilter) ([]*models.Task, error) {
	var where []string
	var args []interface{}
	add := func(col, val string) {
		if val != "" {
			args = append(args, val)
			where = append(where, fmt.Sprintf("%s = $%d", col, len(args)))
		}
	}
	add("status", f.Status)
	add("priority", f.Priority)
	add("assigned_to_id", f.AssignedToID)
	add("created_by_id", f.CreatedByID)

	query := `SELECT ` + taskColumns + ` FROM tasks`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC, id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", translate(err))
	}
	defer rows.Close()
	tasks := []*models.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, translate(err)
	}
	return tasks, nil
}

// Update applies a partial update and returns the updated task.
func (r *TaskRepository) Update(ctx context.Context, id string, u models.TaskUpdate) (*models.Task, error) {
	sets := []string{}
	args := []interface{}{id}
	set := func(col string, val interface{}) {
		args = append(args, val)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if u.Title != nil {
		set("title", *u.Title)
	}
	if u.ClearDescription {
		sets = append(sets, "description = NULL")
	} else if u.Description != nil {
		set("description", *u.Description)
	}
	if u.Status != nil {
		set("status", *u.Status)
	}
	if u.Priority != nil {
		set("priority", *u.Priority)
	}
	if u.ClearAssignee {
		sets = append(sets, "assigned_to_id = NULL")
	} else if u.AssignedToID != nil {
		set("assigned_to_id", *u.AssignedToID)
	}
	if u.ClearDueDate {
		sets = append(sets, "due_date = NULL")
	} else if u.DueDate != nil {
		set("due_date", *u.DueDate)
	}
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}
	query := `UPDATE tasks SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 RETURNING ` + taskColumns
	t, err := scanTask(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Delete removes a task.
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return translate(err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
