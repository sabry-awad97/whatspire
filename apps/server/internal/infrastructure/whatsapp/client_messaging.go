package whatsapp

import (
	"context"
	"strings"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// SendMessage sends a message through WhatsApp
func (c *WhatsmeowClient) SendMessage(ctx context.Context, msg *entity.Message) error {
	// Use circuit breaker if enabled
	if c.circuitBreaker != nil {
		_, err := c.circuitBreaker.Execute(ctx, func() (any, error) {
			return nil, c.sendMessageInternal(ctx, msg)
		})
		return err
	}
	return c.sendMessageInternal(ctx, msg)
}

// sendMessageInternal performs the actual message sending logic
func (c *WhatsmeowClient) sendMessageInternal(ctx context.Context, msg *entity.Message) error {
	c.mu.RLock()
	client, exists := c.clients[msg.SessionID]
	mediaUploader := c.mediaUploader
	c.mu.RUnlock()

	c.logger.Infof("sendMessageInternal: sessionID=%s, exists=%v, clientsCount=%d", msg.SessionID, exists, len(c.clients))

	// Log all client keys for debugging
	c.mu.RLock()
	for k := range c.clients {
		c.logger.Infof("sendMessageInternal: available client key=%s", k)
	}
	c.mu.RUnlock()

	if !exists {
		c.logger.Warnf("sendMessageInternal: session not found: %s", msg.SessionID)
		return errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		c.logger.Warnf("sendMessageInternal: session not connected: %s", msg.SessionID)
		return errors.ErrDisconnected
	}

	c.logger.Infof("sendMessageInternal: client is connected, sending to %s", msg.To)

	// Parse recipient JID - strip leading + from phone number
	phoneNumber := string(msg.To)
	phoneNumber = strings.TrimPrefix(phoneNumber, "+")
	recipientJID, err := types.ParseJID(phoneNumber + "@s.whatsapp.net")
	if err != nil {
		return errors.ErrInvalidPhoneNumber.WithCause(err)
	}

	// Build message based on type
	var waMsg *waE2E.Message

	switch msg.Type {
	case entity.MessageTypeImage:
		if mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
		if msg.Content.ImageURL == nil || *msg.Content.ImageURL == "" {
			return errors.ErrEmptyContent.WithMessage("image URL is required")
		}
		uploadResult, err := mediaUploader.UploadImage(ctx, msg.SessionID, *msg.Content.ImageURL)
		if err != nil {
			return errors.ErrMediaUploadFailed.WithCause(err)
		}
		caption := ""
		if msg.Content.Caption != nil {
			caption = *msg.Content.Caption
		}
		waMsg = BuildImageMessage(uploadResult, caption)

	case entity.MessageTypeDocument:
		if mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
		if msg.Content.DocURL == nil || *msg.Content.DocURL == "" {
			return errors.ErrEmptyContent.WithMessage("document URL is required")
		}
		filename := ""
		if msg.Content.Caption != nil {
			filename = *msg.Content.Caption // Use caption as filename for documents
		}
		uploadResult, err := mediaUploader.UploadDocument(ctx, msg.SessionID, *msg.Content.DocURL, filename)
		if err != nil {
			return errors.ErrMediaUploadFailed.WithCause(err)
		}
		caption := ""
		if msg.Content.Caption != nil {
			caption = *msg.Content.Caption
		}
		waMsg = BuildDocumentMessage(uploadResult, filename, caption)

	case entity.MessageTypeAudio:
		if mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
		if msg.Content.AudioURL == nil || *msg.Content.AudioURL == "" {
			return errors.ErrEmptyContent.WithMessage("audio URL is required")
		}
		uploadResult, err := mediaUploader.UploadAudio(ctx, msg.SessionID, *msg.Content.AudioURL)
		if err != nil {
			return errors.ErrMediaUploadFailed.WithCause(err)
		}
		waMsg = BuildAudioMessage(uploadResult)

	case entity.MessageTypeVideo:
		if mediaUploader == nil {
			return errors.ErrMediaUploadFailed.WithMessage("media uploader not available")
		}
		if msg.Content.VideoURL == nil || *msg.Content.VideoURL == "" {
			return errors.ErrEmptyContent.WithMessage("video URL is required")
		}
		uploadResult, err := mediaUploader.UploadVideo(ctx, msg.SessionID, *msg.Content.VideoURL)
		if err != nil {
			return errors.ErrMediaUploadFailed.WithCause(err)
		}
		caption := ""
		if msg.Content.Caption != nil {
			caption = *msg.Content.Caption
		}
		waMsg = BuildVideoMessage(uploadResult, caption)

	default:
		// Text message
		waMsg, err = BuildTextMessage(msg)
		if err != nil {
			return err
		}
	}

	// Send message with retry
	_, err = c.sendWithRetry(ctx, client, recipientJID, waMsg)
	if err != nil {
		return errors.ErrMessageSendFailed.WithCause(err)
	}

	return nil
}

// sendWithRetry sends a message with exponential backoff retry
func (c *WhatsmeowClient) sendWithRetry(ctx context.Context, client *whatsmeow.Client, to types.JID, msg *waE2E.Message) (whatsmeow.SendResponse, error) {
	retryPolicy := NewRetryPolicy(RetryConfig{
		MaxAttempts:  3,
		InitialDelay: c.config.ReconnectDelay,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	})

	result, err := retryPolicy.ExecuteWithResult(ctx, func() (any, error) {
		return client.SendMessage(ctx, to, msg)
	})

	if err != nil {
		return whatsmeow.SendResponse{}, err
	}
	return result.(whatsmeow.SendResponse), nil
}

// SendReaction sends a reaction to a message
func (c *WhatsmeowClient) SendReaction(ctx context.Context, sessionID, chatJID, messageID, emoji string) error {
	// Use circuit breaker if enabled
	if c.circuitBreaker != nil {
		_, err := c.circuitBreaker.Execute(ctx, func() (any, error) {
			return nil, c.sendReactionInternal(ctx, sessionID, chatJID, messageID, emoji)
		})
		return err
	}
	return c.sendReactionInternal(ctx, sessionID, chatJID, messageID, emoji)
}

// sendReactionInternal performs the actual reaction sending logic
func (c *WhatsmeowClient) sendReactionInternal(ctx context.Context, sessionID, chatJID, messageID, emoji string) error {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return errors.ErrDisconnected
	}

	// Parse chat JID
	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return errors.ErrInvalidInput.WithMessage("invalid chat JID").WithCause(err)
	}

	// Build reaction message
	reactionMsg := BuildReactionMessage(chatJID, messageID, emoji)

	// Send reaction with retry
	_, err = c.sendWithRetry(ctx, client, jid, reactionMsg)
	if err != nil {
		return errors.ErrMessageSendFailed.WithCause(err)
	}

	return nil
}

