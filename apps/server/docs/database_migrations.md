# Database Migrations

This document describes the database migration system used in the Whatspire WhatsApp service.

## Overview

The service uses **GORM AutoMigrate** for database schema management, which automatically creates and updates tables based on Go struct definitions. This approach provides:

- **Automatic schema creation**: Tables and columns are created automatically
- **Safe updates**: Existing data is preserved during schema changes
- **Version tracking**: Migration versions are recorded for audit purposes
- **Multi-database support**: Works with both SQLite and PostgreSQL

## Migration Strategy

### GORM AutoMigrate

The service uses GORM's `AutoMigrate` feature, which:

1. Creates missing tables
2. Adds missing columns
3. Creates missing indexes
4. **Does NOT** drop existing tables or columns (safe by default)

### Version Tracking

Each migration run is recorded in the `migration_versions` table with:

- `version`: Unix timestamp of when the migration was applied
- `name`: Description of the migration (e.g., "auto_migration")
- `applied_at`: Timestamp of when the migration was applied

## Database Support

### SQLite (Development)

**Configuration:**

```bash
export WHATSAPP_DATABASE_DRIVER=sqlite
export WHATSAPP_DATABASE_DSN=/data/whatspire.db
```

**Features:**

- Single-file database
- No separate server required
- WAL mode for better concurrency
- Foreign key constraints enabled

### PostgreSQL (Production)

**Configuration:**

```bash
export WHATSAPP_DATABASE_DRIVER=postgres
export WHATSAPP_DATABASE_DSN="host=localhost user=whatspire password=secret dbname=whatspire port=5432 sslmode=disable"
```

**Docker Image:**

The service uses `pgvector/pgvector:pg18-trixie` which includes:

- PostgreSQL 18 (latest stable version)
- pgvector extension for vector similarity search
- Debian Trixie base for stability
- Future-ready for AI/ML features

**Features:**

- Production-grade reliability
- Better concurrency handling
- Advanced indexing capabilities
- Connection pooling
- Vector similarity search support (pgvector extension)

## Running Migrations

### Automatic (Recommended)

Migrations run automatically on application startup:

```bash
go run ./cmd/whatsapp/main.go
```

Output:

```
🔄 Running database migrations...
📊 Current migration version: 1707123456
✅ Migration recorded: version 1707123457
✅ Database migrations completed
```

### Manual

You can also run migrations manually using the migration runner:

```go
import (
    "context"
    "whatspire/internal/infrastructure/persistence"
)

// Create migration runner
runner := persistence.NewGORMMigrationRunner(db)

// Run migrations
if err := runner.Up(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Migration Status

### Check Current Version

```go
version, err := runner.Version(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Current version: %d\n", version)
```

### Check Migration History

```go
statuses, err := runner.Status(ctx)
if err != nil {
    log.Fatal(err)
}

for _, status := range statuses {
    fmt.Printf("Version %d: %s (applied at %s)\n",
        status.Version, status.Name, *status.AppliedAt)
}
```

## Schema Management

### Current Tables

The service manages the following tables:

1. **sessions**: WhatsApp session data
2. **reactions**: Message reactions
3. **receipts**: Read receipts
4. **presence**: User presence information
5. **api_keys**: API authentication keys
6. **audit_logs**: Audit trail for operations
7. **migration_versions**: Migration tracking

### Adding New Tables

To add a new table:

1. Create a model in `internal/infrastructure/persistence/models/`:

```go
package models

type MyNewTable struct {
    ID        string    `gorm:"primaryKey;type:text"`
    Name      string    `gorm:"type:text;not null"`
    CreatedAt time.Time `gorm:"not null"`
}

func (MyNewTable) TableName() string {
    return "my_new_table"
}
```

2. Add the model to `migrations.go`:

```go
modelsToMigrate := []interface{}{
    &models.Session{},
    // ... existing models ...
    &models.MyNewTable{}, // Add here
}
```

3. Restart the application - the table will be created automatically

## Rollback

### Important Note

GORM AutoMigrate **does not support automatic rollback**. This is by design to prevent accidental data loss.

### Manual Rollback

If you need to rollback a migration:

1. **Backup your database first**
2. Manually drop the affected tables/columns using SQL
3. Restart the application to re-run migrations

Example (PostgreSQL):

```sql
-- Backup
pg_dump whatspire > backup.sql

-- Drop table
DROP TABLE IF EXISTS my_new_table;

-- Restore if needed
psql whatspire < backup.sql
```

Example (SQLite):

```bash
# Backup
cp /data/whatspire.db /data/whatspire.db.backup

# Restore if needed
cp /data/whatspire.db.backup /data/whatspire.db
```

## Testing Migrations

### SQLite

```bash
export WHATSAPP_DATABASE_DRIVER=sqlite
export WHATSAPP_DATABASE_DSN=/tmp/test.db
go test ./test/...
```

### PostgreSQL

```bash
# Start PostgreSQL
docker-compose up -d

# Run tests
export WHATSAPP_DATABASE_DRIVER=postgres
export WHATSAPP_DATABASE_DSN="host=localhost user=whatspire password=whatspire_dev dbname=whatspire_test port=5433 sslmode=disable"
go test ./test/...
```

## Best Practices

1. **Always backup before migrations** in production
2. **Test migrations** with both SQLite and PostgreSQL
3. **Use transactions** for data migrations (not schema changes)
4. **Monitor migration logs** during deployment
5. **Keep models in sync** with database schema
6. **Document breaking changes** in migration comments

## Troubleshooting

### Migration Fails

If a migration fails:

1. Check the error message in logs
2. Verify database connectivity
3. Check database permissions
4. Ensure no conflicting schema changes
5. Review recent model changes

### Schema Mismatch

If you see schema mismatch errors:

1. Check `migration_versions` table for last successful migration
2. Compare model definitions with actual database schema
3. Run `VerifySchema()` to identify missing tables
4. Consider manual schema fixes if needed

### Connection Pool Exhaustion

If you see connection errors:

1. Check `max_open_conns` configuration
2. Monitor active connections
3. Increase pool size if needed
4. Check for connection leaks in code

## Configuration Reference

### Database Configuration

```yaml
database:
  driver: postgres # "sqlite" or "postgres"
  dsn: "host=localhost..." # Connection string
  max_idle_conns: 10 # Maximum idle connections
  max_open_conns: 100 # Maximum open connections
  conn_max_lifetime: 1h # Connection maximum lifetime
  log_level: warn # GORM log level
```

### Environment Variables

```bash
WHATSAPP_DATABASE_DRIVER=postgres
WHATSAPP_DATABASE_DSN="host=localhost user=whatspire password=secret dbname=whatspire port=5432 sslmode=disable"
WHATSAPP_DATABASE_MAX_IDLE_CONNS=10
WHATSAPP_DATABASE_MAX_OPEN_CONNS=100
WHATSAPP_DATABASE_CONN_MAX_LIFETIME=1h
WHATSAPP_DATABASE_LOG_LEVEL=warn
```

## See Also

- [Configuration Guide](configuration.md)
- [Deployment Guide](deployment_guide.md)
- [GORM Documentation](https://gorm.io/docs/)
