package http

import (
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

// QueryEvents handles GET /api/events
func (h *Handler) QueryEvents(c *gin.Context) {
	var req dto.QueryEventsRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_QUERY", "Invalid query parameters", nil)
		return
	}

	if h.eventUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Event use case not configured", nil)
		return
	}

	// Query events
	response, err := h.eventUC.QueryEvents(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetEventByID handles GET /api/events/:id
func (h *Handler) GetEventByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Event ID is required", nil)
		return
	}

	if h.eventUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Event use case not configured", nil)
		return
	}

	// Get event
	event, err := h.eventUC.GetEventByID(c.Request.Context(), id)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, event)
}

// ReplayEvents handles POST /api/events/replay
func (h *Handler) ReplayEvents(c *gin.Context) {
	var req dto.ReplayEventsRequest

	// Bind JSON body
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

	if h.eventUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Event use case not configured", nil)
		return
	}

	// Replay events
	response, err := h.eventUC.ReplayEvents(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	respondWithSuccess(c, statusCode, response)
}
