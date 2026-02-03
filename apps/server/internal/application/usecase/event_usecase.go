package usecase

import (
	"context"
	"fmt"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
)

// EventUseCase handles event query and replay operations
type EventUseCase struct {
	eventRepo repository.EventRepository
	publisher repository.EventPublisher
}

// NewEventUseCase creates a new event use case
func NewEventUseCase(
	eventRepo repository.EventRepository,
	publisher repository.EventPublisher,
) *EventUseCase {
	return &EventUseCase{
		eventRepo: eventRepo,
		publisher: publisher,
	}
}

// QueryEvents retrieves events with filtering and pagination
func (uc *EventUseCase) QueryEvents(ctx context.Context, req dto.QueryEventsRequest) (*dto.QueryEventsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, errors.ErrInvalidInput.WithMessage(err.Error())
	}

	// Build filter
	filter := repository.EventFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	if req.SessionID != "" {
		filter.SessionID = &req.SessionID
	}

	if req.EventType != "" {
		eventType := entity.EventType(req.EventType)
		if !eventType.IsValid() {
			return nil, errors.ErrInvalidInput.WithMessage("invalid event type")
		}
		filter.EventType = &eventType
	}

	if req.Since != "" {
		filter.Since = &req.Since
	}

	if req.Until != "" {
		filter.Until = &req.Until
	}

	// Query events
	events, err := uc.eventRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get total count (without pagination)
	countFilter := filter
	countFilter.Limit = 0
	countFilter.Offset = 0
	total, err := uc.eventRepo.Count(ctx, countFilter)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	eventDTOs := make([]dto.EventDTO, len(events))
	for i, event := range events {
		eventDTOs[i] = dto.EventDTO{
			ID:        event.ID,
			Type:      event.Type.String(),
			SessionID: event.SessionID,
			Data:      event.Data,
			Timestamp: event.Timestamp.Format(time.RFC3339),
		}
	}

	return &dto.QueryEventsResponse{
		Events: eventDTOs,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// ReplayEvents re-publishes events to the event publisher
func (uc *EventUseCase) ReplayEvents(ctx context.Context, req dto.ReplayEventsRequest) (*dto.ReplayEventsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, errors.ErrInvalidInput.WithMessage(err.Error())
	}

	// Build filter
	filter := repository.EventFilter{
		Limit: 1000, // Safety limit to prevent replaying too many events at once
	}

	if req.SessionID != "" {
		filter.SessionID = &req.SessionID
	}

	if req.EventType != "" {
		eventType := entity.EventType(req.EventType)
		if !eventType.IsValid() {
			return nil, errors.ErrInvalidInput.WithMessage("invalid event type")
		}
		filter.EventType = &eventType
	}

	if req.Since != "" {
		filter.Since = &req.Since
	}

	if req.Until != "" {
		filter.Until = &req.Until
	}

	// Query events to replay
	events, err := uc.eventRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Dry run mode - just return what would be replayed
	if req.DryRun {
		return &dto.ReplayEventsResponse{
			Success:        true,
			EventsFound:    len(events),
			EventsReplayed: 0,
			DryRun:         true,
			Message:        fmt.Sprintf("Dry run: would replay %d events", len(events)),
		}, nil
	}

	// Replay events
	successCount := 0
	failedCount := 0
	var lastError error

	for _, event := range events {
		if err := uc.publisher.Publish(ctx, event); err != nil {
			failedCount++
			lastError = err
			// Continue replaying other events even if one fails
		} else {
			successCount++
		}
	}

	message := fmt.Sprintf("Replayed %d events successfully", successCount)
	if failedCount > 0 {
		message = fmt.Sprintf("Replayed %d events successfully, %d failed", successCount, failedCount)
	}

	return &dto.ReplayEventsResponse{
		Success:        failedCount == 0,
		EventsFound:    len(events),
		EventsReplayed: successCount,
		EventsFailed:   failedCount,
		DryRun:         false,
		Message:        message,
		LastError:      lastError,
	}, nil
}

// GetEventByID retrieves a single event by ID
func (uc *EventUseCase) GetEventByID(ctx context.Context, id string) (*dto.EventDTO, error) {
	if id == "" {
		return nil, errors.ErrInvalidInput.WithMessage("event ID is required")
	}

	event, err := uc.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.EventDTO{
		ID:        event.ID,
		Type:      event.Type.String(),
		SessionID: event.SessionID,
		Data:      event.Data,
		Timestamp: event.Timestamp.Format(time.RFC3339),
	}, nil
}

// DeleteOldEvents removes events older than the specified retention period
func (uc *EventUseCase) DeleteOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays < 0 {
		return 0, errors.ErrInvalidInput.WithMessage("retention days must be non-negative")
	}

	// Calculate cutoff timestamp
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	timestamp := cutoff.Format(time.RFC3339)

	// Delete old events
	deleted, err := uc.eventRepo.DeleteOlderThan(ctx, timestamp)
	if err != nil {
		return 0, err
	}

	return deleted, nil
}
