package models

import (
	"time"
)

// APIKey represents an API key in the database
type APIKey struct {
	ID         string     `gorm:"column:id;primaryKey;type:text;not null"`
	KeyHash    string     `gorm:"column:key_hash;type:text;not null;uniqueIndex:idx_api_keys_key_hash"`
	Role       string     `gorm:"column:role;type:text;not null;check:role IN ('read', 'write', 'admin')"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	LastUsedAt *time.Time `gorm:"column:last_used_at;type:timestamp"`
	IsActive   bool       `gorm:"column:is_active;not null;default:true"`
}

// TableName specifies the table name for APIKey model
func (APIKey) TableName() string {
	return "api_keys"
}
