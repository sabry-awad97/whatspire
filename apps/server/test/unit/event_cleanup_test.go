package unit

import (
	"context"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/jobs"
	"whatspire/test/helpers"

	"github.com/stretchr/testify/assert"
)

// mockEventRepository is a simple mock implementation for testing
type mockEventRepository struct {
	deleteOlderThanCalled bool
	deleteOlderThanCount  int64
	deleteOlderThanError  error
}

func (m *mockEventRepository) Create(ctx context.Context, event *entity.Event) error {
	return nil
}

func (m *mockEventRepository) GetByID(ctx context.Context, id string) (*entity.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) List(ctx context.Context, filter repository.EventFilter) ([]*entity.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) Count(ctx context.Context, filter repository.EventFilter) (int64, error) {
	return 0, nil
}

func (m *mockEventRepository) DeleteOlderThan(ctx context.Context, timestamp string) (int64, error) {
	m.deleteOlderThanCalled = true
	return m.deleteOlderThanCount, m.deleteOlderThanError
}

func TestEventCleanupJob_Start(t *testing.T) {
	mockRepo := &mockEventRepository{}
	cfg := &config.EventsConfig{
		Enabled:         true,
		RetentionDays:   30,
		CleanupTime:     "02:00",
		CleanupInterval: 100 * time.Millisecond,
	}

	job := jobs.NewEventCleanupJob(mockRepo, cfg, helpers.CreateTestLogger())

	// Start the job
	err := job.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, job.IsRunning())

	// Stop the job
	err = job.Stop()
	assert.NoError(t, err)
	assert.False(t, job.IsRunning())
}

func TestEventCleanupJob_StartTwice(t *testing.T) {
	mockRepo := &mockEventRepository{}
	cfg := &config.EventsConfig{
		Enabled:         true,
		RetentionDays:   30,
		CleanupTime:     "02:00",
		CleanupInterval: 100 * time.Millisecond,
	}

	job := jobs.NewEventCleanupJob(mockRepo, cfg, helpers.CreateTestLogger())

	// Start the job
	err := job.Start(context.Background())
	assert.NoError(t, err)

	// Start again (should be no-op)
	err = job.Start(context.Background())
	assert.NoError(t, err)

	// Stop the job
	err = job.Stop()
	assert.NoError(t, err)
}

func TestEventCleanupJob_StopWithoutStart(t *testing.T) {
	mockRepo := &mockEventRepository{}
	cfg := &config.EventsConfig{
		Enabled:         true,
		RetentionDays:   30,
		CleanupTime:     "02:00",
		CleanupInterval: 100 * time.Millisecond,
	}

	job := jobs.NewEventCleanupJob(mockRepo, cfg, helpers.CreateTestLogger())

	// Stop without starting (should be no-op)
	err := job.Stop()
	assert.NoError(t, err)
}

func TestEventCleanupJob_ZeroRetention(t *testing.T) {
	mockRepo := &mockEventRepository{}
	cfg := &config.EventsConfig{
		Enabled:         true,
		RetentionDays:   0, // Keep forever
		CleanupTime:     "02:00",
		CleanupInterval: 50 * time.Millisecond,
	}

	job := jobs.NewEventCleanupJob(mockRepo, cfg, helpers.CreateTestLogger())

	// Start the job
	err := job.Start(context.Background())
	assert.NoError(t, err)

	// Wait a bit to ensure cleanup doesn't run
	time.Sleep(100 * time.Millisecond)

	// Stop the job
	err = job.Stop()
	assert.NoError(t, err)

	// Verify DeleteOlderThan was never called (retention is 0)
	assert.False(t, mockRepo.deleteOlderThanCalled)
}

func TestEventCleanupJob_GetLastRunResult(t *testing.T) {
	mockRepo := &mockEventRepository{}
	cfg := &config.EventsConfig{
		Enabled:         true,
		RetentionDays:   30,
		CleanupTime:     "02:00",
		CleanupInterval: 100 * time.Millisecond,
	}

	job := jobs.NewEventCleanupJob(mockRepo, cfg, helpers.CreateTestLogger())

	// Initially no result
	result := job.GetLastRunResult()
	assert.Nil(t, result)
}
