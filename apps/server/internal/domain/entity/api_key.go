package entity

import (
	"encoding/json"
	"time"
)

// APIKey represents an API key with role-based access control
type APIKey struct {
	ID               string     `json:"id"`
	KeyHash          string     `json:"key_hash"` // SHA-256 hash of the API key
	Role             string     `json:"role"`     // read, write, admin
	Description      *string    `json:"description,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	IsActive         bool       `json:"is_active"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevokedBy        *string    `json:"revoked_by,omitempty"`
	RevocationReason *string    `json:"revocation_reason,omitempty"`
}

// NewAPIKey creates a new APIKey with the given ID, key hash, role, and optional description
func NewAPIKey(id, keyHash, role string, description *string) *APIKey {
	now := time.Now()
	return &APIKey{
		ID:          id,
		KeyHash:     keyHash,
		Role:        role,
		Description: description,
		CreatedAt:   now,
		IsActive:    true,
	}
}

// UpdateLastUsed updates the last used timestamp
func (k *APIKey) UpdateLastUsed() {
	now := time.Now()
	k.LastUsedAt = &now
}

// Deactivate marks the API key as inactive
func (k *APIKey) Deactivate() {
	k.IsActive = false
}

// Activate marks the API key as active
func (k *APIKey) Activate() {
	k.IsActive = true
}

// Revoke marks the API key as revoked with optional reason and actor
func (k *APIKey) Revoke(revokedBy string, reason *string) {
	now := time.Now()
	k.IsActive = false
	k.RevokedAt = &now
	k.RevokedBy = &revokedBy
	k.RevocationReason = reason
}

// IsRevoked returns true if the API key has been revoked
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

// MarshalJSON implements json.Marshaler
func (k *APIKey) MarshalJSON() ([]byte, error) {
	type Alias APIKey
	aux := &struct {
		*Alias
		CreatedAt  string  `json:"created_at"`
		LastUsedAt *string `json:"last_used_at,omitempty"`
		RevokedAt  *string `json:"revoked_at,omitempty"`
	}{
		Alias:     (*Alias)(k),
		CreatedAt: k.CreatedAt.Format(time.RFC3339),
	}

	if k.LastUsedAt != nil {
		lastUsed := k.LastUsedAt.Format(time.RFC3339)
		aux.LastUsedAt = &lastUsed
	}

	if k.RevokedAt != nil {
		revokedAt := k.RevokedAt.Format(time.RFC3339)
		aux.RevokedAt = &revokedAt
	}

	return json.Marshal(aux)
}
