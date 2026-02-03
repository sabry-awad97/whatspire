package persistence

import (
	"context"
	"fmt"
	"time"

	"whatspire/internal/domain/repository"

	"gorm.io/gorm"
)

// GORMMigrationRunner implements MigrationRunner using GORM auto-migration
type GORMMigrationRunner struct {
	db *gorm.DB
}

// NewGORMMigrationRunner creates a new GORM migration runner
func NewGORMMigrationRunner(db *gorm.DB) *GORMMigrationRunner {
	return &GORMMigrationRunner{db: db}
}

// Up runs all pending migrations using GORM AutoMigrate
func (r *GORMMigrationRunner) Up(ctx context.Context) error {
	// Run auto-migration for all models
	if err := RunAutoMigration(r.db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Verify schema after migration
	if err := VerifySchema(r.db); err != nil {
		return fmt.Errorf("schema verification failed: %w", err)
	}

	return nil
}

// Down rolls back the last migration
// Note: GORM doesn't support automatic rollback, so this is a no-op
func (r *GORMMigrationRunner) Down(ctx context.Context) error {
	return fmt.Errorf("GORM auto-migration does not support rollback - manual intervention required")
}

// Version returns the current migration version
// For GORM auto-migration, we return a timestamp-based version
func (r *GORMMigrationRunner) Version(ctx context.Context) (int, error) {
	// Check if migration_versions table exists
	var count int64
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM migration_versions").Scan(&count).Error; err != nil {
		// Table doesn't exist, return version 0
		return 0, nil
	}

	// Get the latest version
	var version int
	if err := r.db.WithContext(ctx).Raw("SELECT COALESCE(MAX(version), 0) FROM migration_versions").Scan(&version).Error; err != nil {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, nil
}

// Status returns the status of all migrations
func (r *GORMMigrationRunner) Status(ctx context.Context) ([]repository.MigrationStatus, error) {
	// Check if migration_versions table exists
	if !r.db.Migrator().HasTable("migration_versions") {
		return []repository.MigrationStatus{}, nil
	}

	// Get all migration records
	var records []MigrationVersion
	if err := r.db.WithContext(ctx).Order("version ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	// Convert to domain status
	statuses := make([]repository.MigrationStatus, len(records))
	for i, record := range records {
		appliedAt := record.AppliedAt.Format(time.RFC3339)
		statuses[i] = repository.MigrationStatus{
			Version:   record.Version,
			Name:      record.Name,
			AppliedAt: &appliedAt,
			Applied:   true,
		}
	}

	return statuses, nil
}

// MigrationVersion represents a migration version record in the database
type MigrationVersion struct {
	Version   int       `gorm:"primaryKey;autoIncrement:false"`
	Name      string    `gorm:"type:text;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for MigrationVersion model
func (MigrationVersion) TableName() string {
	return "migration_versions"
}

// RecordMigration records a migration in the database
func (r *GORMMigrationRunner) RecordMigration(ctx context.Context, version int, name string) error {
	// Ensure migration_versions table exists
	if err := r.db.AutoMigrate(&MigrationVersion{}); err != nil {
		return fmt.Errorf("failed to create migration_versions table: %w", err)
	}

	// Record the migration
	record := MigrationVersion{
		Version:   version,
		Name:      name,
		AppliedAt: time.Now().UTC(),
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}
