package http

import (
	"net/http"
	"strings"

	"whatspire/internal/application/dto"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

// CreateAPIKey handles POST /api/apikeys
// Creates a new API key with the specified role and optional description
//
// @Summary Create a new API key
// @Description Generates a new API key with the specified role. The plain-text key is returned only once and cannot be retrieved later.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param request body dto.CreateAPIKeyRequest true "API key creation request"
// @Success 201 {object} dto.CreateAPIKeyResponse "API key created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or validation failed"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing API key"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions (admin role required)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/apikeys [post]
func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Check if it's a validation error from Gin's binding tags
		// Validation errors contain "Field validation" in the error message
		if strings.Contains(err.Error(), "Field validation") {
			respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		// JSON parsing error or other binding error
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	// Additional custom validation if needed
	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	// Get the authenticated user/API key ID from context (set by auth middleware)
	// For now, we'll use a placeholder - this will be properly extracted from auth context
	createdBy := "system" // TODO: Extract from auth context

	// Create API key
	plainKey, apiKey, err := h.apikeyUC.CreateAPIKey(c.Request.Context(), req.Role, req.Description, createdBy)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	// Build response with plain-text key (shown only once)
	response := dto.CreateAPIKeyResponse{
		APIKey: dto.APIKeyResponse{
			ID:               apiKey.ID,
			MaskedKey:        "", // Will be masked on frontend
			Role:             apiKey.Role,
			Description:      apiKey.Description,
			CreatedAt:        apiKey.CreatedAt,
			LastUsedAt:       apiKey.LastUsedAt,
			IsActive:         apiKey.IsActive,
			RevokedAt:        apiKey.RevokedAt,
			RevokedBy:        apiKey.RevokedBy,
			RevocationReason: apiKey.RevocationReason,
		},
		PlainKey: plainKey, // Plain-text key - shown only once
	}

	respondWithSuccess(c, http.StatusCreated, response)
}

// RevokeAPIKey handles DELETE /api/apikeys/:id
// Revokes an API key immediately, preventing further authentication
//
// @Summary Revoke an API key
// @Description Revokes an API key by its ID. The key will be immediately deactivated and cannot be used for authentication. This action cannot be undone.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Param request body dto.RevokeAPIKeyRequest false "Revocation details (optional reason)"
// @Success 200 {object} dto.RevokeAPIKeyResponse "API key revoked successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or validation failed"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing API key"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions (admin role required)"
// @Failure 404 {object} ErrorResponse "API key not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/apikeys/{id} [delete]
func (h *Handler) RevokeAPIKey(c *gin.Context) {
	// Extract API key ID from URL parameter
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "MISSING_ID", "API key ID is required", nil)
		return
	}

	// Parse optional request body for revocation reason
	var req dto.RevokeAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Body is optional, so ignore binding errors
		req = dto.RevokeAPIKeyRequest{}
	}

	// Validate request if reason is provided
	if req.Reason != nil && *req.Reason != "" {
		if err := validator.Validate(req); err != nil {
			details := validator.ValidationErrors(err)
			respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
			return
		}
	}

	// Get the authenticated user/API key ID from context (set by auth middleware)
	// For now, we'll use a placeholder - this will be properly extracted from auth context
	revokedBy := "system" // TODO: Extract from auth context

	// Revoke API key
	apiKey, err := h.apikeyUC.RevokeAPIKey(c.Request.Context(), id, revokedBy, req.Reason)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	// Build response
	response := dto.RevokeAPIKeyResponse{
		ID:        apiKey.ID,
		RevokedAt: *apiKey.RevokedAt,
		RevokedBy: *apiKey.RevokedBy,
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// ListAPIKeys handles GET /api/apikeys
// Lists all API keys with optional filtering and pagination
//
// @Summary List API keys
// @Description Retrieves a paginated list of API keys with optional filters for role and status. Supports sorting by creation date.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param limit query int false "Items per page (default: 50, max: 100)" minimum(1) maximum(100)
// @Param role query string false "Filter by role" Enums(read, write, admin)
// @Param status query string false "Filter by status" Enums(active, revoked)
// @Success 200 {object} dto.ListAPIKeysResponse "API keys retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or validation failed"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing API key"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions (admin role required)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/apikeys [get]
func (h *Handler) ListAPIKeys(c *gin.Context) {
	// Parse query parameters
	var req dto.ListAPIKeysRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_QUERY", "Invalid query parameters", nil)
		return
	}

	// Validate request
	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	// Set default values
	page := req.Page
	if page == 0 {
		page = 1
	}
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}

	// List API keys
	apiKeys, total, err := h.apikeyUC.ListAPIKeys(c.Request.Context(), page, limit, req.Role, req.Status)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	// Build response
	apiKeyResponses := make([]dto.APIKeyResponse, len(apiKeys))
	for i, key := range apiKeys {
		// Mask the key for display
		maskedKey := h.apikeyUC.MaskAPIKey(key.KeyHash) // Note: This masks the hash, not the original key

		apiKeyResponses[i] = dto.APIKeyResponse{
			ID:               key.ID,
			MaskedKey:        maskedKey,
			Role:             key.Role,
			Description:      key.Description,
			CreatedAt:        key.CreatedAt,
			LastUsedAt:       key.LastUsedAt,
			IsActive:         key.IsActive,
			RevokedAt:        key.RevokedAt,
			RevokedBy:        key.RevokedBy,
			RevocationReason: key.RevocationReason,
		}
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := dto.ListAPIKeysResponse{
		APIKeys: apiKeyResponses,
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetAPIKeyDetails handles GET /api/apikeys/:id
// Retrieves detailed information about a specific API key including usage statistics
//
// @Summary Get API key details
// @Description Retrieves comprehensive information about an API key including metadata, usage statistics, and audit trail. Provides insights into key usage patterns and history.
// @Tags API Keys
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} dto.APIKeyDetailsResponse "API key details retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid request - missing or invalid ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing API key"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions (admin role required)"
// @Failure 404 {object} ErrorResponse "API key not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/apikeys/{id} [get]
func (h *Handler) GetAPIKeyDetails(c *gin.Context) {
	// Extract API key ID from URL parameter
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "MISSING_ID", "API key ID is required", nil)
		return
	}

	// Get API key details with usage statistics
	apiKey, totalRequests, last7Days, err := h.apikeyUC.GetAPIKeyDetails(c.Request.Context(), id)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	// Mask the key for display
	maskedKey := h.apikeyUC.MaskAPIKey(apiKey.KeyHash)

	// Build response
	response := dto.APIKeyDetailsResponse{
		APIKey: dto.APIKeyResponse{
			ID:               apiKey.ID,
			MaskedKey:        maskedKey,
			Role:             apiKey.Role,
			Description:      apiKey.Description,
			CreatedAt:        apiKey.CreatedAt,
			LastUsedAt:       apiKey.LastUsedAt,
			IsActive:         apiKey.IsActive,
			RevokedAt:        apiKey.RevokedAt,
			RevokedBy:        apiKey.RevokedBy,
			RevocationReason: apiKey.RevocationReason,
		},
		UsageStats: dto.UsageStats{
			TotalRequests: totalRequests,
			Last7Days:     last7Days,
		},
	}

	respondWithSuccess(c, http.StatusOK, response)
}
