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
func (uc *APIKeyUseCase) maskAPIKey(key string) string {
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
