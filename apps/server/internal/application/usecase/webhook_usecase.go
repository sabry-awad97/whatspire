package usecase

import (
	"context"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"github.com/google/uuid"
)

// WebhookUseCase handles webhook configuration operations
type WebhookUseCase struct {
	repo        repository.WebhookConfigRepository
	sessionRepo repository.SessionRepository
	auditLogger repository.AuditLogger
}

// NewWebhookUseCase creates a new WebhookUseCase
func NewWebhookUseCase(
	repo repository.WebhookConfigRepository,
	sessionRepo repository.SessionRepository,
	auditLogger repository.AuditLogger,
) *WebhookUseCase {
	return &WebhookUseCase{
		repo:        repo,
		sessionRepo: sessionRepo,
		auditLogger: auditLogger,
	}
}

// GetWebhookConfig retrieves webhook configuration for a session
func (uc *WebhookUseCase) GetWebhookConfig(ctx context.Context, sessionID string) (*entity.WebhookConfig, error) {
	// Verify session exists
	if _, err := uc.sessionRepo.GetByID(ctx, sessionID); err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrNotFound.WithMessage("session not found")
		}
		return nil, errors.ErrDatabase.WithCause(err)
	}

	// Get webhook config
	config, err := uc.repo.GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return default config if none exists
			return entity.NewWebhookConfig(uuid.New().String(), sessionID), nil
		}
		return nil, errors.ErrDatabase.WithCause(err)
	}

	return config, nil
}

// UpdateWebhookConfig updates or creates webhook configuration for a session
func (uc *WebhookUseCase) UpdateWebhookConfig(
	ctx context.Context,
	sessionID string,
	enabled bool,
	url string,
	events []string,
	ignoreGroups, ignoreBroadcasts, ignoreChannels bool,
) (*entity.WebhookConfig, error) {
	// Verify session exists
	if _, err := uc.sessionRepo.GetByID(ctx, sessionID); err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrNotFound.WithMessage("session not found")
		}
		return nil, errors.ErrDatabase.WithCause(err)
	}

	// Check if config exists
	config, err := uc.repo.GetBySessionID(ctx, sessionID)
	if err != nil && !errors.IsNotFound(err) {
		return nil, errors.ErrDatabase.WithCause(err)
	}

	// Create new config if it doesn't exist
	if config == nil {
		config = entity.NewWebhookConfig(uuid.New().String(), sessionID)

		// Generate secret if enabled
		if enabled {
			if err := config.GenerateSecret(); err != nil {
				return nil, errors.ErrInternal.WithCause(err)
			}
		}

		config.Update(enabled, url, events, ignoreGroups, ignoreBroadcasts, ignoreChannels)

		if err := uc.repo.Create(ctx, config); err != nil {
			return nil, errors.ErrDatabase.WithCause(err)
		}
	} else {
		// Update existing config
		config.Update(enabled, url, events, ignoreGroups, ignoreBroadcasts, ignoreChannels)

		if err := uc.repo.Update(ctx, config); err != nil {
			return nil, errors.ErrDatabase.WithCause(err)
		}
	}

	// Log webhook config update
	if uc.auditLogger != nil {
		uc.auditLogger.LogSessionAction(ctx, repository.SessionActionEvent{
			SessionID: sessionID,
			Action:    "webhook_config_updated",
			APIKeyID:  "", // API key ID would be extracted from context in production
			Timestamp: time.Now(),
		})
	}

	return config, nil
}

// RotateWebhookSecret generates a new secret for webhook configuration
func (uc *WebhookUseCase) RotateWebhookSecret(ctx context.Context, sessionID string) (*entity.WebhookConfig, error) {
	// Verify session exists
	if _, err := uc.sessionRepo.GetByID(ctx, sessionID); err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrNotFound.WithMessage("session not found")
		}
		return nil, errors.ErrDatabase.WithCause(err)
	}

	// Get existing config
	config, err := uc.repo.GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrNotFound.WithMessage("webhook configuration not found")
		}
		return nil, errors.ErrDatabase.WithCause(err)
	}

	// Generate new secret
	if err := config.GenerateSecret(); err != nil {
		return nil, errors.ErrInternal.WithCause(err)
	}

	// Update in repository
	if err := uc.repo.Update(ctx, config); err != nil {
		return nil, errors.ErrDatabase.WithCause(err)
	}

	// Log secret rotation
	if uc.auditLogger != nil {
		uc.auditLogger.LogSessionAction(ctx, repository.SessionActionEvent{
			SessionID: sessionID,
			Action:    "webhook_secret_rotated",
			APIKeyID:  "", // API key ID would be extracted from context in production
			Timestamp: time.Now(),
		})
	}

	return config, nil
}

// DeleteWebhookConfig removes webhook configuration for a session
func (uc *WebhookUseCase) DeleteWebhookConfig(ctx context.Context, sessionID string) error {
	// Verify session exists
	if _, err := uc.sessionRepo.GetByID(ctx, sessionID); err != nil {
		if errors.IsNotFound(err) {
			return errors.ErrNotFound.WithMessage("session not found")
		}
		return errors.ErrDatabase.WithCause(err)
	}

	// Delete config (ignore not found errors - idempotent)
	if err := uc.repo.Delete(ctx, sessionID); err != nil {
		if errors.IsNotFound(err) {
			return nil // Already deleted
		}
		return errors.ErrDatabase.WithCause(err)
	}

	// Log webhook config deletion
	if uc.auditLogger != nil {
		uc.auditLogger.LogSessionAction(ctx, repository.SessionActionEvent{
			SessionID: sessionID,
			Action:    "webhook_config_deleted",
			APIKeyID:  "", // API key ID would be extracted from context in production
			Timestamp: time.Now(),
		})
	}

	return nil
}
