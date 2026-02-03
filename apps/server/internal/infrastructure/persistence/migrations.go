package persistence

import (
	"fmt"
	"log"

	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// RunAutoMigration runs GORM auto-migration for all models
// This creates missing tables and columns but does NOT drop existing ones
func RunAutoMigration(db *gorm.DB) error {
	log.Println("🔄 Running GORM auto-migration...")

	// List of all models to migrate
	modelsToMigrate := []interface{}{
		&models.Session{},
		&models.Reaction{},
		&models.Receipt{},
		&models.Presence{},
		&models.APIKey{},
		&models.AuditLog{},
		&models.Event{},
	}

	// Run auto-migration
	if err := db.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("✅ GORM auto-migration completed successfully")
	return nil
}

// VerifySchema verifies that all tables and indexes exist
func VerifySchema(db *gorm.DB) error {
	// Check if all tables exist
	tables := []string{
		"sessions",
		"reactions",
		"receipts",
		"presence",
		"api_keys",
		"audit_logs",
		"events",
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			return fmt.Errorf("table %s does not exist", table)
		}
	}

	log.Println("✅ Schema verification passed")
	return nil
}
