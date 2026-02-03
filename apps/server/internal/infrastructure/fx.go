package infrastructure

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/domain/valueobject"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/health"
	"whatspire/internal/infrastructure/jobs"
	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/internal/infrastructure/storage"
	"whatspire/internal/infrastructure/webhook"
	"whatspire/internal/infrastructure/websocket"
	"whatspire/internal/infrastructure/whatsapp"

	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// ensureDir creates a directory if it doesn't exist
// Returns the directory path and any error encountered
func ensureDir(path string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return dir, nil
}

// Module provides all infrastructure layer dependencies
var Module = fx.Module("infrastructure",
	fx.Provide(
		NewDB,
		fx.Annotate(
			NewSessionRepository,
			fx.As(new(repository.SessionRepository)),
		),
		NewWhatsmeowClient,
		fx.Annotate(
			func(c *whatsapp.WhatsmeowClient) *whatsapp.WhatsmeowClient { return c },
			fx.As(new(repository.WhatsAppClient)),
		),
		fx.Annotate(
			func(c *whatsapp.WhatsmeowClient) *whatsapp.WhatsmeowClient { return c },
			fx.As(new(repository.GroupFetcher)),
		),
		fx.Annotate(
			NewGorillaEventPublisher,
			fx.ResultTags(`name:"websocket"`),
		),
		NewAuditLogger,
		NewAuditLogRepository,
		NewWebhookPublisher,
		fx.Annotate(
			NewCompositeEventPublisher,
			fx.ParamTags(`name:"websocket"`),
			fx.As(new(repository.EventPublisher)),
		),
		NewEventHub,
		NewHealthCheckers,
		NewMediaUploader,
		fx.Annotate(
			NewReactionRepository,
			fx.As(new(repository.ReactionRepository)),
		),
		fx.Annotate(
			NewReceiptRepository,
			fx.As(new(repository.ReceiptRepository)),
		),
		fx.Annotate(
			NewPresenceRepository,
			fx.As(new(repository.PresenceRepository)),
		),
		fx.Annotate(
			NewAPIKeyRepository,
			fx.As(new(repository.APIKeyRepository)),
		),
		fx.Annotate(
			NewEventRepository,
			fx.As(new(repository.EventRepository)),
		),
		NewLocalMediaStorage,
		NewEventCleanupJob,
	),
	// Wire EventHub to WhatsApp client events
	fx.Invoke(WireEventHubToWhatsAppClient),
	fx.Invoke(WireMessageHandler),
	fx.Invoke(WireReactionHandler),
	fx.Invoke(RunMigrations),
	fx.Invoke(StartEventCleanupJob),
)

// NewDB creates a new GORM database connection using the configured driver
func NewDB(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	// Ensure the data directory exists for SQLite databases
	if cfg.Database.Driver == "sqlite" {
		dbDir, err := ensureDir(cfg.Database.DSN)
		if err != nil {
			return nil, err
		}
		log.Printf("✅ Database directory ensured: %s", dbDir)
	}

	// Create database factory
	factory := persistence.NewDatabaseFactory()

	// Create database connection using the factory
	db, err := factory.CreateDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Closing database connection...")
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return db, nil
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	return persistence.NewSessionRepository(db)
}

// NewReactionRepository creates a new reaction repository
func NewReactionRepository(db *gorm.DB) repository.ReactionRepository {
	return persistence.NewReactionRepository(db)
}

// NewReceiptRepository creates a new receipt repository
func NewReceiptRepository(db *gorm.DB) repository.ReceiptRepository {
	return persistence.NewReceiptRepository(db)
}

