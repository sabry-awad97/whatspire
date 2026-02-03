package models

import (
	"time"
)

// Reaction represents a message reaction in the database
type Reaction struct {
	ID        string    `gorm:"column:id;primaryKey;type:text;not null"`
	MessageID string    `gorm:"column:message_id;type:text;not null;index:idx_reactions_message_id"`
	SessionID string    `gorm:"column:session_id;type:text;not null;index:idx_reactions_session_id"`
	FromJID   string    `gorm:"column:from_jid;type:text;not null"`
	ToJID     string    `gorm:"column:to_jid;type:text;not null"`
	Emoji     string    `gorm:"column:emoji;type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName specifies the table name for Reaction model
func (Reaction) TableName() string {
	return "reactions"
}
