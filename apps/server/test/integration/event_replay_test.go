package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/internal/infrastructure/persistence/models"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockEventPublisher tracks published events for testing
type mockEventPublisher struct {
	publishedEvents []*entity.Event
	publishError    error
	publishFunc     func(ctx context.Context, event *entity.Event) error
}

func (m *mockEventPublisher) Publish(ctx context.Context, event *entity.Event) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, event)
	}
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventPublisher) Connect(ctx context.Context) error {
	return nil
}

func (m *mockEventPublisher) Disconnect(ctx context.Context) error {
	return nil
}

func (m *mockEventPublisher) IsConnected() bool {
	return true
}

func (m *mockEventPublisher) QueueSize() int {
	return 0
}

func setupEventReplayTest(t *testing.T) (*gorm.DB, repository.EventRepository, *mockEventPublisher, *gin.Engine) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&models.Event{})
	require.NoError(t, err)

	// Create repositories
	eventRepo := persistence.NewEventRepository(db)

	// Create mock publisher
	publisher := &mockEventPublisher{
		publishedEvents: make([]*entity.Event, 0),
	}

	// Create use case
	eventUC := usecase.NewEventUseCase(eventRepo, publisher)

	// Create handler
	handler := helpers.NewTestHandlerBuilder().
		WithEventUseCase(eventUC).
		Build()

	// Create router
	router := helpers.CreateTestRouterWithDefaults(handler)

	return db, eventRepo, publisher, router
}