// NewPresenceRepository creates a new presence repository
func NewPresenceRepository(db *gorm.DB) repository.PresenceRepository {
	return persistence.NewPresenceRepository(db)
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) repository.APIKeyRepository {
	return persistence.NewAPIKeyRepository(db)
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *gorm.DB) repository.EventRepository {
	return persistence.NewEventRepository(db)
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *persistence.AuditLogRepository {
	return persistence.NewAuditLogRepository(db)
}

// NewWhatsmeowClient creates a new WhatsApp client
func NewWhatsmeowClient(lc fx.Lifecycle, cfg *config.Config) (*whatsapp.WhatsmeowClient, error) {
	// Ensure the data directory exists for WhatsApp database
	dbDir, err := ensureDir(cfg.WhatsApp.DBPath)
	if err != nil {
		return nil, err
	}
	log.Printf("✅ WhatsApp database directory ensured: %s", dbDir)

	clientConfig := whatsapp.ClientConfig{
		DBPath:           cfg.WhatsApp.DBPath,
		QRTimeout:        cfg.WhatsApp.QRTimeout,
		ReconnectDelay:   cfg.WhatsApp.ReconnectDelay,
		MaxReconnects:    cfg.WhatsApp.MaxReconnects,
		MessageRateLimit: cfg.WhatsApp.MessageRateLimit,
	}

	client, err := whatsapp.NewWhatsmeowClient(context.Background(), clientConfig)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Shutting down WhatsApp client...")

			// Close all WhatsApp connections gracefully
			if err := client.Close(); err != nil {
				log.Printf("⚠️  WhatsApp client shutdown error: %v", err)
				return err
			}

			log.Println("✅ WhatsApp client stopped gracefully")
			return nil
		},
	})

	return client, nil
}

