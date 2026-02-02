package persistence

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"whatspire/internal/domain/repository"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry in the database
type AuditLog struct {
	ID        string
	EventType string
	APIKeyID  *string
	SessionID *string
	Endpoint  *string
	Action    *string
	Details   string // JSON
	IPAddress *string
	CreatedAt time.Time
}

// InMemoryAuditLogRepository implements audit log persistence with in-memory storage
type InMemoryAuditLogRepository struct {
	logs map[string]*AuditLog
	mu   sync.RWMutex
}

// NewInMemoryAuditLogRepository creates a new in-memory audit log repository
func NewInMemoryAuditLogRepository() *InMemoryAuditLogRepository {
	return &InMemoryAuditLogRepository{
		logs: make(map[string]*AuditLog),
	}
}

// SaveAPIKeyUsage saves an API key usage event
func (r *InMemoryAuditLogRepository) SaveAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"endpoint":   event.Endpoint,
		"method":     event.Method,
		"ip_address": event.IPAddress,
	})
	if err != nil {
		return err
	}

	log := &AuditLog{
		ID:        uuid.New().String(),
		EventType: "api_key_usage",
		APIKeyID:  &event.APIKeyID,
		Endpoint:  &event.Endpoint,
		Details:   string(details),
		IPAddress: &event.IPAddress,
		CreatedAt: event.Timestamp,
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = log
	return nil
}

// SaveSessionAction saves a session action event
func (r *InMemoryAuditLogRepository) SaveSessionAction(ctx context.Context, event repository.SessionActionEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"session_id": event.SessionID,
		"action":     event.Action,
		"api_key_id": event.APIKeyID,
	})
	if err != nil {
		return err
	}

	log := &AuditLog{
		ID:        uuid.New().String(),
		EventType: "session_action",
		APIKeyID:  &event.APIKeyID,
		SessionID: &event.SessionID,
		Action:    &event.Action,
		Details:   string(details),
		CreatedAt: event.Timestamp,
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = log
	return nil
}

// SaveMessageSent saves a message sent event
func (r *InMemoryAuditLogRepository) SaveMessageSent(ctx context.Context, event repository.MessageSentEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"session_id":   event.SessionID,
		"recipient":    event.Recipient,
		"message_type": event.MessageType,
	})
	if err != nil {
		return err
	}

	log := &AuditLog{
		ID:        uuid.New().String(),
		EventType: "message_sent",
		SessionID: &event.SessionID,
		Details:   string(details),
		CreatedAt: event.Timestamp,
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = log
	return nil
}

// SaveAuthFailure saves an authentication failure event
func (r *InMemoryAuditLogRepository) SaveAuthFailure(ctx context.Context, event repository.AuthFailureEvent) error {
	details, err := json.Marshal(map[string]interface{}{
		"api_key":    event.APIKey,
		"endpoint":   event.Endpoint,
		"reason":     event.Reason,
		"ip_address": event.IPAddress,
	})
	if err != nil {
		return err
	}

	log := &AuditLog{
		ID:        uuid.New().String(),
		EventType: "auth_failure",
		Endpoint:  &event.Endpoint,
		Details:   string(details),
		IPAddress: &event.IPAddress,
		CreatedAt: event.Timestamp,
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = log
	return nil
}

// SaveWebhookDelivery saves a webhook delivery event
func (r *InMemoryAuditLogRepository) SaveWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) error {
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
		return err
	}

	log := &AuditLog{
		ID:        uuid.New().String(),
		EventType: "webhook_delivery",
		Details:   string(details),
		CreatedAt: event.Timestamp,
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = log
	return nil
}

// FindByEventType retrieves audit logs by event type
func (r *InMemoryAuditLogRepository) FindByEventType(ctx context.Context, eventType string, limit, offset int) ([]*AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*AuditLog
	for _, log := range r.logs {
		if log.EventType == eventType {
			logCopy := *log
			filtered = append(filtered, &logCopy)
		}
	}

	// Apply pagination
	start := offset
	if start >= len(filtered) {
		return []*AuditLog{}, nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// FindByAPIKeyID retrieves audit logs by API key ID
func (r *InMemoryAuditLogRepository) FindByAPIKeyID(ctx context.Context, apiKeyID string, limit, offset int) ([]*AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*AuditLog
	for _, log := range r.logs {
		if log.APIKeyID != nil && *log.APIKeyID == apiKeyID {
			logCopy := *log
			filtered = append(filtered, &logCopy)
		}
	}

	// Apply pagination
	start := offset
	if start >= len(filtered) {
		return []*AuditLog{}, nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// Clear removes all audit logs (for testing)
func (r *InMemoryAuditLogRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = make(map[string]*AuditLog)
}
