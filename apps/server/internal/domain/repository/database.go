package repository

import (
	"gorm.io/gorm"
)

// DatabaseDialector defines the interface for database dialectors
// This allows switching between SQLite (development) and PostgreSQL (production)
type DatabaseDialector interface {
	// Name returns the name of the database driver (e.g., "sqlite", "postgres")
	Name() string

	// Initialize initializes the database connection with the given DSN
	Initialize(dsn string) (gorm.Dialector, error)
}
