package logger

import (
	"context"

	"whatspire/internal/domain/repository"
)

// AuditLogger implements the repository.AuditLogger interface using structured logging
type AuditLogger struct {
	logger Logger
}

// NewAuditLogger creates a new AuditLogger
func NewAuditLogger(logger Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
	}
}

// LogAPIKeyUsage logs an API key usage event
func (al *AuditLogger) LogAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) {
	al.logger.WithContext(ctx).Info("API key used",
		String("event_type", "api_key_usage"),
		String("api_key_id", event.APIKeyID),
		String("endpoint", event.Endpoint),
		String("method", event.Method),
		String("ip_address", event.IPAddress),
		Any("timestamp", event.Timestamp),
	)
}

// LogAPIKeyCreated logs an API key creation event
func (al *AuditLogger) LogAPIKeyCreated(ctx context.Context, event repository.APIKeyCreatedEvent) {
	fields := []Field{
		String("event_type", "api_key_created"),
		String("api_key_id", event.APIKeyID),
		String("role", event.Role),
		String("created_by", event.CreatedBy),
		Any("timestamp", event.Timestamp),
	}

	if event.Description != nil {
		fields = append(fields, String("description", *event.Description))
	}

	al.logger.WithContext(ctx).Info("API key created", fields...)
}

// LogAPIKeyRevoked logs an API key revocation event
func (al *AuditLogger) LogAPIKeyRevoked(ctx context.Context, event repository.APIKeyRevokedEvent) {
	fields := []Field{
		String("event_type", "api_key_revoked"),
		String("api_key_id", event.APIKeyID),
		String("revoked_by", event.RevokedBy),
		Any("timestamp", event.Timestamp),
	}

	if event.RevocationReason != nil {
		fields = append(fields, String("revocation_reason", *event.RevocationReason))
	}

	al.logger.WithContext(ctx).Warn("API key revoked", fields...)
}

// LogSessionAction logs a session action event
func (al *AuditLogger) LogSessionAction(ctx context.Context, event repository.SessionActionEvent) {
	al.logger.WithContext(ctx).Info("Session action",
		String("event_type", "session_action"),
		String("session_id", event.SessionID),
		String("action", event.Action),
		String("api_key_id", event.APIKeyID),
		Any("timestamp", event.Timestamp),
	)
}

// LogMessageSent logs a message sent event
func (al *AuditLogger) LogMessageSent(ctx context.Context, event repository.MessageSentEvent) {
	al.logger.WithContext(ctx).Info("Message sent",
		String("event_type", "message_sent"),
		String("session_id", event.SessionID),
		String("recipient", event.Recipient),
		String("message_type", event.MessageType),
		Any("timestamp", event.Timestamp),
	)
}

// LogAuthFailure logs an authentication failure event
func (al *AuditLogger) LogAuthFailure(ctx context.Context, event repository.AuthFailureEvent) {
	al.logger.WithContext(ctx).Warn("Authentication failed",
		String("event_type", "auth_failure"),
		String("api_key", event.APIKey),
		String("endpoint", event.Endpoint),
		String("reason", event.Reason),
		String("ip_address", event.IPAddress),
		Any("timestamp", event.Timestamp),
	)
}

// LogWebhookDelivery logs a webhook delivery event
func (al *AuditLogger) LogWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) {
	fields := []Field{
		String("event_type", "webhook_delivery"),
		String("webhook_url", event.WebhookURL),
		String("event_type_name", event.EventType),
		Int("status_code", event.StatusCode),
		Bool("success", event.Success),
		Any("timestamp", event.Timestamp),
	}

	if event.Error != nil {
		fields = append(fields, String("error", *event.Error))
	}

	if event.Success {
		al.logger.WithContext(ctx).Info("Webhook delivered", fields...)
	} else {
		al.logger.WithContext(ctx).Error("Webhook delivery failed", fields...)
	}
}
