package persistence

import (
	"context"
	"sync"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
)

// InMemoryAPIKeyRepository implements APIKeyRepository with in-memory storage
type InMemoryAPIKeyRepository struct {
	apiKeys      map[string]*entity.APIKey // ID -> APIKey
	keyHashIndex map[string]string         // KeyHash -> ID
	mu           sync.RWMutex
}

// NewInMemoryAPIKeyRepository creates a new in-memory API key repository
func NewInMemoryAPIKeyRepository() *InMemoryAPIKeyRepository {
	return &InMemoryAPIKeyRepository{
		apiKeys:      make(map[string]*entity.APIKey),
		keyHashIndex: make(map[string]string),
	}
}

// Save stores an API key in the repository
func (r *InMemoryAPIKeyRepository) Save(ctx context.Context, apiKey *entity.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if key hash already exists
	if existingID, exists := r.keyHashIndex[apiKey.KeyHash]; exists && existingID != apiKey.ID {
		return errors.ErrDuplicate.WithMessage("API key with this hash already exists")
	}

	// Store a copy to prevent external mutation
	apiKeyCopy := *apiKey
	r.apiKeys[apiKey.ID] = &apiKeyCopy

	// Index by key hash
	r.keyHashIndex[apiKey.KeyHash] = apiKey.ID

	return nil
}

// FindByKeyHash retrieves an API key by its key hash
func (r *InMemoryAPIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.keyHashIndex[keyHash]
	if !exists {
		return nil, errors.ErrNotFound.WithMessage("API key not found")
	}

	apiKey, exists := r.apiKeys[id]
	if !exists {
		return nil, errors.ErrNotFound.WithMessage("API key not found")
	}

	// Return a copy to prevent external mutation
	apiKeyCopy := *apiKey
	return &apiKeyCopy, nil
}

// FindByID retrieves an API key by its ID
func (r *InMemoryAPIKeyRepository) FindByID(ctx context.Context, id string) (*entity.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apiKey, exists := r.apiKeys[id]
	if !exists {
		return nil, errors.ErrNotFound.WithMessage("API key not found")
	}

	// Return a copy to prevent external mutation
	apiKeyCopy := *apiKey
	return &apiKeyCopy, nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *InMemoryAPIKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, exists := r.keyHashIndex[keyHash]
	if !exists {
		return errors.ErrNotFound.WithMessage("API key not found")
	}

	apiKey, exists := r.apiKeys[id]
	if !exists {
		return errors.ErrNotFound.WithMessage("API key not found")
	}

	now := time.Now()
	apiKey.LastUsedAt = &now
	return nil
}

// Delete removes an API key by its ID
func (r *InMemoryAPIKeyRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	apiKey, exists := r.apiKeys[id]
	if !exists {
		return errors.ErrNotFound.WithMessage("API key not found")
	}

	// Remove from key hash index
	delete(r.keyHashIndex, apiKey.KeyHash)

	// Remove from main storage
	delete(r.apiKeys, id)
	return nil
}

// List retrieves all API keys with pagination
func (r *InMemoryAPIKeyRepository) List(ctx context.Context, limit, offset int) ([]*entity.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect all API keys
	allKeys := make([]*entity.APIKey, 0, len(r.apiKeys))
	for _, apiKey := range r.apiKeys {
		apiKeyCopy := *apiKey
		allKeys = append(allKeys, &apiKeyCopy)
	}

	// Apply pagination
	start := offset
	if start >= len(allKeys) {
		return []*entity.APIKey{}, nil
	}

	end := start + limit
	if end > len(allKeys) {
		end = len(allKeys)
	}

	return allKeys[start:end], nil
}

// Clear removes all API keys (for testing)
func (r *InMemoryAPIKeyRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.apiKeys = make(map[string]*entity.APIKey)
	r.keyHashIndex = make(map[string]string)
}
