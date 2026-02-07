package models

import (
	"time"
)

// WebhookConfig represents a per-session webhook configuration in the database
type WebhookConfig struct {
	ID        string `gorm:"column:id;primaryKey;type:text;not null"`
	SessionID string `gorm:"column:session_id;type:text;not null;uniqueIndex:idx_webhook_configs_session_id"`
	Enabled   bool   `gorm:"column:enabled;not null;default:false"`
	URL       string `gorm:"column:url;type:text"`
	Secret    string `gorm:"column:secret;type:text"` // HMAC secret for signature verification
	Events    string `gorm:"column:events;type:text"` // JSON array of event types to deliver

	// Message filtering options
	IgnoreGroups     bool `gorm:"column:ignore_groups;not null;default:false"`
	IgnoreBroadcasts bool `gorm:"column:ignore_broadcasts;not null;default:false"`
	IgnoreChannels   bool `gorm:"column:ignore_channels;not null;default:false"`

	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName specifies the table name for WebhookConfig model
func (WebhookConfig) TableName() string {
	return "webhook_configs"
}
