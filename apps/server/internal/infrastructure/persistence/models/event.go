package models

import (
	"time"
)

// Event represents a WhatsApp event in the database
type Event struct {
	ID        string    `gorm:"column:id;primaryKey;type:text;not null"`
	Type      string    `gorm:"column:type;type:text;not null;index:idx_events_type"`
	SessionID string    `gorm:"column:session_id;type:text;not null;index:idx_events_session_id"`
	Data      []byte    `gorm:"column:data;type:bytea"` // JSON data stored as bytes (works with both SQLite and PostgreSQL)
	Timestamp time.Time `gorm:"column:timestamp;not null;index:idx_events_timestamp"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName specifies the table name for Event model
func (Event) TableName() string {
	return "events"
}

// Indexes defines composite indexes for better query performance
// These are created automatically by GORM during migration
func (Event) Indexes() []interface{} {
	return []interface{}{
		// Composite index for session + timestamp queries (most common)
		"CREATE INDEX IF NOT EXISTS idx_events_session_timestamp ON events(session_id, timestamp DESC)",
		// Composite index for type + timestamp queries
		"CREATE INDEX IF NOT EXISTS idx_events_type_timestamp ON events(type, timestamp DESC)",
		// Composite index for session + type queries
		"CREATE INDEX IF NOT EXISTS idx_events_session_type ON events(session_id, type)",
	}
}
