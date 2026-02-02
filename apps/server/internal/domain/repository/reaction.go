package repository

import (
	"context"

	"whatspire/internal/domain/entity"
)

// ReactionRepository defines reaction persistence operations
type ReactionRepository interface {
	// Save stores a reaction in the repository
	Save(ctx context.Context, reaction *entity.Reaction) error

	// FindByMessageID retrieves all reactions for a specific message
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.Reaction, error)

	// FindBySessionID retrieves all reactions for a specific session
	FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Reaction, error)

	// Delete removes a reaction by its ID
	Delete(ctx context.Context, id string) error

	// DeleteByMessageIDAndFrom removes a reaction by message ID and sender
	DeleteByMessageIDAndFrom(ctx context.Context, messageID, from string) error
}
