package http

import (
	"net/http"

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
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	// Validate request
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
