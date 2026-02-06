package usecase

import (
	"context"
	"sync"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
	"whatspire/internal/domain/valueobject"
	"whatspire/internal/infrastructure/logger"

	"github.com/google/uuid"
)

// MessageUseCaseConfig holds configuration for the MessageUseCase
type MessageUseCaseConfig struct {
	// MaxRetries is the maximum number of retry attempts for failed messages
	MaxRetries int
	// RateLimitPerSecond is the maximum number of messages per second
	RateLimitPerSecond int
	// QueueSize is the maximum size of the message queue
	QueueSize int
}

// DefaultMessageUseCaseConfig returns the default configuration
func DefaultMessageUseCaseConfig() MessageUseCaseConfig {
	return MessageUseCaseConfig{
		MaxRetries:         3,
		RateLimitPerSecond: 10,
		QueueSize:          1000,
	}
}

// MessageUseCase handles message business logic
type MessageUseCase struct {
	waClient      repository.WhatsAppClient
	publisher     repository.EventPublisher
	mediaUploader repository.MediaUploader
	auditLogger   repository.AuditLogger
	config        MessageUseCaseConfig
	logger        *logger.Logger

	// Rate limiting
	mu            sync.Mutex
	lastSendTime  time.Time
	sendCount     int
	rateLimitChan chan struct{}

	// Message queue for rate limiting
	queue chan *entity.Message
	done  chan struct{}
}

// MessageUseCaseBuilder provides a builder pattern for creating MessageUseCase instances
type MessageUseCaseBuilder struct {
	usecase *MessageUseCase
}

// NewMessageUseCaseBuilder creates a new MessageUseCaseBuilder with required fields
func NewMessageUseCaseBuilder(config MessageUseCaseConfig, log *logger.Logger) *MessageUseCaseBuilder {
	return &MessageUseCaseBuilder{
		usecase: &MessageUseCase{
			config:        config,
			logger:        log.Sub("message_usecase"),
			rateLimitChan: make(chan struct{}, config.RateLimitPerSecond),
			queue:         make(chan *entity.Message, config.QueueSize),
			done:          make(chan struct{}),
		},
	}
}

// WithWhatsAppClient sets the WhatsApp client
func (b *MessageUseCaseBuilder) WithWhatsAppClient(client repository.WhatsAppClient) *MessageUseCaseBuilder {
	b.usecase.waClient = client
	return b
}

// WithEventPublisher sets the event publisher
func (b *MessageUseCaseBuilder) WithEventPublisher(publisher repository.EventPublisher) *MessageUseCaseBuilder {
	b.usecase.publisher = publisher
	return b
}

// WithMediaUploader sets the media uploader
func (b *MessageUseCaseBuilder) WithMediaUploader(uploader repository.MediaUploader) *MessageUseCaseBuilder {
	b.usecase.mediaUploader = uploader
	return b
}

// WithAuditLogger sets the audit logger
func (b *MessageUseCaseBuilder) WithAuditLogger(auditLogger repository.AuditLogger) *MessageUseCaseBuilder {
	b.usecase.auditLogger = auditLogger
	return b
}

// Build returns the constructed MessageUseCase and starts the message processor
func (b *MessageUseCaseBuilder) Build() *MessageUseCase {
	// Start the message processor
	go b.usecase.processQueue()

	return b.usecase
}

// SendMessage sends a WhatsApp message
func (uc *MessageUseCase) SendMessage(ctx context.Context, req dto.SendMessageRequest) (*entity.Message, error) {
	// Validate phone number
	_, err := valueobject.NewPhoneNumber(req.To)
	if err != nil {
		return nil, errors.ErrInvalidPhoneNumber
	}

	// Create message entity
	msgID := uuid.New().String()
	content := uc.buildMessageContent(req)
	msgType := uc.getMessageType(req.Type)

	msg := entity.NewMessageBuilder(msgID, req.SessionID).
		From(""). // From will be set by the WhatsApp client
		To(req.To).
		WithContent(content).
		WithType(msgType).
		Build()

	// Validate media if it's a media message
	if err := uc.validateMediaMessage(msg); err != nil {
		return nil, err
	}

	// Enqueue the message for rate-limited sending
	select {
	case uc.queue <- msg:
		// Message queued successfully
	default:
		return nil, errors.ErrMessageSendFailed.WithMessage("message queue is full")
	}

	// Emit pending status event
	uc.emitMessageStatusEvent(ctx, msg, entity.MessageStatusPending)

	return msg, nil
}

