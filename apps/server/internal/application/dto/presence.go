package dto

import "time"

// SendPresenceRequest represents a request to send presence update
type SendPresenceRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	ChatJID   string `json:"chat_jid" validate:"required"`
	State     string `json:"state" validate:"required,oneof=typing paused online offline"`
}

// PresenceResponse represents the response after sending presence
type PresenceResponse struct {
	ChatJID   string    `json:"chat_jid"`
	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
}
