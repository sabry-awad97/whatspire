package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// EventRepository implements repository.EventRepository with GORM
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new GORM event repository
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create stores a new event in the repository
func (r *EventRepository) Create(ctx context.Context, event *entity.Event) error {
	model := &models.Event{
		ID:        event.ID,
		Type:      event.Type.String(),
		SessionID: event.SessionID,
		Data:      event.Data,
		Timestamp: event.Timestamp,
		CreatedAt: time.Now().UTC(),
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return nil
}

// GetByID retrieves an event by its ID
func (r *EventRepository) GetByID(ctx context.Context, id string) (*entity.Event, error) {
	var model models.Event

	result := r.db.WithContext(ctx).First(&model, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.ErrNotFound.WithMessage("event not found")
		}
		return nil, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	// Convert model to domain entity
	event := &entity.Event{
		ID:        model.ID,
		Type:      entity.EventType(model.Type),
		SessionID: model.SessionID,
		Data:      model.Data,
		Timestamp: model.Timestamp,
	}

	return event, nil
}

// List retrieves events with optional filters
func (r *EventRepository) List(ctx context.Context, filter repository.EventFilter) ([]*entity.Event, error) {
	query := r.db.WithContext(ctx).Model(&models.Event{})

	// Apply filters
	if filter.SessionID != nil {
		query = query.Where("session_id = ?", *filter.SessionID)
	}

	if filter.EventType != nil {
		query = query.Where("type = ?", filter.EventType.String())
	}

	if filter.Since != nil {
		sinceTime, err := time.Parse(time.RFC3339, *filter.Since)
		if err != nil {
			return nil, domainErrors.ErrInvalidInput.WithMessage("invalid since timestamp format").WithCause(err)
		}
		query = query.Where("timestamp >= ?", sinceTime)
	}

	if filter.Until != nil {
		untilTime, err := time.Parse(time.RFC3339, *filter.Until)
		if err != nil {
			return nil, domainErrors.ErrInvalidInput.WithMessage("invalid until timestamp format").WithCause(err)
		}
		query = query.Where("timestamp <= ?", untilTime)
	}

	// Apply ordering (newest first)
	query = query.Order("timestamp DESC")

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Execute query
	var modelEvents []models.Event
	result := query.Find(&modelEvents)
	if result.Error != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	// Convert models to domain entities
	events := make([]*entity.Event, 0, len(modelEvents))
	for _, model := range modelEvents {
		event := &entity.Event{
			ID:        model.ID,
			Type:      entity.EventType(model.Type),
			SessionID: model.SessionID,
			Data:      model.Data,
			Timestamp: model.Timestamp,
		}
		events = append(events, event)
	}

	return events, nil
}

// DeleteOlderThan removes events older than the specified timestamp
func (r *EventRepository) DeleteOlderThan(ctx context.Context, timestamp string) (int64, error) {
	deleteTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0, domainErrors.ErrInvalidInput.WithMessage("invalid timestamp format").WithCause(err)
	}

	result := r.db.WithContext(ctx).Where("timestamp < ?", deleteTime).Delete(&models.Event{})
	if result.Error != nil {
		return 0, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return result.RowsAffected, nil
}

// Count returns the total number of events matching the filter
func (r *EventRepository) Count(ctx context.Context, filter repository.EventFilter) (int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Event{})

	// Apply filters
	if filter.SessionID != nil {
		query = query.Where("session_id = ?", *filter.SessionID)
	}

	if filter.EventType != nil {
		query = query.Where("type = ?", filter.EventType.String())
	}

	if filter.Since != nil {
		sinceTime, err := time.Parse(time.RFC3339, *filter.Since)
		if err != nil {
			return 0, domainErrors.ErrInvalidInput.WithMessage("invalid since timestamp format").WithCause(err)
		}
		query = query.Where("timestamp >= ?", sinceTime)
	}

	if filter.Until != nil {
		untilTime, err := time.Parse(time.RFC3339, *filter.Until)
		if err != nil {
			return 0, domainErrors.ErrInvalidInput.WithMessage("invalid until timestamp format").WithCause(err)
		}
		query = query.Where("timestamp <= ?", untilTime)
	}

	// Execute count query
	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		return 0, domainErrors.ErrDatabaseError.WithCause(result.Error)
	}

	return count, nil
}

// EventStats represents statistics about stored events
type EventStats struct {
	TotalEvents      int64            `json:"total_events"`
	EventsByType     map[string]int64 `json:"events_by_type"`
	EventsBySession  map[string]int64 `json:"events_by_session"`
	OldestEvent      *time.Time       `json:"oldest_event,omitempty"`
	NewestEvent      *time.Time       `json:"newest_event,omitempty"`
	StorageSizeBytes int64            `json:"storage_size_bytes,omitempty"`
}

// GetStats returns statistics about stored events
func (r *EventRepository) GetStats(ctx context.Context) (*EventStats, error) {
	stats := &EventStats{
		EventsByType:    make(map[string]int64),
		EventsBySession: make(map[string]int64),
	}

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Event{}).Count(&stats.TotalEvents).Error; err != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(err)
	}

	// Get events by type
	var typeStats []struct {
		Type  string
		Count int64
	}
	if err := r.db.WithContext(ctx).Model(&models.Event{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(err)
	}
	for _, ts := range typeStats {
		stats.EventsByType[ts.Type] = ts.Count
	}

	// Get events by session
	var sessionStats []struct {
		SessionID string
		Count     int64
	}
	if err := r.db.WithContext(ctx).Model(&models.Event{}).
		Select("session_id, COUNT(*) as count").
		Group("session_id").
		Scan(&sessionStats).Error; err != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(err)
	}
	for _, ss := range sessionStats {
		stats.EventsBySession[ss.SessionID] = ss.Count
	}

	// Get oldest and newest event timestamps
	var oldest, newest time.Time
	if err := r.db.WithContext(ctx).Model(&models.Event{}).
		Select("MIN(timestamp) as oldest, MAX(timestamp) as newest").
		Row().Scan(&oldest, &newest); err != nil {
		return nil, domainErrors.ErrDatabaseError.WithCause(err)
	}
	if !oldest.IsZero() {
		stats.OldestEvent = &oldest
	}
	if !newest.IsZero() {
		stats.NewestEvent = &newest
	}

	return stats, nil
}

// ExportEvents exports events to JSON format for backup or analysis
func (r *EventRepository) ExportEvents(ctx context.Context, filter repository.EventFilter) ([]byte, error) {
	events, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	return data, nil
}
