package persistence

import (
	"context"
	"errors"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// PresenceRepository implements PresenceRepository with GORM
type PresenceRepository struct {
	db *gorm.DB
}

// NewPresenceRepository creates a new GORM presence repository
func NewPresenceRepository(db *gorm.DB) *PresenceRepository {
	return &PresenceRepository{db: db}
}

// Save stores a presence update in the repository
func (r *PresenceRepository) Save(ctx context.Context, presence *entity.Presence) error {
	model := &models.Presence{
		ID:        presence.ID,
		SessionID: presence.SessionID,
		UserJID:   presence.UserJID,
		ChatJID:   presence.ChatJID,
		State:     presence.State.String(),
		CreatedAt: presence.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return nil
}

// FindBySessionID retrieves all presence updates for a specific session with pagination
func (r *PresenceRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Presence, error) {
	var modelPresences []models.Presence

	result := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelPresences)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	presences := make([]*entity.Presence, 0, len(modelPresences))
	for _, model := range modelPresences {
		presence := &entity.Presence{
			ID:        model.ID,
			SessionID: model.SessionID,
			UserJID:   model.UserJID,
			ChatJID:   model.ChatJID,
			State:     entity.PresenceState(model.State),
			Timestamp: model.CreatedAt,
		}
		presences = append(presences, presence)
	}

	return presences, nil
}

// FindByUserJID retrieves all presence updates for a specific user with pagination
func (r *PresenceRepository) FindByUserJID(ctx context.Context, userJID string, limit, offset int) ([]*entity.Presence, error) {
	var modelPresences []models.Presence

	result := r.db.WithContext(ctx).
		Where("user_jid = ?", userJID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelPresences)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	presences := make([]*entity.Presence, 0, len(modelPresences))
	for _, model := range modelPresences {
		presence := &entity.Presence{
			ID:        model.ID,
			SessionID: model.SessionID,
			UserJID:   model.UserJID,
			ChatJID:   model.ChatJID,
			State:     entity.PresenceState(model.State),
			Timestamp: model.CreatedAt,
		}
		presences = append(presences, presence)
	}

	return presences, nil
}

// GetLatestByUserJID retrieves the most recent presence update for a user
func (r *PresenceRepository) GetLatestByUserJID(ctx context.Context, userJID string) (*entity.Presence, error) {
	var model models.Presence

	result := r.db.WithContext(ctx).
		Where("user_jid = ?", userJID).
		Order("created_at DESC").
		First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert model to domain entity
	presence := &entity.Presence{
		ID:        model.ID,
		SessionID: model.SessionID,
		UserJID:   model.UserJID,
		ChatJID:   model.ChatJID,
		State:     entity.PresenceState(model.State),
		Timestamp: model.CreatedAt,
	}

	return presence, nil
}

// Delete removes a presence update by its ID
func (r *PresenceRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Presence{}, "id = ?", id)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}
