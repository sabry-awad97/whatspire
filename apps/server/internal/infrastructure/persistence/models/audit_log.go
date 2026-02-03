package models

import (
	"time"
)

// AuditLog represents an audit log entry in the database
type AuditLog struct {
	ID        string    `gorm:"column:id;primaryKey;type:text;not null"`
	EventType string    `gorm:"column:event_type;type:text;not null;index:idx_audit_logs_event_type"`
	APIKeyID  *string   `gorm:"column:api_key_id;type:text;index:idx_audit_logs_api_key_id"`
	SessionID *string   `gorm:"column:session_id;type:text"`
	Endpoint  *string   `gorm:"column:endpoint;type:text"`
	Action    *string   `gorm:"column:action;type:text"`
	Details   string    `gorm:"column:details;type:text"` // JSON
	IPAddress *string   `gorm:"column:ip_address;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;not null;index:idx_audit_logs_created_at"`
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}
