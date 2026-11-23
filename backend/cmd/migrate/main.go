// Command migrate applies or rolls back database migrations.
//
//	go run cmd/migrate/main.go up    # apply all pending migrations
//	go run cmd/migrate/main.go down  # roll back the latest migration
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"taskboard/internal/config"
	"taskboard/internal/database"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Fprintln(os.Stderr, "usage: migrate up|down")
		os.Exit(2)
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, config.Load().DatabaseURL, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if os.Args[1] == "up" {
		n, err := database.MigrateUp(ctx, pool)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("applied %d migration(s)\n", n)
		return
	}
	v, err := database.MigrateDown(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}
	if v == "" {
		fmt.Println("nothing to roll back")
	} else {
		fmt.Printf("rolled back %s\n", v)
	}
}
