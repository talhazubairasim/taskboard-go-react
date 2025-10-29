package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"taskboard/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = uuid.New().String()
	
	query := `
		INSERT INTO users (id, email, password_hash, name)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`
	
	err := r.db.QueryRow(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	
	query := `
		SELECT id, email, password_hash, name, avatar, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Avatar, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	
	query := `
		SELECT id, email, password_hash, name, avatar, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Avatar, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.Name,
			&user.Avatar, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}
	
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.User, error) {
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{}
	argPos := 1
	
	if name, ok := updates["name"].(string); ok {
		query += fmt.Sprintf(", name = $%d", argPos)
		args = append(args, name)
		argPos++
	}
	
	if avatar, ok := updates["avatar"]; ok {
		query += fmt.Sprintf(", avatar = $%d", argPos)
		args = append(args, avatar)
		argPos++
	}
	
	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return r.GetByID(ctx, id)
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}
	
	return exists, nil
}