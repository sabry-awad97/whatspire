package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
)

// APIKeyUseCase handles API key operations (create, revoke, list, details)
type APIKeyUseCase struct {
	repo        repository.APIKeyRepository
	auditLogger repository.AuditLogger
}

// NewAPIKeyUseCase creates a new APIKeyUseCase
func NewAPIKeyUseCase(
	repo repository.APIKeyRepository,
	auditLogger repository.AuditLogger,
) *APIKeyUseCase {
	return &APIKeyUseCase{
		repo:        repo,
		auditLogger: auditLogger,
	}
}

// generateAPIKey generates a cryptographically secure random API key
// Returns a 32-byte base64-encoded string (43 characters)
func (uc *APIKeyUseCase) generateAPIKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.ErrInternal.WithMessage("failed to generate random key").WithCause(err)
	}

	// Encode to base64 (URL-safe, no padding)
	key := base64.RawURLEncoding.EncodeToString(bytes)
	return key, nil
}

// hashAPIKey hashes an API key using SHA-256
// Returns a hex-encoded hash string (64 characters)
func (uc *APIKeyUseCase) hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// maskAPIKey masks an API key for display purposes
// Shows first 8 and last 4 characters: "abcd1234...xyz9"
// This is exported so handlers can use it for masking keys in responses
func (uc *APIKeyUseCase) MaskAPIKey(key string) string {
	if len(key) <= 12 {
		// Key too short to mask meaningfully
		return "****"
	}

	prefix := key[:8]
	suffix := key[len(key)-4:]
	return fmt.Sprintf("%s...%s", prefix, suffix)
}

// CreateAPIKey generates a new API key with the specified role and optional description
// Returns the plain-text key (shown only once) and the created entity
func (uc *APIKeyUseCase) CreateAPIKey(ctx context.Context, role string, description *string, createdBy string) (plainKey string, apiKey *entity.APIKey, err error) {
	// Validate role
	if role != "read" && role != "write" && role != "admin" {
		return "", nil, errors.ErrValidationFailed.WithMessage("invalid role: must be read, write, or admin")
	}

	// Generate plain-text API key
	plainKey, err = uc.generateAPIKey()
	if err != nil {
		return "", nil, err
	}

	// Hash the key for storage
	keyHash := uc.hashAPIKey(plainKey)

	// Generate UUID for API key ID
	id := fmt.Sprintf("key_%d", time.Now().UnixNano())

	// Create entity
	apiKey = entity.NewAPIKey(id, keyHash, role, description)

	// Save to repository
	if err := uc.repo.Save(ctx, apiKey); err != nil {
		return "", nil, err
	}

	// Log API key creation
	if uc.auditLogger != nil {
		uc.auditLogger.LogAPIKeyCreated(ctx, repository.APIKeyCreatedEvent{
			APIKeyID:    apiKey.ID,
			Role:        apiKey.Role,
			Description: apiKey.Description,
			CreatedBy:   createdBy,
			Timestamp:   apiKey.CreatedAt,
		})
	}

	return plainKey, apiKey, nil
}

// RevokeAPIKey revokes an API key by its ID with optional reason
// The key will be immediately deactivated and cannot be used for authentication
func (uc *APIKeyUseCase) RevokeAPIKey(ctx context.Context, id string, revokedBy string, reason *string) (*entity.APIKey, error) {
	// Find the API key by ID
	apiKey, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if already revoked
	if apiKey.IsRevoked() {
		return nil, errors.ErrValidationFailed.WithMessage("API key is already revoked")
	}

	// Revoke the key
	apiKey.Revoke(revokedBy, reason)

	// Update in repository
	if err := uc.repo.Update(ctx, apiKey); err != nil {
		return nil, err
	}

	// Log API key revocation
	if uc.auditLogger != nil {
		uc.auditLogger.LogAPIKeyRevoked(ctx, repository.APIKeyRevokedEvent{
			APIKeyID:         apiKey.ID,
			RevokedBy:        revokedBy,
			RevocationReason: reason,
			Timestamp:        *apiKey.RevokedAt,
		})
	}

	return apiKey, nil
}

// ListAPIKeys retrieves a paginated list of API keys with optional filters
// Supports filtering by role and status, with pagination and sorting
func (uc *APIKeyUseCase) ListAPIKeys(ctx context.Context, page, limit int, role, status *string) ([]*entity.APIKey, int64, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50 // Default page size
	}

	// Validate role filter if provided
	if role != nil && *role != "" {
		if *role != "read" && *role != "write" && *role != "admin" {
			return nil, 0, errors.ErrValidationFailed.WithMessage("invalid role filter: must be read, write, or admin")
		}
	}

	// Validate status filter if provided
	var isActive *bool
	if status != nil && *status != "" {
		if *status == "active" {
			active := true
			isActive = &active
		} else if *status == "revoked" {
			inactive := false
			isActive = &inactive
		} else {
			return nil, 0, errors.ErrValidationFailed.WithMessage("invalid status filter: must be active or revoked")
		}
	}

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Retrieve API keys from repository
	apiKeys, err := uc.repo.List(ctx, limit, offset, role, isActive)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	total, err := uc.repo.Count(ctx, role, isActive)
	if err != nil {
		return nil, 0, err
	}

	return apiKeys, total, nil
}

// GetAPIKeyDetails retrieves detailed information about an API key including usage statistics
// Returns the API key entity and calculated usage stats
func (uc *APIKeyUseCase) GetAPIKeyDetails(ctx context.Context, id string) (*entity.APIKey, int, int, error) {
	// Find the API key by ID
	apiKey, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, 0, 0, err
	}

	// Calculate usage statistics
	// Note: This is a simplified implementation. In production, you would query
	// the audit_logs table to get actual usage counts.
	// For now, we'll return placeholder values that can be enhanced later.
	totalRequests := 0
	last7DaysRequests := 0

	// TODO: Query audit_logs table for actual usage statistics
	// Example query:
	// SELECT COUNT(*) FROM audit_logs
	// WHERE api_key_id = ? AND event_type = 'api_key_usage'
	// AND created_at >= NOW() - INTERVAL '7 days'

	return apiKey, totalRequests, last7DaysRequests, nil
}