// NewGorillaEventPublisher creates a new WebSocket event publisher
func NewGorillaEventPublisher(lc fx.Lifecycle, cfg *config.Config) repository.EventPublisher {
	publisherConfig := websocket.PublisherConfig{
		URL:            cfg.WebSocket.URL,
		APIKey:         cfg.WebSocket.APIKey,
		PingInterval:   cfg.WebSocket.PingInterval,
		PongTimeout:    cfg.WebSocket.PongTimeout,
		ReconnectDelay: cfg.WebSocket.ReconnectDelay,
		MaxReconnects:  cfg.WebSocket.MaxReconnects,
		QueueSize:      cfg.WebSocket.QueueSize,
	}

	publisher := websocket.NewGorillaEventPublisher(publisherConfig)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Connect to API server in background (non-blocking)
			go func() {
				_ = publisher.Connect(context.Background())
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Shutting down WebSocket publisher...")

			// Log queue status before shutdown
			queueSize := publisher.QueueSize()
			if queueSize > 0 {
				log.Printf("📤 Flushing %d queued events before shutdown...", queueSize)
			}

			// Disconnect will flush the queue automatically
			if err := publisher.Disconnect(ctx); err != nil {
				log.Printf("⚠️  WebSocket publisher shutdown error: %v", err)
				return err
			}

			log.Println("✅ WebSocket publisher stopped gracefully")
			return nil
		},
	})

	return publisher
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(cfg *config.Config) repository.AuditLogger {
	// Create structured logger
	structuredLogger := logger.NewStructuredLogger(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	// Create audit logger
	auditLogger := logger.NewAuditLogger(structuredLogger)

	return auditLogger
}

// NewWebhookPublisher creates a new webhook publisher (optional, based on config)
func NewWebhookPublisher(cfg *config.Config, auditLogger repository.AuditLogger) *webhook.WebhookPublisher {
	// Return nil if webhooks are not enabled
	if !cfg.Webhook.Enabled {
		return nil
	}

	webhookConfig := webhook.WebhookConfig{
		URL:    cfg.Webhook.URL,
		Secret: cfg.Webhook.Secret,
		Events: cfg.Webhook.Events,
	}

	// Create logger for webhook publisher
	structuredLogger := logger.NewStructuredLogger(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	publisher := webhook.NewWebhookPublisher(webhookConfig, structuredLogger, auditLogger)

	log.Printf("✅ Webhook publisher created (URL: %s, Events: %v)", cfg.Webhook.URL, cfg.Webhook.Events)

	return publisher
}

// NewCompositeEventPublisher creates a composite event publisher that publishes to both WebSocket and Webhook
func NewCompositeEventPublisher(
	websocketPublisher repository.EventPublisher,
	webhookPublisher *webhook.WebhookPublisher,
	cfg *config.Config,
) repository.EventPublisher {
	// Create logger for composite publisher
	logger := logger.NewStructuredLogger(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	return webhook.NewCompositeEventPublisher(websocketPublisher, webhookPublisher, logger)
}

// HealthCheckers holds all health checker instances
type HealthCheckers struct {
	WhatsAppClient *health.WhatsAppClientHealthChecker
	EventPublisher *health.EventPublisherHealthChecker
}

// NewHealthCheckers creates all health checkers
func NewHealthCheckers(
	waClient repository.WhatsAppClient,
	publisher repository.EventPublisher,
) *HealthCheckers {
	return &HealthCheckers{
		WhatsAppClient: health.NewWhatsAppClientHealthChecker(waClient),
		EventPublisher: health.NewEventPublisherHealthChecker(publisher),
	}
}

// NewMediaUploader creates a new media uploader
func NewMediaUploader(waClient *whatsapp.WhatsmeowClient, cfg *config.Config) repository.MediaUploader {
	// Create media constraints
	constraints := valueobject.DefaultMediaConstraints()

	// Create downloader config
	downloaderConfig := whatsapp.DefaultDownloaderConfig()

	// Create downloader
	downloader := whatsapp.NewHTTPMediaDownloader(downloaderConfig, constraints)

	// Create the media uploader
	mediaUploader := whatsapp.NewWhatsmeowMediaUploader(waClient, downloader, constraints)

	// Wire the media uploader to the client for sending media messages
	waClient.SetMediaUploader(mediaUploader)

	return mediaUploader
}

// WireEventHubToWhatsAppClient connects the EventHub and CompositeEventPublisher to receive events from the WhatsApp client
// This enables real-time event broadcasting to connected WebSocket clients AND external webhooks
func WireEventHubToWhatsAppClient(
	waClient *whatsapp.WhatsmeowClient,
	hub *websocket.EventHub,
	publisher repository.EventPublisher, // This is now the CompositeEventPublisher
	eventRepo repository.EventRepository,
	cfg *config.Config,
) {
	// Register an event handler that broadcasts events to the EventHub (for frontend WebSocket clients)
	waClient.RegisterEventHandler(func(event *entity.Event) {
		// Broadcast the event to all connected WebSocket clients
		hub.Broadcast(event)
	})

	// Register an event handler that publishes events via the CompositeEventPublisher
	// This will publish to both WebSocket (API server) and Webhook (if configured)
	waClient.RegisterEventHandler(func(event *entity.Event) {
		// Publish the event (non-blocking)
		go func() {
			ctx := context.Background()
			_ = publisher.Publish(ctx, event)
			// Ignore error - events are queued internally
		}()
	})

	// Register an event handler that persists events to the database (if enabled)
	if cfg.Events.Enabled {
		waClient.RegisterEventHandler(func(event *entity.Event) {
			// Persist the event (non-blocking)
			go func() {
				ctx := context.Background()
				if err := eventRepo.Create(ctx, event); err != nil {
					log.Printf("⚠️  Failed to persist event: %v", err)
				}
			}()
		})
		log.Println("✅ Event persistence enabled")
	}
}

// NewEventHub creates a new WebSocket event hub for broadcasting events to connected clients
func NewEventHub(lc fx.Lifecycle, cfg *config.Config) *websocket.EventHub {
	hubConfig := websocket.EventHubConfig{
		APIKey:       cfg.WebSocket.APIKey,
		PingInterval: cfg.WebSocket.PingInterval,
		WriteTimeout: 10 * time.Second,
		AuthTimeout:  10 * time.Second,
	}

	hub := websocket.NewEventHub(hubConfig)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start event hub in background
			go hub.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Shutting down WebSocket event hub...")

			// Log connected clients
			clientCount := hub.ClientCount()
			if clientCount > 0 {
				log.Printf("📡 Closing %d WebSocket client connections...", clientCount)
			}

			// Stop the hub (will close all client connections)
			hub.Stop()

			log.Println("✅ WebSocket event hub stopped gracefully")
			return nil
		},
	})

	return hub
}

