// Package repository persists users and tasks in PostgreSQL.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrNotFound means no row matched.
	ErrNotFound = errors.New("not found")
	// ErrDuplicate means a unique constraint (e.g. email) was violated.
	ErrDuplicate = errors.New("already exists")
	// ErrInvalidReference means a foreign key (e.g. assignee) points nowhere.
	ErrInvalidReference = errors.New("referenced record does not exist")
)

// translate maps PostgreSQL error codes to repository errors.
func translate(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicate
		case "23503":
			return ErrInvalidReference
		case "22P02": // invalid input syntax, e.g. a non-UUID id
			return ErrNotFound
		}
	}
	return err
}
