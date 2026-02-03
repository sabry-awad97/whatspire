package persistence

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgreSQLDialector implements DatabaseDialector for PostgreSQL
type PostgreSQLDialector struct{}

// NewPostgreSQLDialector creates a new PostgreSQL dialector
func NewPostgreSQLDialector() *PostgreSQLDialector {
	return &PostgreSQLDialector{}
}

// Name returns the name of the database driver
func (d *PostgreSQLDialector) Name() string {
	return "postgres"
}

// Initialize initializes the PostgreSQL database connection with the given DSN
func (d *PostgreSQLDialector) Initialize(dsn string) (gorm.Dialector, error) {
	if dsn == "" {
		return nil, fmt.Errorf("PostgreSQL DSN cannot be empty")
	}

	// PostgreSQL DSN format:
	// host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai
	// or connection string:
	// postgres://user:password@localhost:5432/dbname?sslmode=disable
	return postgres.Open(dsn), nil
}
