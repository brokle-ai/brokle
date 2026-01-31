//go:build tools
// +build tools

// Package tools tracks development tool dependencies.
// Import tools here to ensure they're tracked in go.mod.
// Install with: go install $(go list -f '{{join .Imports " "}}' tools.go)
package tools

import (
	_ "github.com/air-verse/air"
	_ "github.com/swaggo/swag/cmd/swag"
)
