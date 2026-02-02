package infrastructure

import (
	"context"
	"log"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/domain/valueobject"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/health"
	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/internal/infrastructure/storage"
	"whatspire/internal/infrastructure/webhook"
	"whatspire/internal/infrastructure/websocket"
	"whatspire/internal/infrastructure/whatsapp"

	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/fx"
)

// Module provides all infrastructure layer dependencies
var Module = fx.Module("infrastructure",
	fx.Provide(
		NewInMemorySessionRepository,
		NewWhatsmeowClient,
		fx.Annotate(
			func(c *whatsapp.WhatsmeowClient) *whatsapp.WhatsmeowClient { return c },
			fx.As(new(repository.WhatsAppClient)),
		),
		fx.Annotate(
			func(c *whatsapp.WhatsmeowClient) *whatsapp.WhatsmeowClient { return c },
			fx.As(new(repository.GroupFetcher)),
		),
		NewGorillaEventPublisher,
		NewAuditLogger,
		NewAuditLogRepository,
		NewWebhookPublisher,
		NewCompositeEventPublisher,
		NewEventHub,
		NewHealthCheckers,
		NewMediaUploader,
		NewReactionRepository,
		NewReceiptRepository,
		NewPresenceRepository,
		NewLocalMediaStorage,
	),
	// Wire EventHub to WhatsApp client events
	fx.Invoke(WireEventHubToWhatsAppClient),
	fx.Invoke(WireMessageHandler),
	fx.Invoke(WireReactionHandler),
)

// NewInMemorySessionRepository creates a new in-memory session repository
// Session state is stored in memory; whatsmeow's SQLite preserves auth for auto-reconnect
func NewInMemorySessionRepository() repository.SessionRepository {
	return persistence.NewInMemorySessionRepository()
}

// NewWhatsmeowClient creates a new WhatsApp client
func NewWhatsmeowClient(lc fx.Lifecycle, cfg *config.Config) (*whatsapp.WhatsmeowClient, error) {
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

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository() *persistence.InMemoryAuditLogRepository {
	return persistence.NewInMemoryAuditLogRepository()
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
			if err := publisher.Publish(ctx, event); err != nil {
				// Log error but don't block (events are queued internally)
			}
		}()
	})
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

// NewReactionRepository creates a new reaction repository
func NewReactionRepository() repository.ReactionRepository {
	return persistence.NewInMemoryReactionRepository()
}

// NewReceiptRepository creates a new receipt repository
func NewReceiptRepository() repository.ReceiptRepository {
	return persistence.NewInMemoryReceiptRepository()
}

// NewPresenceRepository creates a new presence repository
func NewPresenceRepository() repository.PresenceRepository {
	return persistence.NewInMemoryPresenceRepository()
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
