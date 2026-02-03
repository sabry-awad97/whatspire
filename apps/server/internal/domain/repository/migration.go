package repository

import (
	"context"
)

// MigrationRunner defines the interface for database migrations
type MigrationRunner interface {
	// Up runs all pending migrations
	Up(ctx context.Context) error

	// Down rolls back the last migration
	Down(ctx context.Context) error

	// Version returns the current migration version
	Version(ctx context.Context) (int, error)

	// Status returns the status of all migrations
	Status(ctx context.Context) ([]MigrationStatus, error)
}

// MigrationStatus represents the status of a single migration
type MigrationStatus struct {
	Version   int
	Name      string
	AppliedAt *string // nil if not applied
	Applied   bool
}
