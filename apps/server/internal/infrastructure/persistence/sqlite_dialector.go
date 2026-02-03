package persistence

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteDialector implements DatabaseDialector for SQLite
type SQLiteDialector struct{}

// NewSQLiteDialector creates a new SQLite dialector
func NewSQLiteDialector() *SQLiteDialector {
	return &SQLiteDialector{}
}

// Name returns the name of the database driver
func (d *SQLiteDialector) Name() string {
	return "sqlite"
}

// Initialize initializes the SQLite database connection with the given DSN
func (d *SQLiteDialector) Initialize(dsn string) (gorm.Dialector, error) {
	if dsn == "" {
		return nil, fmt.Errorf("SQLite DSN cannot be empty")
	}

	// SQLite-specific optimizations can be added to DSN
	// Example: file:test.db?cache=shared&mode=rwc&_journal_mode=WAL
	return sqlite.Open(dsn), nil
}
