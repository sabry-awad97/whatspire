package dto

import "time"

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Role        string  `json:"role" binding:"required,oneof=read write admin"`
	Description *string `json:"description,omitempty"`
}

// CreateAPIKeyResponse represents the response after creating an API key
type CreateAPIKeyResponse struct {
	APIKey   APIKeyResponse `json:"api_key"`
	PlainKey string         `json:"plain_key"` // Only returned once during creation
}

// ListAPIKeysRequest represents the request to list API keys with filters
type ListAPIKeysRequest struct {
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Limit  int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Role   *string `form:"role" binding:"omitempty,oneof=read write admin"`
	Status *string `form:"status" binding:"omitempty,oneof=active revoked"`
}

// ListAPIKeysResponse represents the response for listing API keys
type ListAPIKeysResponse struct {
	APIKeys    []APIKeyResponse `json:"api_keys"`
	Pagination PaginationInfo   `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// RevokeAPIKeyRequest represents the request to revoke an API key
type RevokeAPIKeyRequest struct {
	Reason *string `json:"reason,omitempty"`
}

// RevokeAPIKeyResponse represents the response after revoking an API key
type RevokeAPIKeyResponse struct {
	ID        string    `json:"id"`
	RevokedAt time.Time `json:"revoked_at"`
	RevokedBy string    `json:"revoked_by"`
}

// APIKeyDetailsResponse represents detailed information about an API key
type APIKeyDetailsResponse struct {
	APIKey     APIKeyResponse `json:"api_key"`
	UsageStats UsageStats     `json:"usage_stats"`
}

// UsageStats contains usage statistics for an API key
type UsageStats struct {
	TotalRequests int `json:"total_requests"`
	Last7Days     int `json:"last_7_days"`
}

// APIKeyResponse represents an API key in responses
type APIKeyResponse struct {
	ID               string     `json:"id"`
	MaskedKey        string     `json:"masked_key"`
	Role             string     `json:"role"`
	Description      *string    `json:"description,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	IsActive         bool       `json:"is_active"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevokedBy        *string    `json:"revoked_by,omitempty"`
	RevocationReason *string    `json:"revocation_reason,omitempty"`
}
