package http

import (
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

// QueryEvents handles GET /api/events
// @Summary Query events with filtering and pagination
// @Description Retrieve events from the database with optional filters for session, type, and time range
// @Tags events
// @Accept json
// @Produce json
// @Param session_id query string false "Filter by session ID"
// @Param event_type query string false "Filter by event type (e.g., message.received, message.sent)"
// @Param since query string false "Start timestamp in RFC3339 format"
// @Param until query string false "End timestamp in RFC3339 format"
// @Param limit query int false "Maximum number of results (1-1000, default: 100)"
// @Param offset query int false "Pagination offset (default: 0)"
// @Success 200 {object} dto.QueryEventsResponse "Successfully retrieved events"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /api/events [get]
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
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetEventByID handles GET /api/events/:id
// @Summary Get event by ID
// @Description Retrieve a single event by its unique identifier
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} dto.EventDTO "Successfully retrieved event"
// @Failure 400 {object} map[string]interface{} "Invalid event ID"
// @Failure 404 {object} map[string]interface{} "Event not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /api/events/{id} [get]
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
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, event)
}

// ReplayEvents handles POST /api/events/replay
// @Summary Replay events to event publisher
// @Description Re-publish events to webhook and WebSocket subscribers. Supports dry-run mode for testing.
// @Tags events
// @Accept json
// @Produce json
// @Param request body dto.ReplayEventsRequest true "Replay request with filters"
// @Success 200 {object} dto.ReplayEventsResponse "Successfully replayed events"
// @Failure 206 {object} dto.ReplayEventsResponse "Partial success - some events failed"
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation failed"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /api/events/replay [post]
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
		handleDomainError(c, err, h.logger)
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	respondWithSuccess(c, statusCode, response)
}
