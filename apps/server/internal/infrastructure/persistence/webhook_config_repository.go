package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// WebhookConfigRepository implements WebhookConfigRepository with GORM
type WebhookConfigRepository struct {
	db *gorm.DB
}

// NewWebhookConfigRepository creates a new GORM webhook config repository
func NewWebhookConfigRepository(db *gorm.DB) *WebhookConfigRepository {
	return &WebhookConfigRepository{db: db}
}

// Create creates a new webhook configuration
func (r *WebhookConfigRepository) Create(ctx context.Context, config *entity.WebhookConfig) error {
	// Marshal events to JSON
	eventsJSON, err := json.Marshal(config.Events)
	if err != nil {
		return domainErrors.ErrDatabase.WithCause(err)
	}

	model := &models.WebhookConfig{
		ID:               config.ID,
		SessionID:        config.SessionID,
		Enabled:          config.Enabled,
		URL:              config.URL,
		Secret:           config.Secret,
		Events:           string(eventsJSON),
		IgnoreGroups:     config.IgnoreGroups,
		IgnoreBroadcasts: config.IgnoreBroadcasts,
		IgnoreChannels:   config.IgnoreChannels,
		CreatedAt:        config.CreatedAt,
		UpdatedAt:        config.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return domainErrors.ErrDatabase.WithMessage("webhook configuration already exists for this session")
		}
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return nil
}

// GetBySessionID retrieves webhook configuration for a session
func (r *WebhookConfigRepository) GetBySessionID(ctx context.Context, sessionID string) (*entity.WebhookConfig, error) {
	var model models.WebhookConfig

	result := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrNotFound.WithMessage("webhook configuration not found")
		}
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Unmarshal events from JSON
	var events []string
	if model.Events != "" {
		if err := json.Unmarshal([]byte(model.Events), &events); err != nil {
			return nil, domainErrors.ErrDatabase.WithCause(err)
		}
	}

	config := &entity.WebhookConfig{
		ID:               model.ID,
		SessionID:        model.SessionID,
		Enabled:          model.Enabled,
		URL:              model.URL,
		Secret:           model.Secret,
		Events:           events,
		IgnoreGroups:     model.IgnoreGroups,
		IgnoreBroadcasts: model.IgnoreBroadcasts,
		IgnoreChannels:   model.IgnoreChannels,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}

	return config, nil
}

// Update updates an existing webhook configuration
func (r *WebhookConfigRepository) Update(ctx context.Context, config *entity.WebhookConfig) error {
	// Marshal events to JSON
	eventsJSON, err := json.Marshal(config.Events)
	if err != nil {
		return domainErrors.ErrDatabase.WithCause(err)
	}

	updates := map[string]interface{}{
		"enabled":           config.Enabled,
		"url":               config.URL,
		"secret":            config.Secret,
		"events":            string(eventsJSON),
		"ignore_groups":     config.IgnoreGroups,
		"ignore_broadcasts": config.IgnoreBroadcasts,
		"ignore_channels":   config.IgnoreChannels,
		"updated_at":        time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&models.WebhookConfig{}).
		Where("session_id = ?", config.SessionID).
		Updates(updates)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound.WithMessage("webhook configuration not found")
	}

	return nil
}

// Delete removes a webhook configuration by session ID
func (r *WebhookConfigRepository) Delete(ctx context.Context, sessionID string) error {
	result := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&models.WebhookConfig{})

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound.WithMessage("webhook configuration not found")
	}

	return nil
}

// Exists checks if a webhook configuration exists for a session
func (r *WebhookConfigRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.WebhookConfig{}).
		Where("session_id = ?", sessionID).
		Count(&count)

	if result.Error != nil {
		return false, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return count > 0, nil
}
