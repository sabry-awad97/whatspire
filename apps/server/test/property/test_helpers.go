package property

import (
	"testing"

	"whatspire/internal/infrastructure/persistence"
	"whatspire/test/helpers"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"pgregory.net/rapid"
)

// setupTestDB creates an in-memory SQLite database for property testing
func setupTestDB(t testing.TB) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for property tests
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	err = persistence.RunAutoMigration(db, helpers.CreateTestLogger())
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// setupTestDBForRapid creates an in-memory SQLite database for rapid property testing
func setupTestDBForRapid(t *rapid.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for property tests
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	err = persistence.RunAutoMigration(db, helpers.CreateTestLogger())
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}
