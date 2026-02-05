package models

import (
	"time"
)

// Presence represents a presence update in the database
type Presence struct {
	ID        string    `gorm:"column:id;primaryKey;type:text;not null"`
	SessionID string    `gorm:"column:session_id;type:text;not null;index:idx_presence_session_id"`
	UserJID   string    `gorm:"column:user_jid;type:text;not null;index:idx_presence_user_jid"`
	ChatJID   string    `gorm:"column:chat_jid;type:text;not null"`
	State     string    `gorm:"column:state;type:text;not null;check:state IN ('typing', 'paused', 'online', 'offline')"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName specifies the table name for Presence model
func (Presence) TableName() string {
	return "presence"
}
