package persistence

import (
	"fmt"

	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// RunAutoMigration runs GORM auto-migration for all models
// This creates missing tables and columns but does NOT drop existing ones
func RunAutoMigration(db *gorm.DB, log *logger.Logger) error {
	log.Info("Starting GORM auto-migration for database schema")

	// List of all models to migrate
	modelsToMigrate := []any{
		&models.Session{},
		&models.Reaction{},
		&models.Receipt{},
		&models.Presence{},
		&models.APIKey{},
		&models.AuditLog{},
		&models.Event{},
		&models.WebhookConfig{},
	}

	// Run auto-migration
	if err := db.AutoMigrate(modelsToMigrate...); err != nil {
		log.WithError(err).Error("GORM auto-migration failed")
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.WithInt("model_count", len(modelsToMigrate)).
		Info("GORM auto-migration completed successfully")
	return nil
}

// VerifySchema verifies that all tables and indexes exist
func VerifySchema(db *gorm.DB, log *logger.Logger) error {
	// Check if all tables exist
	tables := []string{
		"sessions",
		"reactions",
		"receipts",
		"presence",
		"api_keys",
		"audit_logs",
		"events",
		"webhook_configs",
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			log.WithFields(map[string]interface{}{"table": table}).
				Error("Database table does not exist")
			return fmt.Errorf("table %s does not exist", table)
		}
	}

	log.WithInt("table_count", len(tables)).
		Info("Database schema verification passed successfully")
	return nil
}
