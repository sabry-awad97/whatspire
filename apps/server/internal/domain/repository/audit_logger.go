package repository

import (
	"context"
	"time"
)

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	LogAPIKeyUsage(ctx context.Context, event APIKeyUsageEvent)
	LogSessionAction(ctx context.Context, event SessionActionEvent)
	LogMessageSent(ctx context.Context, event MessageSentEvent)
	LogAuthFailure(ctx context.Context, event AuthFailureEvent)
	LogWebhookDelivery(ctx context.Context, event WebhookDeliveryEvent)
}

// APIKeyUsageEvent represents an API key usage event
type APIKeyUsageEvent struct {
	APIKeyID  string
	Endpoint  string
	Method    string
	Timestamp time.Time
	IPAddress string
}

// SessionActionEvent represents a session action event
type SessionActionEvent struct {
	SessionID string
	Action    string // created, deleted, connected, disconnected
	APIKeyID  string
	Timestamp time.Time
}

// MessageSentEvent represents a message sent event
type MessageSentEvent struct {
	SessionID   string
	Recipient   string
	MessageType string
	Timestamp   time.Time
}

// AuthFailureEvent represents an authentication failure event
type AuthFailureEvent struct {
	APIKey    string
	Endpoint  string
	Reason    string
	Timestamp time.Time
	IPAddress string
}

// WebhookDeliveryEvent represents a webhook delivery event
type WebhookDeliveryEvent struct {
	WebhookURL string
	EventType  string
	StatusCode int
	Success    bool
	Error      *string
	Timestamp  time.Time
}
