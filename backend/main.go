package main

import (
	"log"
	"net/http"
	"os"
	"taskboard-go-react/backend/db"
	"taskboard-go-react/backend/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := db.Connect(); err != nil {
		log.Fatal("db connect error:", err)
	}

	router := chi.NewRouter()

	// GraphQL server (gqlgen)
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	// optionally mount playground at /
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
