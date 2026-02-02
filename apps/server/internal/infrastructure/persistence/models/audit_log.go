package models

import (
	"time"
)

// AuditLog represents an audit log entry in the database
type AuditLog struct {
	ID        string    `gorm:"primaryKey;type:text;not null"`
	EventType string    `gorm:"type:text;not null;index:idx_audit_logs_event_type"`
	APIKeyID  *string   `gorm:"type:text;index:idx_audit_logs_api_key_id"`
	SessionID *string   `gorm:"type:text"`
	Endpoint  *string   `gorm:"type:text"`
	Action    *string   `gorm:"type:text"`
	Details   string    `gorm:"type:text"` // JSON
	IPAddress *string   `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;index:idx_audit_logs_created_at"`
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}
