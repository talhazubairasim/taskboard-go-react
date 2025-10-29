package graph

import (
	"context"
	"strconv"
	"taskboard-go-react/backend/auth"
	"taskboard-go-react/backend/db"
)

// Task is a minimal struct used by resolvers (gqlgen will also create models if generated)
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Resolver struct{}

// Tasks resolver - returns all tasks
func (r *Resolver) Tasks(ctx context.Context) ([]*Task, error) {
	rows, err := db.Pool.Query(ctx, "SELECT id, title, description, status FROM tasks ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

// CreateTask - insert and return created task
func (r *Resolver) CreateTask(ctx context.Context, title string, description *string) (*Task, error) {
	desc := ""
	if description != nil {
		desc = *description
	}
	var id int
	err := db.Pool.QueryRow(ctx, "INSERT INTO tasks (title, description, status) VALUES ($1,$2,$3) RETURNING id", title, desc, "todo").Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Task{ID: id, Title: title, Description: desc, Status: "todo"}, nil
}

// UpdateTask sets status
func (r *Resolver) UpdateTask(ctx context.Context, idStr string, status string) (*Task, error) {
	id, _ := strconv.Atoi(idStr)
	_, err := db.Pool.Exec(ctx, "UPDATE tasks SET status=$1 WHERE id=$2", status, id)
	if err != nil {
		return nil, err
	}
	row := db.Pool.QueryRow(ctx, "SELECT id, title, description, status FROM tasks WHERE id=$1", id)
	var t Task
	if err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status); err != nil {
		return nil, err
	}
	return &t, nil
}

// DeleteTask removes a task by ID
func (r *Resolver) DeleteTask(ctx context.Context, idStr string) (bool, error) {
	id, _ := strconv.Atoi(idStr)
	_, err := db.Pool.Exec(ctx, "DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Login - minimal login that returns a JWT token (accepts any username/password in scaffold).
func (r *Resolver) Login(ctx context.Context, username string, password string) (string, error) {
	// For the demo: accept any username/password. Replace with real auth later.
	token, err := auth.GenerateToken(username)
	if err != nil {
		return "", err
	}
	return token, nil
}
