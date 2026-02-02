package repository

import (
	"context"

	"whatspire/internal/domain/entity"
)

// PresenceRepository defines presence persistence operations
type PresenceRepository interface {
	// Save stores a presence update in the repository
	Save(ctx context.Context, presence *entity.Presence) error

	// FindBySessionID retrieves all presence updates for a specific session
	FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Presence, error)

	// FindByUserJID retrieves all presence updates for a specific user
	FindByUserJID(ctx context.Context, userJID string, limit, offset int) ([]*entity.Presence, error)

	// GetLatestByUserJID retrieves the most recent presence update for a user
	GetLatestByUserJID(ctx context.Context, userJID string) (*entity.Presence, error)

	// Delete removes a presence update by its ID
	Delete(ctx context.Context, id string) error
}
