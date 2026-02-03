package persistence

import (
	"fmt"
	"time"

	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseFactory creates database connections based on configuration
type DatabaseFactory struct {
	dialectors map[string]repository.DatabaseDialector
}

// NewDatabaseFactory creates a new database factory with registered dialectors
func NewDatabaseFactory() *DatabaseFactory {
	factory := &DatabaseFactory{
		dialectors: make(map[string]repository.DatabaseDialector),
	}

	// Register built-in dialectors
	factory.RegisterDialector(NewSQLiteDialector())
	factory.RegisterDialector(NewPostgreSQLDialector())

	return factory
}

// RegisterDialector registers a new database dialector
func (f *DatabaseFactory) RegisterDialector(dialector repository.DatabaseDialector) {
	f.dialectors[dialector.Name()] = dialector
}

// CreateDatabase creates a new database connection based on the configuration
func (f *DatabaseFactory) CreateDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// Get the appropriate dialector
	dialector, exists := f.dialectors[cfg.Driver]
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Initialize the dialector with DSN
	gormDialector, err := dialector.Initialize(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize %s dialector: %w", cfg.Driver, err)
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(GetLogLevel(cfg.LogLevel))

	// Open database connection with optimized settings
	db, err := gorm.Open(gormDialector, &gorm.Config{
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
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