// SendMessageSync sends a message synchronously (bypassing the queue)
func (uc *MessageUseCase) SendMessageSync(ctx context.Context, req dto.SendMessageRequest) (*entity.Message, error) {
	// Validate phone number
	_, err := valueobject.NewPhoneNumber(req.To)
	if err != nil {
		return nil, errors.ErrInvalidPhoneNumber
	}

	// Create message entity
	msgID := uuid.New().String()
	content := uc.buildMessageContent(req)
	msgType := uc.getMessageType(req.Type)

	msg := entity.NewMessageBuilder(msgID, req.SessionID).
		From("").
		To(req.To).
		WithContent(content).
		WithType(msgType).
		Build()

	// Apply rate limiting
	if err := uc.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	// Send the message
	if err := uc.sendWithRetry(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// HandleIncomingMessage processes an incoming WhatsApp message
func (uc *MessageUseCase) HandleIncomingMessage(ctx context.Context, msg *entity.Message) error {
	if msg == nil {
		return errors.ErrInvalidInput.WithMessage("message cannot be nil")
	}

	// Emit message received event
	uc.emitMessageStatusEvent(ctx, msg, entity.MessageStatusDelivered)

	// Publish the incoming message event
	if uc.publisher != nil && uc.publisher.IsConnected() {
		event, err := entity.NewEventWithPayload(
			uuid.New().String(),
			entity.EventTypeMessageReceived,
			msg.SessionID,
			msg,
		)
		if err == nil {
			_ = uc.publisher.Publish(ctx, event)
		}
	}

	return nil
}

// HandleMessageStatusUpdate handles status updates for sent messages
func (uc *MessageUseCase) HandleMessageStatusUpdate(ctx context.Context, msgID, sessionID string, status entity.MessageStatus) error {
	// Determine the event type based on status
	var eventType entity.EventType
	switch status {
	case entity.MessageStatusSent:
		eventType = entity.EventTypeMessageSent
	case entity.MessageStatusDelivered:
		eventType = entity.EventTypeMessageDelivered
	case entity.MessageStatusRead:
		eventType = entity.EventTypeMessageRead
	case entity.MessageStatusFailed:
		eventType = entity.EventTypeMessageFailed
	default:
		return nil // Ignore unknown statuses
	}

	// Publish the status event
	if uc.publisher != nil && uc.publisher.IsConnected() {
		event, err := entity.NewEventWithPayload(
			uuid.New().String(),
			eventType,
			sessionID,
			map[string]interface{}{
				"message_id": msgID,
				"status":     status.String(),
			},
		)
		if err == nil {
			_ = uc.publisher.Publish(ctx, event)
		}
	}

	return nil
}

// Close stops the message processor
func (uc *MessageUseCase) Close() {
	close(uc.done)
}

// QueueSize returns the current number of messages in the queue
func (uc *MessageUseCase) QueueSize() int {
	return len(uc.queue)
}

// processQueue processes messages from the queue with rate limiting
func (uc *MessageUseCase) processQueue() {
	uc.logger.WithInt("queue_size", uc.config.QueueSize).
		Debug("Message queue processor started")

	for {
		select {
		case <-uc.done:
			uc.logger.Info("Message queue processor stopped gracefully")
			return
		case msg := <-uc.queue:
			uc.logger.WithFields(map[string]interface{}{
				"session_id":   msg.SessionID,
				"recipient":    msg.To,
				"message_id":   msg.ID,
				"message_type": msg.Type.String(),
			}).Debug("Processing message from queue")

			ctx := context.Background()

			// Apply rate limiting
			_ = uc.waitForRateLimit(ctx)

			// Send with retry
			if err := uc.sendWithRetry(ctx, msg); err != nil {
				uc.logger.WithError(err).
					WithFields(map[string]interface{}{
						"message_id":  msg.ID,
						"session_id":  msg.SessionID,
						"recipient":   msg.To,
						"retry_count": uc.config.MaxRetries,
					}).
					Error("Message send failed after all retry attempts")

				// Message failed after all retries
				msg.SetStatus(entity.MessageStatusFailed)
				uc.emitMessageStatusEvent(ctx, msg, entity.MessageStatusFailed)
			} else {
				uc.logger.WithFields(map[string]interface{}{
					"message_id": msg.ID,
					"session_id": msg.SessionID,
					"recipient":  msg.To,
				}).Info("Message sent successfully")
			}
		}
	}
}

// sendWithRetry sends a message with exponential backoff retry
func (uc *MessageUseCase) sendWithRetry(ctx context.Context, msg *entity.Message) error {
	var lastErr error

	for attempt := 0; attempt < uc.config.MaxRetries; attempt++ {
		uc.logger.WithFields(map[string]interface{}{
			"attempt":     attempt + 1,
			"max_retries": uc.config.MaxRetries,
			"session_id":  msg.SessionID,
			"message_id":  msg.ID,
		}).Debug("Attempting to send message")

		if uc.waClient == nil {
			lastErr = errors.ErrConnectionFailed.WithMessage("WhatsApp client not available")
			uc.logger.Warn("WhatsApp client is nil, cannot send message")
			continue
		}

		err := uc.waClient.SendMessage(ctx, msg)
		if err == nil {
			// Success
			uc.logger.WithFields(map[string]interface{}{
				"message_id": msg.ID,
				"session_id": msg.SessionID,
				"recipient":  msg.To,
			}).Debug("WhatsApp client successfully sent message")

			msg.SetStatus(entity.MessageStatusSent)
			uc.emitMessageStatusEvent(ctx, msg, entity.MessageStatusSent)

			// Log message sent
			if uc.auditLogger != nil {
				uc.auditLogger.LogMessageSent(ctx, repository.MessageSentEvent{
					SessionID:   msg.SessionID,
					Recipient:   msg.To,
					MessageType: msg.Type.String(),
					Timestamp:   time.Now(),
				})
			}

			return nil
		}

		lastErr = err
		uc.logger.WithError(err).
			WithFields(map[string]interface{}{
				"attempt":    attempt + 1,
				"message_id": msg.ID,
				"session_id": msg.SessionID,
			}).
			Warn("WhatsApp client failed to send message, will retry")

		// Exponential backoff: 1s, 2s, 4s
		backoff := time.Duration(1<<attempt) * time.Second
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return errors.ErrMessageSendFailed.WithCause(lastErr)
}

// waitForRateLimit waits until sending is allowed under rate limiting
func (uc *MessageUseCase) waitForRateLimit(ctx context.Context) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := time.Now()

	// Reset counter if a second has passed
	if now.Sub(uc.lastSendTime) >= time.Second {
		uc.sendCount = 0
		uc.lastSendTime = now
	}

	// Check if we've exceeded the rate limit
	if uc.sendCount >= uc.config.RateLimitPerSecond {
		// Wait until the next second
		waitTime := time.Second - now.Sub(uc.lastSendTime)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			uc.sendCount = 0
			uc.lastSendTime = time.Now()
		}
	}

	uc.sendCount++
	return nil
}

// emitMessageStatusEvent emits a message status event
func (uc *MessageUseCase) emitMessageStatusEvent(ctx context.Context, msg *entity.Message, status entity.MessageStatus) {
	if uc.publisher == nil || !uc.publisher.IsConnected() {
		return
	}

	var eventType entity.EventType
	switch status {
	case entity.MessageStatusPending:
		return // Don't emit pending status
	case entity.MessageStatusSent:
		eventType = entity.EventTypeMessageSent
	case entity.MessageStatusDelivered:
		eventType = entity.EventTypeMessageDelivered
	case entity.MessageStatusRead:
		eventType = entity.EventTypeMessageRead
	case entity.MessageStatusFailed:
		eventType = entity.EventTypeMessageFailed
	default:
		return
	}

	event, err := entity.NewEventWithPayload(
		uuid.New().String(),
		eventType,
		msg.SessionID,
		map[string]interface{}{
			"message_id": msg.ID,
			"to":         msg.To,
			"type":       msg.Type.String(),
			"status":     status.String(),
			"timestamp":  msg.Timestamp,
		},
	)
	if err == nil {
		_ = uc.publisher.Publish(ctx, event)
	}
}

// buildMessageContent builds MessageContent from the request
func (uc *MessageUseCase) buildMessageContent(req dto.SendMessageRequest) entity.MessageContent {
	content := entity.MessageContent{}

	if req.Content.Text != nil {
		content.Text = req.Content.Text
	}
	if req.Content.ImageURL != nil {
		content.ImageURL = req.Content.ImageURL
	}
	if req.Content.DocURL != nil {
		content.DocURL = req.Content.DocURL
	}
	if req.Content.AudioURL != nil {
		content.AudioURL = req.Content.AudioURL
	}
	if req.Content.VideoURL != nil {
		content.VideoURL = req.Content.VideoURL
	}
	if req.Content.Caption != nil {
		content.Caption = req.Content.Caption
	}
	if req.Content.Filename != nil {
		content.Filename = req.Content.Filename
	}

	return content
}

// getMessageType converts string type to MessageType
func (uc *MessageUseCase) getMessageType(typeStr string) entity.MessageType {
	switch typeStr {
	case "text":
		return entity.MessageTypeText
	case "image":
		return entity.MessageTypeImage
	case "document":
		return entity.MessageTypeDocument
	case "audio":
		return entity.MessageTypeAudio
	case "video":
		return entity.MessageTypeVideo
	default:
		return entity.MessageTypeText
	}
}

// validateMediaMessage validates media messages before sending
func (uc *MessageUseCase) validateMediaMessage(msg *entity.Message) error {
	switch msg.Type {
	case entity.MessageTypeImage:
		if msg.Content.ImageURL == nil || *msg.Content.ImageURL == "" {
			return errors.ErrEmptyContent.WithMessage("image URL is required for image messages")
		}
		if uc.mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
	case entity.MessageTypeDocument:
		if msg.Content.DocURL == nil || *msg.Content.DocURL == "" {
			return errors.ErrEmptyContent.WithMessage("document URL is required for document messages")
		}
		if uc.mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
	case entity.MessageTypeAudio:
		if msg.Content.AudioURL == nil || *msg.Content.AudioURL == "" {
			return errors.ErrEmptyContent.WithMessage("audio URL is required for audio messages")
		}
		if uc.mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
	case entity.MessageTypeVideo:
		if msg.Content.VideoURL == nil || *msg.Content.VideoURL == "" {
			return errors.ErrEmptyContent.WithMessage("video URL is required for video messages")
		}
		if uc.mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
	case entity.MessageTypeText:
		if msg.Content.Text == nil || *msg.Content.Text == "" {
			return errors.ErrEmptyContent.WithMessage("text content is required for text messages")
		}
	}
	return nil
}

// IsMediaUploadAvailable returns true if media upload is available
func (uc *MessageUseCase) IsMediaUploadAvailable() bool {
	return uc.mediaUploader != nil
}

// GetMediaConstraints returns the media constraints if media upload is available
func (uc *MessageUseCase) GetMediaConstraints() *valueobject.MediaConstraints {
	if uc.mediaUploader == nil {
		return nil
	}
	return uc.mediaUploader.GetConstraints()
}
