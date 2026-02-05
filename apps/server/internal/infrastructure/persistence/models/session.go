package models

import (
	"time"
)

// Session represents a WhatsApp session in the database
type Session struct {
	ID        string    `gorm:"column:id;primaryKey;type:text;not null"`
	Name      string    `gorm:"column:name;type:text;not null"`
	JID       string    `gorm:"column:jid;type:text;index:idx_sessions_jid"`
	Status    string    `gorm:"column:status;type:text;not null;check:status IN ('disconnected', 'connecting', 'connected', 'qr_pending', 'authenticating', 'authenticated', 'failed', 'pending', 'logged_out');index:idx_sessions_status"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}
