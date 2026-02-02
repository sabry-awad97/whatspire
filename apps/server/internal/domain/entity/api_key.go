package entity

import (
	"encoding/json"
	"time"
)

// APIKey represents an API key with role-based access control
type APIKey struct {
	ID         string     `json:"id"`
	KeyHash    string     `json:"key_hash"` // SHA-256 hash of the API key
	Role       string     `json:"role"`     // read, write, admin
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	IsActive   bool       `json:"is_active"`
}

// NewAPIKey creates a new APIKey with the given ID, key hash, and role
func NewAPIKey(id, keyHash, role string) *APIKey {
	now := time.Now()
	return &APIKey{
		ID:        id,
		KeyHash:   keyHash,
		Role:      role,
		CreatedAt: now,
		IsActive:  true,
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

// MarshalJSON implements json.Marshaler
func (k *APIKey) MarshalJSON() ([]byte, error) {
	type Alias APIKey
	aux := &struct {
		*Alias
		CreatedAt  string  `json:"created_at"`
		LastUsedAt *string `json:"last_used_at,omitempty"`
	}{
		Alias:     (*Alias)(k),
		CreatedAt: k.CreatedAt.Format(time.RFC3339),
	}

	if k.LastUsedAt != nil {
		lastUsed := k.LastUsedAt.Format(time.RFC3339)
		aux.LastUsedAt = &lastUsed
	}

	return json.Marshal(aux)
}
