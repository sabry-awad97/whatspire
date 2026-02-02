package models

import (
	"time"
)

// Session represents a WhatsApp session in the database
type Session struct {
	ID        string    `gorm:"primaryKey;type:text;not null"`
	Name      string    `gorm:"type:text;not null"`
	JID       string    `gorm:"type:text;index:idx_sessions_jid"`
	Status    string    `gorm:"type:text;not null;check:status IN ('disconnected', 'connecting', 'connected', 'qr_pending', 'authenticating', 'authenticated', 'failed');index:idx_sessions_status"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}
