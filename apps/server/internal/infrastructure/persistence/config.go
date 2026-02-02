package persistence

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds GORM database configuration
type DBConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
}

// DefaultDBConfig returns default GORM configuration
func DefaultDBConfig() DBConfig {
	return DBConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        logger.Warn,
	}
}

// NewDB creates a new GORM database connection
func NewDB(config DBConfig) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(config.LogLevel)

	// Open database connection with optimized settings
	db, err := gorm.Open(sqlite.Open(config.DSN), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Better performance - no automatic transactions
		PrepareStmt:            true, // Prepared statement cache for better performance
		NowFunc: func() time.Time { // Use UTC for consistency
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// GetLogLevel converts string log level to GORM log level
func GetLogLevel(level string) logger.LogLevel {
	switch level {
	case "debug":
		return logger.Info // GORM Info level shows SQL queries
	case "info":
		return logger.Warn
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Warn
	}
}