// NewLocalMediaStorage creates a new local media storage
func NewLocalMediaStorage(cfg *config.Config) (repository.MediaStorage, error) {
	storageConfig := repository.MediaStorageConfig{
		BasePath:    cfg.Media.BasePath,
		BaseURL:     cfg.Media.BaseURL,
		MaxFileSize: cfg.Media.MaxFileSize,
	}

	storage, err := storage.NewLocalMediaStorage(storageConfig)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

// WireMessageHandler creates and wires the message handler to the WhatsApp client
func WireMessageHandler(
	waClient *whatsapp.WhatsmeowClient,
	cfg *config.Config,
	mediaStorage repository.MediaStorage,
	reactionRepo repository.ReactionRepository,
	presenceRepo repository.PresenceRepository,
	publisher repository.EventPublisher,
) {
	// Create media download helper
	mediaDownloadHelper := whatsapp.NewMediaDownloadHelper(mediaStorage)

	// Create message parser
	messageParser := whatsapp.NewMessageParser()

	// Create logger
	logger := waLog.Stdout("MessageHandler", "INFO", true)

	// Create message handler
	messageHandler := whatsapp.NewMessageHandler(
		messageParser,
		mediaDownloadHelper,
		mediaStorage,
		logger,
	)

	// Create reaction handler
	reactionLogger := waLog.Stdout("ReactionHandler", "INFO", true)
	reactionHandler := whatsapp.NewReactionHandler(
		reactionRepo,
		publisher,
		reactionLogger,
	)

	// Wire reaction handler to message handler
	messageHandler.SetReactionHandler(reactionHandler)

	// Wire message handler to the client
	waClient.SetMessageHandler(messageHandler)

	// Wire presence repository to the client
	waClient.SetPresenceRepository(presenceRepo)

	log.Println("✅ Message handler and reaction handler wired to WhatsApp client")
}

// WireReactionHandler is deprecated - reaction handler is now wired in WireMessageHandler
func WireReactionHandler(
	waClient *whatsapp.WhatsmeowClient,
	reactionRepo repository.ReactionRepository,
	publisher repository.EventPublisher,
) {
	// This function is kept for backward compatibility but does nothing
	// The reaction handler is now wired in WireMessageHandler
}

// RunMigrations runs GORM auto-migration on startup with version tracking
func RunMigrations(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("🔄 Running database migrations...")

			// Create migration runner
			runner := persistence.NewGORMMigrationRunner(db)

			// Get current version
			currentVersion, err := runner.Version(ctx)
			if err != nil {
				log.Printf("⚠️  Failed to get migration version: %v", err)
			} else {
				log.Printf("📊 Current migration version: %d", currentVersion)
			}

			// Run migrations
			if err := runner.Up(ctx); err != nil {
				log.Printf("⚠️  Migration error: %v", err)
				// Don't fail startup - continue with existing schema
				return nil
			}

			// Record migration if successful
			newVersion := int(time.Now().Unix())
			if err := runner.RecordMigration(ctx, newVersion, "auto_migration"); err != nil {
				log.Printf("⚠️  Failed to record migration: %v", err)
			} else {
				log.Printf("✅ Migration recorded: version %d", newVersion)
			}

			log.Println("✅ Database migrations completed")
			return nil
		},
	})
}

// NewEventCleanupJob creates a new event cleanup job
func NewEventCleanupJob(eventRepo repository.EventRepository, cfg *config.Config) *jobs.EventCleanupJob {
	return jobs.NewEventCleanupJob(eventRepo, &cfg.Events)
}

// StartEventCleanupJob starts the event cleanup job if event persistence is enabled
func StartEventCleanupJob(lc fx.Lifecycle, job *jobs.EventCleanupJob, cfg *config.Config) {
	// Only start if event persistence is enabled
	if !cfg.Events.Enabled {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return job.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Stopping event cleanup job...")
			return job.Stop()
		},
	})
}
