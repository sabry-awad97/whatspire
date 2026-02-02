package persistence

import (
	"context"
	"encoding/json"

	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/persistence/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogRepository implements audit log persistence with GORM
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new GORM audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// SaveAPIKeyUsage saves an API key usage event
func (r *AuditLogRepository) SaveAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"endpoint":   event.Endpoint,
		"method":     event.Method,
		"ip_address": event.IPAddress,
	})
	if err != nil {
		return domainErrors.ErrDatabaseError.WithCause(err)
	}

	model := &models.AuditLog{
		ID:        uuid.New().String(),
		EventType: "api_key_usage",
		APIKeyID:  &event.APIKeyID,
		Endpoint:  &event.Endpoint,
		Details:   string(details),
		IPAddress: &event.IPAddress,
		CreatedAt: event.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// SaveSessionAction saves a session action event
func (r *AuditLogRepository) SaveSessionAction(ctx context.Context, event repository.SessionActionEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"session_id": event.SessionID,
		"action":     event.Action,
		"api_key_id": event.APIKeyID,
	})
	if err != nil {
		return domainErrors.ErrDatabaseError.WithCause(err)
	}

	model := &models.AuditLog{
		ID:        uuid.New().String(),
		EventType: "session_action",
		APIKeyID:  &event.APIKeyID,
		SessionID: &event.SessionID,
		Action:    &event.Action,
		Details:   string(details),
		CreatedAt: event.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// SaveMessageSent saves a message sent event
func (r *AuditLogRepository) SaveMessageSent(ctx context.Context, event repository.MessageSentEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"session_id":   event.SessionID,
		"recipient":    event.Recipient,
		"message_type": event.MessageType,
	})
	if err != nil {
		return domainErrors.ErrDatabaseError.WithCause(err)
	}

	model := &models.AuditLog{
		ID:        uuid.New().String(),
		EventType: "message_sent",
		SessionID: &event.SessionID,
		Details:   string(details),
		CreatedAt: event.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// SaveAuthFailure saves an authentication failure event
func (r *AuditLogRepository) SaveAuthFailure(ctx context.Context, event repository.AuthFailureEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"api_key":    event.APIKey,
		"endpoint":   event.Endpoint,
		"reason":     event.Reason,
		"ip_address": event.IPAddress,
	})
	if err != nil {
		return domainErrors.ErrDatabaseError.WithCause(err)
	}

	model := &models.AuditLog{
		ID:        uuid.New().String(),
		EventType: "auth_failure",
		Endpoint:  &event.Endpoint,
		Details:   string(details),
		IPAddress: &event.IPAddress,
		CreatedAt: event.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// SaveWebhookDelivery saves a webhook delivery event
func (r *AuditLogRepository) SaveWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) error {
	detailsMap := map[string]interface{}{
		"webhook_url": event.WebhookURL,
		"event_type":  event.EventType,
		"status_code": event.StatusCode,
		"success":     event.Success,
	}
	if event.Error != nil {
		detailsMap["error"] = *event.Error
	}

	details, err := json.Marshal(detailsMap)
	if err != nil {
		return domainErrors.ErrDatabaseError.WithCause(err)
	}

	model := &models.AuditLog{
		ID:        uuid.New().String(),
		EventType: "webhook_delivery",
		Details:   string(details),
		CreatedAt: event.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// FindByEventType retrieves audit logs by event type with pagination
func (r *AuditLogRepository) FindByEventType(ctx context.Context, eventType string, limit, offset int) ([]*models.AuditLog, error) {
	var modelLogs []*models.AuditLog

	result := r.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelLogs)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return modelLogs, nil
}
