package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/test/helpers"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for benchmarking
func setupTestDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	if err := persistence.RunAutoMigration(db, helpers.CreateTestLogger()); err != nil {
		b.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// seedEvents creates test events in the database
func seedEvents(b *testing.B, repo repository.EventRepository, count int) {
	ctx := context.Background()
	sessionIDs := []string{"session1", "session2", "session3"}
	eventTypes := []entity.EventType{
		entity.EventTypeMessageReceived,
		entity.EventTypeMessageSent,
		entity.EventTypeConnected,
		entity.EventTypeDisconnected,
	}

	for i := 0; i < count; i++ {
		event, err := entity.NewEventWithPayload(
			fmt.Sprintf("event-%d", i),
			eventTypes[i%len(eventTypes)],
			sessionIDs[i%len(sessionIDs)],
			map[string]string{"test": "data"},
		)
		if err != nil {
			b.Fatalf("Failed to create event: %v", err)
		}

		// Vary timestamps
		event.Timestamp = time.Now().Add(-time.Duration(i) * time.Minute)

		if err := repo.Create(ctx, event); err != nil {
			b.Fatalf("Failed to create event: %v", err)
		}
	}
}

// BenchmarkEventQuery_GetByID benchmarks retrieving a single event by ID
func BenchmarkEventQuery_GetByID(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 1000 events
	seedEvents(b, repo, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID(ctx, "event-500")
		if err != nil {
			b.Fatalf("Failed to get event: %v", err)
		}
	}
}

// BenchmarkEventQuery_ListBySession benchmarks listing events by session
func BenchmarkEventQuery_ListBySession(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 10000 events
	seedEvents(b, repo, 10000)

	sessionID := "session1"
	filter := repository.EventFilter{
		SessionID: &sessionID,
		Limit:     100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("Failed to list events: %v", err)
		}
	}
}

// BenchmarkEventQuery_ListByType benchmarks listing events by type
func BenchmarkEventQuery_ListByType(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 10000 events
	seedEvents(b, repo, 10000)

	eventType := entity.EventTypeMessageReceived
	filter := repository.EventFilter{
		EventType: &eventType,
		Limit:     100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("Failed to list events: %v", err)
		}
	}
}

// BenchmarkEventQuery_ListByTimeRange benchmarks listing events by time range
func BenchmarkEventQuery_ListByTimeRange(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 10000 events
	seedEvents(b, repo, 10000)

	since := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	until := time.Now().Format(time.RFC3339)
	filter := repository.EventFilter{
		Since: &since,
		Until: &until,
		Limit: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("Failed to list events: %v", err)
		}
	}
}

// BenchmarkEventQuery_ListWithPagination benchmarks paginated event listing
func BenchmarkEventQuery_ListWithPagination(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 10000 events
	seedEvents(b, repo, 10000)

	filter := repository.EventFilter{
		Limit:  50,
		Offset: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("Failed to list events: %v", err)
		}
	}
}

// BenchmarkEventQuery_Count benchmarks counting events
func BenchmarkEventQuery_Count(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 10000 events
	seedEvents(b, repo, 10000)

	sessionID := "session1"
	filter := repository.EventFilter{
		SessionID: &sessionID,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.Count(ctx, filter)
		if err != nil {
			b.Fatalf("Failed to count events: %v", err)
		}
	}
}

// BenchmarkEventQuery_ComplexFilter benchmarks complex filtered queries
func BenchmarkEventQuery_ComplexFilter(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed 10000 events
	seedEvents(b, repo, 10000)

	sessionID := "session1"
	eventType := entity.EventTypeMessageReceived
	since := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	filter := repository.EventFilter{
		SessionID: &sessionID,
		EventType: &eventType,
		Since:     &since,
		Limit:     50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, filter)
		if err != nil {
			b.Fatalf("Failed to list events: %v", err)
		}
	}
}

// BenchmarkEventCreate benchmarks event creation
func BenchmarkEventCreate(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event, err := entity.NewEventWithPayload(
			fmt.Sprintf("event-%d", i),
			entity.EventTypeMessageReceived,
			"session1",
			map[string]string{"test": "data"},
		)
		if err != nil {
			b.Fatalf("Failed to create event: %v", err)
		}

		if err := repo.Create(ctx, event); err != nil {
			b.Fatalf("Failed to persist event: %v", err)
		}
	}
}

// BenchmarkEventDelete benchmarks event deletion
func BenchmarkEventDelete(b *testing.B) {
	db := setupTestDB(b)
	repo := persistence.NewEventRepository(db)
	ctx := context.Background()

	// Seed events for each iteration
	for i := 0; i < b.N; i++ {
		seedEvents(b, repo, 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timestamp := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		_, err := repo.DeleteOlderThan(ctx, timestamp)
		if err != nil {
			b.Fatalf("Failed to delete events: %v", err)
		}
	}
}
