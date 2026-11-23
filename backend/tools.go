//go:build tools

// Package tools pins code-generation tools in go.mod so `go run` uses the same version everywhere.
package tools

import _ "github.com/99designs/gqlgen"
