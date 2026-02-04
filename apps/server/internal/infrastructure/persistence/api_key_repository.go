package persistence

import (
	"context"
	"errors"
	"time"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// APIKeyRepository implements APIKeyRepository with GORM
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new GORM API key repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Save stores an API key in the repository
func (r *APIKeyRepository) Save(ctx context.Context, apiKey *entity.APIKey) error {
	model := &models.APIKey{
		ID:               apiKey.ID,
		KeyHash:          apiKey.KeyHash,
		Role:             apiKey.Role,
		Description:      apiKey.Description,
		CreatedAt:        apiKey.CreatedAt,
		LastUsedAt:       apiKey.LastUsedAt,
		IsActive:         apiKey.IsActive,
		RevokedAt:        apiKey.RevokedAt,
		RevokedBy:        apiKey.RevokedBy,
		RevocationReason: apiKey.RevocationReason,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return domainErrors.ErrDuplicate.WithMessage("API key with this hash already exists")
		}
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// FindByKeyHash retrieves an API key by its key hash
func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	var model models.APIKey

	result := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrNotFound.WithMessage("API key not found")
		}
		return nil, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	// Convert model to domain entity
	apiKey := &entity.APIKey{
		ID:               model.ID,
		KeyHash:          model.KeyHash,
		Role:             model.Role,
		Description:      model.Description,
		CreatedAt:        model.CreatedAt,
		LastUsedAt:       model.LastUsedAt,
		IsActive:         model.IsActive,
		RevokedAt:        model.RevokedAt,
		RevokedBy:        model.RevokedBy,
		RevocationReason: model.RevocationReason,
	}

	return apiKey, nil
}

// FindByID retrieves an API key by its ID
func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*entity.APIKey, error) {
	var model models.APIKey

	result := r.db.WithContext(ctx).First(&model, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrNotFound.WithMessage("API key not found")
		}
		return nil, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	// Convert model to domain entity
	apiKey := &entity.APIKey{
		ID:               model.ID,
		KeyHash:          model.KeyHash,
		Role:             model.Role,
		Description:      model.Description,
		CreatedAt:        model.CreatedAt,
		LastUsedAt:       model.LastUsedAt,
		IsActive:         model.IsActive,
		RevokedAt:        model.RevokedAt,
		RevokedBy:        model.RevokedBy,
		RevocationReason: model.RevocationReason,
	}

	return apiKey, nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.APIKey{}).
		Where("key_hash = ?", keyHash).
		Update("last_used_at", now)

	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound.WithMessage("API key not found")
	}

	return nil
}

// Delete removes an API key by its ID
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.APIKey{}, "id = ?", id)

	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound.WithMessage("API key not found")
	}

	return nil
}

// List retrieves all API keys with pagination and optional filters
func (r *APIKeyRepository) List(ctx context.Context, limit, offset int, role *string, isActive *bool) ([]*entity.APIKey, error) {
	var modelKeys []models.APIKey

	query := r.db.WithContext(ctx)

	// Apply role filter if provided
	if role != nil && *role != "" {
		query = query.Where("role = ?", *role)
	}

	// Apply status filter if provided
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	result := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelKeys)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	// Convert models to domain entities
	apiKeys := make([]*entity.APIKey, 0, len(modelKeys))
	for _, model := range modelKeys {
		apiKey := &entity.APIKey{
			ID:               model.ID,
			KeyHash:          model.KeyHash,
			Role:             model.Role,
			Description:      model.Description,
			CreatedAt:        model.CreatedAt,
			LastUsedAt:       model.LastUsedAt,
			IsActive:         model.IsActive,
			RevokedAt:        model.RevokedAt,
			RevokedBy:        model.RevokedBy,
			RevocationReason: model.RevocationReason,
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// Update updates an existing API key
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *entity.APIKey) error {
	model := &models.APIKey{
		ID:               apiKey.ID,
		KeyHash:          apiKey.KeyHash,
		Role:             apiKey.Role,
		Description:      apiKey.Description,
		CreatedAt:        apiKey.CreatedAt,
		LastUsedAt:       apiKey.LastUsedAt,
		IsActive:         apiKey.IsActive,
		RevokedAt:        apiKey.RevokedAt,
		RevokedBy:        apiKey.RevokedBy,
		RevocationReason: apiKey.RevocationReason,
	}

	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound.WithMessage("API key not found")
	}

	return nil
}

// Count returns the total number of API keys matching the filters
func (r *APIKeyRepository) Count(ctx context.Context, role *string, isActive *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.APIKey{})

	// Apply role filter if provided
	if role != nil && *role != "" {
		query = query.Where("role = ?", *role)
	}

	// Apply status filter if provided
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	result := query.Count(&count)

	if result.Error != nil {
		return 0, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return count, nil
}
