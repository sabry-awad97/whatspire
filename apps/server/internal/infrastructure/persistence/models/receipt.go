package models

import (
	"time"
)

// Receipt represents a message receipt in the database
type Receipt struct {
	ID        string    `gorm:"primaryKey;type:text;not null"`
	MessageID string    `gorm:"type:text;not null;index:idx_receipts_message_id"`
	SessionID string    `gorm:"type:text;not null;index:idx_receipts_session_id"`
	FromJID   string    `gorm:"column:from_jid;type:text;not null"`
	ToJID     string    `gorm:"column:to_jid;type:text;not null"`
	Type      string    `gorm:"type:text;not null;check:type IN ('delivered', 'read')"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName specifies the table name for Receipt model
func (Receipt) TableName() string {
	return "receipts"
}
