package application

import (
	"context"

	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure"
	"whatspire/internal/infrastructure/config"

	"go.uber.org/fx"
)

// Module provides all application layer dependencies (use cases)
var Module = fx.Module("application",
	fx.Provide(
		NewSessionUseCase,
		NewMessageUseCase,
		NewHealthUseCase,
		NewGroupsUseCase,
		NewReactionUseCase,
		NewReceiptUseCase,
		NewPresenceUseCase,
		NewContactUseCase,
		NewEventUseCase,
		NewAPIKeyUseCase,
	),
)

// NewSessionUseCase creates a new session use case
func NewSessionUseCase(
	repo repository.SessionRepository,
	waClient repository.WhatsAppClient,
	publisher repository.EventPublisher,
	auditLogger repository.AuditLogger,
) *usecase.SessionUseCase {
	return usecase.NewSessionUseCase(repo, waClient, publisher, auditLogger)
}

// NewMessageUseCase creates a new message use case with lifecycle management
func NewMessageUseCase(
	lc fx.Lifecycle,
	waClient repository.WhatsAppClient,
	publisher repository.EventPublisher,
	mediaUploader repository.MediaUploader,
	auditLogger repository.AuditLogger,
	cfg *config.Config,
) *usecase.MessageUseCase {
	// Convert rate limit from per minute to per second
	rateLimitPerSecond := max(cfg.WhatsApp.MessageRateLimit/60, 1)

	msgConfig := usecase.MessageUseCaseConfig{
		MaxRetries:         3,
		RateLimitPerSecond: rateLimitPerSecond,
		QueueSize:          1000,
	}

	uc := usecase.NewMessageUseCase(waClient, publisher, mediaUploader, auditLogger, msgConfig)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			uc.Close()
			return nil
		},
	})

	return uc
}

// NewHealthUseCase creates a new health use case with all health checkers
func NewHealthUseCase(checkers *infrastructure.HealthCheckers) *usecase.HealthUseCase {
	return usecase.NewHealthUseCase(
		checkers.WhatsAppClient,
		checkers.EventPublisher,
	)
}

// NewGroupsUseCase creates a new groups use case
func NewGroupsUseCase(groupFetcher repository.GroupFetcher) *usecase.GroupsUseCase {
	return usecase.NewGroupsUseCase(groupFetcher)
}

// NewReactionUseCase creates a new reaction use case
func NewReactionUseCase(
	waClient repository.WhatsAppClient,
	reactionRepo repository.ReactionRepository,
	publisher repository.EventPublisher,
) *usecase.ReactionUseCase {
	return usecase.NewReactionUseCase(waClient, reactionRepo, publisher)
}

// NewReceiptUseCase creates a new receipt use case
func NewReceiptUseCase(
	waClient repository.WhatsAppClient,
	receiptRepo repository.ReceiptRepository,
	publisher repository.EventPublisher,
) *usecase.ReceiptUseCase {
	return usecase.NewReceiptUseCase(waClient, receiptRepo, publisher)
}

// NewPresenceUseCase creates a new presence use case
func NewPresenceUseCase(
	waClient repository.WhatsAppClient,
	presenceRepo repository.PresenceRepository,
	publisher repository.EventPublisher,
) *usecase.PresenceUseCase {
	return usecase.NewPresenceUseCase(waClient, presenceRepo, publisher)
}

// NewContactUseCase creates a new contact use case
func NewContactUseCase(
	waClient repository.WhatsAppClient,
) *usecase.ContactUseCase {
	return usecase.NewContactUseCase(waClient)
}

// NewEventUseCase creates a new event use case
func NewEventUseCase(
	eventRepo repository.EventRepository,
	publisher repository.EventPublisher,
) *usecase.EventUseCase {
	return usecase.NewEventUseCase(eventRepo, publisher)
}

// NewAPIKeyUseCase creates a new API key use case
func NewAPIKeyUseCase(
	repo repository.APIKeyRepository,
	auditLogger repository.AuditLogger,
) *usecase.APIKeyUseCase {
	return usecase.NewAPIKeyUseCase(repo, auditLogger)
}
