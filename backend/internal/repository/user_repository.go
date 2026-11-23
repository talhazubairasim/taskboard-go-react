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

const userColumns = `id, email, password_hash, name, avatar, created_at, updated_at`

// UserRepository stores users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository builds a UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Avatar, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

// Create inserts a user, assigning a new ID. Emails are stored lower-cased.
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = uuid.NewString()
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash, name, avatar) VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at, updated_at`,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Avatar,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", translate(err))
	}
	return user, nil
}

// GetByID loads a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

// GetByEmail loads a user by email (case-insensitive).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(email))))
}

// GetByIDs loads several users at once, keyed by ID.
func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) (map[string]*models.User, error) {
	out := make(map[string]*models.User, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM users WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", translate(err))
	}
	defer rows.Close()
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out[u.ID] = u
	}
	return out, rows.Err()
}

// List returns all users, newest first.
func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := []*models.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Update applies a partial profile update and returns the updated user.
func (r *UserRepository) Update(ctx context.Context, id string, upd models.UserUpdate) (*models.User, error) {
	return scanUser(r.db.QueryRow(ctx,
		`UPDATE users SET name = COALESCE($2, name), avatar = COALESCE($3, avatar)
		 WHERE id = $1 RETURNING `+userColumns,
		id, upd.Name, upd.Avatar))
}

// EmailExists reports whether an account uses this email.
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		strings.ToLower(strings.TrimSpace(email))).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email: %w", err)
	}
	return exists, nil
}
