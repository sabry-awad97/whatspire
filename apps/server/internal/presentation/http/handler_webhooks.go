package http

import (
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

// GetWebhookConfig handles GET /api/sessions/:id/webhook
// Retrieves webhook configuration for a session
func (h *Handler) GetWebhookConfig(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	config, err := h.webhookUC.GetWebhookConfig(c.Request.Context(), sessionID)
	if err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewWebhookConfigResponse(config))
}

// UpdateWebhookConfig handles PUT /api/sessions/:id/webhook
// Updates or creates webhook configuration for a session
func (h *Handler) UpdateWebhookConfig(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	var req dto.UpdateWebhookConfigRequest
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

	// Additional validation: URL required if enabled
	if req.Enabled && req.URL == "" {
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "URL is required when webhook is enabled", nil)
		return
	}

	config, err := h.webhookUC.UpdateWebhookConfig(
		c.Request.Context(),
		sessionID,
		req.Enabled,
		req.URL,
		req.Events,
		req.IgnoreGroups,
		req.IgnoreBroadcasts,
		req.IgnoreChannels,
	)
	if err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewWebhookConfigResponse(config))
}

// RotateWebhookSecret handles POST /api/sessions/:id/webhook/rotate-secret
// Generates a new secret for webhook configuration
func (h *Handler) RotateWebhookSecret(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	config, err := h.webhookUC.RotateWebhookSecret(c.Request.Context(), sessionID)
	if err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewWebhookConfigResponse(config))
}

// DeleteWebhookConfig handles DELETE /api/sessions/:id/webhook
// Removes webhook configuration for a session
func (h *Handler) DeleteWebhookConfig(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	if err := h.webhookUC.DeleteWebhookConfig(c.Request.Context(), sessionID); err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]string{"message": "Webhook configuration deleted successfully"})
}
