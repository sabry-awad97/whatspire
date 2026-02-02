package models

import (
	"time"
)

// APIKey represents an API key in the database
type APIKey struct {
	ID         string     `gorm:"primaryKey;type:text;not null"`
	KeyHash    string     `gorm:"type:text;not null;uniqueIndex:idx_api_keys_key_hash"`
	Role       string     `gorm:"type:text;not null;check:role IN ('read', 'write', 'admin')"`
	CreatedAt  time.Time  `gorm:"not null"`
	LastUsedAt *time.Time `gorm:"type:timestamp"`
	IsActive   bool       `gorm:"not null;default:true"`
}

// TableName specifies the table name for APIKey model
func (APIKey) TableName() string {
	return "api_keys"
}
