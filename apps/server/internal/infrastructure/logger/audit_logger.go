package logger

import (
	"context"

	"whatspire/internal/domain/repository"
)

// AuditLogger implements the repository.AuditLogger interface using zerolog
type AuditLogger struct {
	logger *Logger
}

// NewAuditLogger creates a new AuditLogger
func NewAuditLogger(logger *Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
	}
}

// LogAPIKeyUsage logs an API key usage event
func (al *AuditLogger) LogAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) {
	al.logger.WithContext(ctx).
		WithStr("event_type", "api_key_usage").
		WithStr("api_key_id", event.APIKeyID).
		WithStr("endpoint", event.Endpoint).
		WithStr("method", event.Method).
		WithStr("ip_address", event.IPAddress).
		Info("API key used")
}

// LogAPIKeyCreated logs an API key creation event
func (al *AuditLogger) LogAPIKeyCreated(ctx context.Context, event repository.APIKeyCreatedEvent) {
	l := al.logger.WithContext(ctx).
		WithStr("event_type", "api_key_created").
		WithStr("api_key_id", event.APIKeyID).
		WithStr("role", event.Role).
		WithStr("created_by", event.CreatedBy)

	if event.Description != nil {
		l = l.WithStr("description", *event.Description)
	}

	l.Info("API key created")
}

// LogAPIKeyRevoked logs an API key revocation event
func (al *AuditLogger) LogAPIKeyRevoked(ctx context.Context, event repository.APIKeyRevokedEvent) {
	l := al.logger.WithContext(ctx).
		WithStr("event_type", "api_key_revoked").
		WithStr("api_key_id", event.APIKeyID).
		WithStr("revoked_by", event.RevokedBy)

	if event.RevocationReason != nil {
		l = l.WithStr("revocation_reason", *event.RevocationReason)
	}

	l.Warn("API key revoked")
}

// LogSessionAction logs a session action event
func (al *AuditLogger) LogSessionAction(ctx context.Context, event repository.SessionActionEvent) {
	al.logger.WithContext(ctx).
		WithStr("event_type", "session_action").
		WithStr("session_id", event.SessionID).
		WithStr("action", event.Action).
		WithStr("api_key_id", event.APIKeyID).
		Info("Session action")
}

// LogMessageSent logs a message sent event
func (al *AuditLogger) LogMessageSent(ctx context.Context, event repository.MessageSentEvent) {
	al.logger.WithContext(ctx).
		WithStr("event_type", "message_sent").
		WithStr("session_id", event.SessionID).
		WithStr("recipient", event.Recipient).
		WithStr("message_type", event.MessageType).
		Info("Message sent")
}

// LogAuthFailure logs an authentication failure event
func (al *AuditLogger) LogAuthFailure(ctx context.Context, event repository.AuthFailureEvent) {
	al.logger.WithContext(ctx).
		WithStr("event_type", "auth_failure").
		WithStr("api_key", event.APIKey).
		WithStr("endpoint", event.Endpoint).
		WithStr("reason", event.Reason).
		WithStr("ip_address", event.IPAddress).
		Warn("Authentication failed")
}

// LogWebhookDelivery logs a webhook delivery event
func (al *AuditLogger) LogWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) {
	l := al.logger.WithContext(ctx).
		WithStr("event_type", "webhook_delivery").
		WithStr("webhook_url", event.WebhookURL).
		WithStr("event_type_name", event.EventType).
		WithInt("status_code", event.StatusCode)

	if event.Error != nil {
		l = l.WithStr("error", *event.Error)
	}

	if event.Success {
		l.Info("Webhook delivered")
	} else {
		l.Error("Webhook delivery failed")
	}
}
