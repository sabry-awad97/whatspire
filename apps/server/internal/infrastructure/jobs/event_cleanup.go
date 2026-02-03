package jobs

import (
	"context"
	"log"
	"time"

	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/config"
)

// EventCleanupJob handles periodic cleanup of old events based on retention policy
type EventCleanupJob struct {
	eventRepo     repository.EventRepository
	cfg           *config.EventsConfig
	ticker        *time.Ticker
	stopCh        chan struct{}
	running       bool
	lastRunTime   time.Time
	lastRunResult *CleanupResult
}

// CleanupResult holds the result of a cleanup operation
type CleanupResult struct {
	DeletedCount int64
	Error        error
	Timestamp    time.Time
	Duration     time.Duration
}

// NewEventCleanupJob creates a new event cleanup job
func NewEventCleanupJob(eventRepo repository.EventRepository, cfg *config.EventsConfig) *EventCleanupJob {
	return &EventCleanupJob{
		eventRepo: eventRepo,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
	}
}

// Start starts the cleanup job
func (j *EventCleanupJob) Start(ctx context.Context) error {
	if j.running {
		return nil
	}

	j.running = true
	j.ticker = time.NewTicker(j.cfg.CleanupInterval)

	log.Printf("✅ Event cleanup job started (retention: %d days, interval: %v, cleanup time: %s)",
		j.cfg.RetentionDays, j.cfg.CleanupInterval, j.cfg.CleanupTime)

	// Run initial cleanup if we're past the scheduled time today
	go j.runIfScheduled(ctx)

	// Start the ticker loop
	go j.run(ctx)

	return nil
}

// Stop stops the cleanup job
func (j *EventCleanupJob) Stop() error {
	if !j.running {
		return nil
	}

	j.running = false
	close(j.stopCh)

	if j.ticker != nil {
		j.ticker.Stop()
	}

	log.Println("✅ Event cleanup job stopped")
	return nil
}

// run is the main loop that checks if cleanup should run
func (j *EventCleanupJob) run(ctx context.Context) {
	for {
		select {
		case <-j.ticker.C:
			j.runIfScheduled(ctx)
		case <-j.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runIfScheduled runs cleanup if we're at or past the scheduled time
func (j *EventCleanupJob) runIfScheduled(ctx context.Context) {
	// Skip if retention is 0 (keep forever)
	if j.cfg.RetentionDays == 0 {
		return
	}

	now := time.Now()

	// Check if we've already run today
	if j.lastRunTime.Year() == now.Year() &&
		j.lastRunTime.YearDay() == now.YearDay() {
		return
	}

	// Parse cleanup time (HH:MM format)
	scheduledTime, err := parseCleanupTime(j.cfg.CleanupTime)
	if err != nil {
		log.Printf("⚠️  Invalid cleanup time format: %v", err)
		return
	}

	// Check if current time is past scheduled time
	currentMinutes := now.Hour()*60 + now.Minute()
	scheduledMinutes := scheduledTime.Hour()*60 + scheduledTime.Minute()

	if currentMinutes >= scheduledMinutes {
		j.runCleanup(ctx)
	}
}

// runCleanup performs the actual cleanup operation
func (j *EventCleanupJob) runCleanup(ctx context.Context) {
	startTime := time.Now()

	log.Printf("🧹 Starting event cleanup (retention: %d days)...", j.cfg.RetentionDays)

	// Calculate cutoff timestamp
	cutoff := time.Now().AddDate(0, 0, -j.cfg.RetentionDays)
	timestamp := cutoff.Format(time.RFC3339)

	// Delete old events
	deleted, err := j.eventRepo.DeleteOlderThan(ctx, timestamp)

	duration := time.Since(startTime)

	// Store result
	j.lastRunTime = startTime
	j.lastRunResult = &CleanupResult{
		DeletedCount: deleted,
		Error:        err,
		Timestamp:    startTime,
		Duration:     duration,
	}

	if err != nil {
		log.Printf("❌ Event cleanup failed: %v (duration: %v)", err, duration)
		return
	}

	if deleted > 0 {
		log.Printf("✅ Event cleanup completed: deleted %d events (duration: %v)", deleted, duration)
	} else {
		log.Printf("✅ Event cleanup completed: no events to delete (duration: %v)", duration)
	}
}

// GetLastRunResult returns the result of the last cleanup run
func (j *EventCleanupJob) GetLastRunResult() *CleanupResult {
	return j.lastRunResult
}

// IsRunning returns whether the job is currently running
func (j *EventCleanupJob) IsRunning() bool {
	return j.running
}

// parseCleanupTime parses a time string in HH:MM format
func parseCleanupTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		timeStr = "02:00" // Default to 2 AM
	}

	// Parse as time in current timezone
	now := time.Now()
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}

	// Combine with today's date
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
}
