package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() error {
	connStr := "postgres://postgres:password@db:5432/taskboard?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return err
	}
	// simple ping loop until db ready (useful with docker-compose)
	for i := 0; i < 8; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = pool.Ping(ctx)
		cancel()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return err
	}
	Pool = pool

	// create table if not exists
	_, err = Pool.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'todo'
	);`)
	if err != nil {
		log.Println("error creating tasks table:", err)
		return err
	}

	fmt.Println("connected to db")
	return nil
}
