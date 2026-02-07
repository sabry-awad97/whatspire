package entity

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

// WebhookConfig represents a per-session webhook configuration
type WebhookConfig struct {
	ID        string   `json:"id"`
	SessionID string   `json:"session_id"`
	Enabled   bool     `json:"enabled"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"`
	Events    []string `json:"events"` // Event types to deliver

	// Message filtering options
	IgnoreGroups     bool `json:"ignore_groups"`
	IgnoreBroadcasts bool `json:"ignore_broadcasts"`
	IgnoreChannels   bool `json:"ignore_channels"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewWebhookConfig creates a new webhook configuration for a session
func NewWebhookConfig(id, sessionID string) *WebhookConfig {
	now := time.Now()
	return &WebhookConfig{
		ID:        id,
		SessionID: sessionID,
		Enabled:   false,
		Events:    []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GenerateSecret generates a cryptographically secure random secret
func (w *WebhookConfig) GenerateSecret() error {
	// Generate 32 bytes (256 bits) of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}

	// Encode as hex string
	w.Secret = hex.EncodeToString(bytes)
	w.UpdatedAt = time.Now()
	return nil
}

// Update updates the webhook configuration fields
func (w *WebhookConfig) Update(enabled bool, url string, events []string, ignoreGroups, ignoreBroadcasts, ignoreChannels bool) {
	w.Enabled = enabled
	w.URL = url
	w.Events = events
	w.IgnoreGroups = ignoreGroups
	w.IgnoreBroadcasts = ignoreBroadcasts
	w.IgnoreChannels = ignoreChannels
	w.UpdatedAt = time.Now()
}

// ShouldDeliverEvent checks if an event should be delivered based on configuration
func (w *WebhookConfig) ShouldDeliverEvent(eventType EventType) bool {
	// If not enabled, don't deliver
	if !w.Enabled {
		return false
	}

	// If no events configured, deliver all
	if len(w.Events) == 0 {
		return true
	}

	// Check if event type is in the configured list
	for _, configuredEvent := range w.Events {
		if configuredEvent == string(eventType) {
			return true
		}
	}

	return false
}

// MarshalJSON implements json.Marshaler
func (w *WebhookConfig) MarshalJSON() ([]byte, error) {
	type Alias WebhookConfig
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(w),
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	})
}