// SendReadReceipt sends read receipts for multiple messages atomically
func (c *WhatsmeowClient) SendReadReceipt(ctx context.Context, sessionID, chatJID string, messageIDs []string) error {
	// Use circuit breaker if enabled
	if c.circuitBreaker != nil {
		_, err := c.circuitBreaker.Execute(ctx, func() (any, error) {
			return nil, c.sendReadReceiptInternal(ctx, sessionID, chatJID, messageIDs)
		})
		return err
	}
	return c.sendReadReceiptInternal(ctx, sessionID, chatJID, messageIDs)
}

// sendReadReceiptInternal performs the actual read receipt sending logic
func (c *WhatsmeowClient) sendReadReceiptInternal(ctx context.Context, sessionID, chatJID string, messageIDs []string) error {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return errors.ErrDisconnected
	}

	// Parse chat JID
	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return errors.ErrInvalidInput.WithMessage("invalid chat JID").WithCause(err)
	}

	// Convert string message IDs to types.MessageID
	msgIDs := make([]types.MessageID, len(messageIDs))
	for i, id := range messageIDs {
		msgIDs[i] = types.MessageID(id)
	}

	// Send read receipts for all messages atomically
	// Whatsmeow's MarkRead method handles multiple message IDs
	err = client.MarkRead(ctx, msgIDs, time.Now(), jid, jid, types.ReceiptTypeRead)
	if err != nil {
		return errors.ErrMessageSendFailed.WithMessage("failed to send read receipts").WithCause(err)
	}

	return nil
}

// SendPresence sends a presence update (typing, paused, online, offline)
func (c *WhatsmeowClient) SendPresence(ctx context.Context, sessionID, chatJID, state string) error {
	// Use circuit breaker if enabled
	if c.circuitBreaker != nil {
		_, err := c.circuitBreaker.Execute(ctx, func() (any, error) {
			return nil, c.sendPresenceInternal(ctx, sessionID, chatJID, state)
		})
		return err
	}
	return c.sendPresenceInternal(ctx, sessionID, chatJID, state)
}

// sendPresenceInternal performs the actual presence sending logic
func (c *WhatsmeowClient) sendPresenceInternal(ctx context.Context, sessionID, chatJID, state string) error {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return errors.ErrDisconnected
	}

	// Parse chat JID
	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return errors.ErrInvalidInput.WithMessage("invalid chat JID").WithCause(err)
	}

	// Map state string to whatsmeow presence type
	var presenceType types.Presence
	switch state {
	case "typing":
		presenceType = types.PresenceAvailable // General presence
	case "paused":
		presenceType = types.PresenceAvailable
	case "online":
		presenceType = types.PresenceAvailable
	case "offline":
		presenceType = types.PresenceUnavailable
	default:
		return errors.ErrInvalidInput.WithMessage("invalid presence state")
	}

	// Send general presence update
	err = client.SendPresence(ctx, presenceType)
	if err != nil {
		return errors.ErrMessageSendFailed.WithMessage("failed to send presence").WithCause(err)
	}

	// If chat-specific presence (typing/paused), send to specific chat
	if state == "typing" || state == "paused" {
		var chatPresence types.ChatPresence
		if state == "typing" {
			chatPresence = types.ChatPresenceComposing
		} else {
			chatPresence = types.ChatPresencePaused
		}

		err = client.SendChatPresence(ctx, jid, chatPresence, types.ChatPresenceMediaText)
		if err != nil {
			return errors.ErrMessageSendFailed.WithMessage("failed to send chat presence").WithCause(err)
		}
	}

	return nil
}
