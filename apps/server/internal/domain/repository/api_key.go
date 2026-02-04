package repository

import (
	"context"

	"whatspire/internal/domain/entity"
)

// APIKeyRepository defines API key persistence operations
type APIKeyRepository interface {
	// Save stores an API key in the repository
	Save(ctx context.Context, apiKey *entity.APIKey) error

	// FindByKeyHash retrieves an API key by its key hash
	FindByKeyHash(ctx context.Context, keyHash string) (*entity.APIKey, error)

	// FindByID retrieves an API key by its ID
	FindByID(ctx context.Context, id string) (*entity.APIKey, error)

	// UpdateLastUsed updates the last used timestamp for an API key
	UpdateLastUsed(ctx context.Context, keyHash string) error

	// Update updates an existing API key
	Update(ctx context.Context, apiKey *entity.APIKey) error

	// Delete removes an API key by its ID
	Delete(ctx context.Context, id string) error

	// List retrieves all API keys with pagination
	List(ctx context.Context, limit, offset int) ([]*entity.APIKey, error)

	// Count returns the total number of API keys
	Count(ctx context.Context) (int64, error)
}