func TestEventReplay_WithWebhookDelivery(t *testing.T) {
	db, eventRepo, publisher, router := setupEventReplayTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	ctx := context.Background()

	// Create test events
	sessionID := "test-session-123"
	now := time.Now()
	events := []*entity.Event{
		{
			ID:        "evt-1",
			Type:      entity.EventTypeMessageReceived,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Hello"}`),
			Timestamp: now.Add(-2 * time.Hour), // Older event
		},
		{
			ID:        "evt-2",
			Type:      entity.EventTypeMessageSent,
			SessionID: sessionID,
			Data:      []byte(`{"text":"World"}`),
			Timestamp: now.Add(-1 * time.Hour), // Newer event
		},
	}

	// Store events in database
	for _, event := range events {
		err := eventRepo.Create(ctx, event)
		require.NoError(t, err)
	}

	// Test dry run first
	t.Run("DryRun", func(t *testing.T) {
		reqBody := dto.ReplayEventsRequest{
			SessionID: sessionID,
			DryRun:    true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.True(t, data["dry_run"].(bool))
		assert.Equal(t, float64(2), data["events_found"].(float64))
		assert.Equal(t, float64(0), data["events_replayed"].(float64))

		// Verify no events were published
		assert.Len(t, publisher.publishedEvents, 0)
	})

	// Test actual replay
	t.Run("ActualReplay", func(t *testing.T) {
		// Reset publisher
		publisher.publishedEvents = make([]*entity.Event, 0)

		reqBody := dto.ReplayEventsRequest{
			SessionID: sessionID,
			DryRun:    false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.False(t, data["dry_run"].(bool))
		assert.Equal(t, float64(2), data["events_found"].(float64))
		assert.Equal(t, float64(2), data["events_replayed"].(float64))

		// Verify events were published in correct order (newest first due to DESC ordering)
		assert.Len(t, publisher.publishedEvents, 2)
		assert.Equal(t, "evt-2", publisher.publishedEvents[0].ID, "evt-2 should be first (newer)")
		assert.Equal(t, "evt-1", publisher.publishedEvents[1].ID, "evt-1 should be second (older)")
	})
}

func TestEventReplay_WithEventTypeFilter(t *testing.T) {
	db, eventRepo, publisher, router := setupEventReplayTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	ctx := context.Background()

	// Create test events with different types
	sessionID := "test-session-456"
	events := []*entity.Event{
		{
			ID:        "evt-3",
			Type:      entity.EventTypeMessageReceived,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Message 1"}`),
			Timestamp: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "evt-4",
			Type:      entity.EventTypeMessageSent,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Message 2"}`),
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "evt-5",
			Type:      entity.EventTypePresenceUpdate,
			SessionID: sessionID,
			Data:      []byte(`{"status":"available"}`),
			Timestamp: time.Now(),
		},
	}

	// Store events in database
	for _, event := range events {
		err := eventRepo.Create(ctx, event)
		require.NoError(t, err)
	}

	// Replay only message.received events
	reqBody := dto.ReplayEventsRequest{
		SessionID: sessionID,
		EventType: string(entity.EventTypeMessageReceived),
		DryRun:    false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["events_found"].(float64))
	assert.Equal(t, float64(1), data["events_replayed"].(float64))

	// Verify only message.received event was published
	assert.Len(t, publisher.publishedEvents, 1)
	assert.Equal(t, "evt-3", publisher.publishedEvents[0].ID)
	assert.Equal(t, entity.EventTypeMessageReceived, publisher.publishedEvents[0].Type)
}

func TestEventReplay_WithTimeRangeFilter(t *testing.T) {
	db, eventRepo, publisher, router := setupEventReplayTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	ctx := context.Background()

	// Create test events at different times
	sessionID := "test-session-789"
	now := time.Now()
	events := []*entity.Event{
		{
			ID:        "evt-6",
			Type:      entity.EventTypeMessageReceived,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Old message"}`),
			Timestamp: now.Add(-3 * time.Hour),
		},
		{
			ID:        "evt-7",
			Type:      entity.EventTypeMessageReceived,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Recent message"}`),
			Timestamp: now.Add(-30 * time.Minute),
		},
	}

	// Store events in database
	for _, event := range events {
		err := eventRepo.Create(ctx, event)
		require.NoError(t, err)
	}

	// Replay only events from the last hour
	since := now.Add(-1 * time.Hour).Format(time.RFC3339)
	reqBody := dto.ReplayEventsRequest{
		SessionID: sessionID,
		Since:     since,
		DryRun:    false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["events_found"].(float64))
	assert.Equal(t, float64(1), data["events_replayed"].(float64))

	// Verify only recent event was published
	assert.Len(t, publisher.publishedEvents, 1)
	assert.Equal(t, "evt-7", publisher.publishedEvents[0].ID)
}

func TestEventReplay_ValidationErrors(t *testing.T) {
	_, _, _, router := setupEventReplayTest(t)

	t.Run("NoFilters", func(t *testing.T) {
		reqBody := dto.ReplayEventsRequest{
			DryRun: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEventReplay_PartialFailure(t *testing.T) {
	db, eventRepo, _, _ := setupEventReplayTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	ctx := context.Background()

	// Create custom publisher that fails on second event
	callCount := 0
	failingPublisher := &mockEventPublisher{
		publishedEvents: make([]*entity.Event, 0),
	}

	failingPublisher.publishFunc = func(ctx context.Context, event *entity.Event) error {
		callCount++
		if callCount == 2 {
			return assert.AnError
		}
		failingPublisher.publishedEvents = append(failingPublisher.publishedEvents, event)
		return nil
	}

	// Create use case with failing publisher
	eventUC := usecase.NewEventUseCase(eventRepo, failingPublisher)
	handler := helpers.NewTestHandlerBuilder().
		WithEventUseCase(eventUC).
		Build()
	testRouter := helpers.CreateTestRouterWithDefaults(handler)

	// Create test events
	sessionID := "test-session-fail"
	events := []*entity.Event{
		{
			ID:        "evt-8",
			Type:      entity.EventTypeMessageReceived,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Event 1"}`),
			Timestamp: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "evt-9",
			Type:      entity.EventTypeMessageReceived,
			SessionID: sessionID,
			Data:      []byte(`{"text":"Event 2"}`),
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
	}

	// Store events in database
	for _, event := range events {
		err := eventRepo.Create(ctx, event)
		require.NoError(t, err)
	}

	// Replay events
	reqBody := dto.ReplayEventsRequest{
		SessionID: sessionID,
		DryRun:    false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/replay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testRouter.ServeHTTP(w, req)

	// Should return 206 Partial Content for partial success
	assert.Equal(t, http.StatusPartialContent, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["events_found"].(float64))
	assert.Equal(t, float64(1), data["events_replayed"].(float64))
	assert.Equal(t, float64(1), data["events_failed"].(float64))
}
