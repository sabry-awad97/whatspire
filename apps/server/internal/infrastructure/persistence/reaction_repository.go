package persistence

import (
	"context"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// ReactionRepository implements ReactionRepository with GORM
type ReactionRepository struct {
	db *gorm.DB
}

// NewReactionRepository creates a new GORM reaction repository
func NewReactionRepository(db *gorm.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

// Save stores a reaction in the repository
func (r *ReactionRepository) Save(ctx context.Context, reaction *entity.Reaction) error {
	model := &models.Reaction{
		ID:        reaction.ID,
		MessageID: reaction.MessageID,
		SessionID: reaction.SessionID,
		FromJID:   reaction.From,
		ToJID:     reaction.To,
		Emoji:     reaction.Emoji,
		CreatedAt: reaction.Timestamp,
	}

	// Use Create with OnConflict to handle upsert
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return nil
}

// FindByMessageID retrieves all reactions for a specific message
func (r *ReactionRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.Reaction, error) {
	var modelReactions []models.Reaction

	result := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("created_at DESC").
		Find(&modelReactions)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	reactions := make([]*entity.Reaction, 0, len(modelReactions))
	for _, model := range modelReactions {
		reaction := &entity.Reaction{
			ID:        model.ID,
			MessageID: model.MessageID,
			SessionID: model.SessionID,
			From:      model.FromJID,
			To:        model.ToJID,
			Emoji:     model.Emoji,
			Timestamp: model.CreatedAt,
		}
		reactions = append(reactions, reaction)
	}

	return reactions, nil
}

// FindBySessionID retrieves all reactions for a specific session with pagination
func (r *ReactionRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Reaction, error) {
	var modelReactions []models.Reaction

	result := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelReactions)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	reactions := make([]*entity.Reaction, 0, len(modelReactions))
	for _, model := range modelReactions {
		reaction := &entity.Reaction{
			ID:        model.ID,
			MessageID: model.MessageID,
			SessionID: model.SessionID,
			From:      model.FromJID,
			To:        model.ToJID,
			Emoji:     model.Emoji,
			Timestamp: model.CreatedAt,
		}
		reactions = append(reactions, reaction)
	}

	return reactions, nil
}

// Delete removes a reaction by its ID
func (r *ReactionRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Reaction{}, "id = ?", id)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// DeleteByMessageIDAndFrom removes a reaction by message ID and sender
func (r *ReactionRepository) DeleteByMessageIDAndFrom(ctx context.Context, messageID, from string) error {
	result := r.db.WithContext(ctx).
		Where("message_id = ? AND from_jid = ?", messageID, from).
		Delete(&models.Reaction{})

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return nil
}
