package whatsapp

import (
	"context"
	"fmt"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// MessageHandler handles incoming WhatsApp messages with media download support
type MessageHandler struct {
	messageParser      *MessageParser
	mediaDownloader    *MediaDownloadHelper
	mediaStorage       repository.MediaStorage
	reactionHandler    *ReactionHandler
	logger             waLog.Logger
	eventQueue         *EventQueue
	sessionConnections map[string]bool // Track session connection status
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	messageParser *MessageParser,
	mediaDownloader *MediaDownloadHelper,
	mediaStorage repository.MediaStorage,
	logger waLog.Logger,
) *MessageHandler {
	return &MessageHandler{
		messageParser:      messageParser,
		mediaDownloader:    mediaDownloader,
		mediaStorage:       mediaStorage,
		logger:             logger,
		eventQueue:         NewEventQueue(),
		sessionConnections: make(map[string]bool),
	}
}

// SetReactionHandler sets the reaction handler for processing reactions
func (h *MessageHandler) SetReactionHandler(handler *ReactionHandler) {
	h.reactionHandler = handler
}

// HandleIncomingMessage processes an incoming WhatsApp message
// It parses the message, downloads media if present, and returns a domain event
func (h *MessageHandler) HandleIncomingMessage(
	ctx context.Context,
	sessionID string,
	client *whatsmeow.Client,
	msg *events.Message,
) (*entity.Event, error) {
	// Save pushname to contact store if available
	if msg.Info.PushName != "" && client != nil && client.Store != nil && client.Store.Contacts != nil {
		_, _, err := client.Store.Contacts.PutPushName(ctx, msg.Info.Sender, msg.Info.PushName)
		if err != nil {
			h.logger.Warnf("Failed to save pushname for %s: %v", msg.Info.Sender.String(), err)
		}
	}

	// Parse the message
	parsedMsg, err := h.messageParser.ParseRealtimeMessage(sessionID, msg)
	if err != nil {
		h.logger.Warnf("Failed to parse message: %v", err)
		return nil, err
	}

	// Resolve LID to phone number JID if needed
	if msg.Info.Sender.Server == "lid" && client != nil && client.Store != nil && client.Store.LIDs != nil {
		pnJID, err := client.Store.LIDs.GetPNForLID(ctx, msg.Info.Sender)
		if err == nil && !pnJID.IsEmpty() {
			parsedMsg.SenderJID = pnJID.String()
			h.logger.Debugf("Resolved LID %s to PN %s", msg.Info.Sender.String(), pnJID.String())
		}
	}

	// Download and store media if this is a media message
	if h.isMediaMessage(parsedMsg) && h.mediaDownloader != nil && h.mediaStorage != nil {
		if err := h.downloadAndStoreMedia(ctx, sessionID, client, msg, parsedMsg); err != nil {
			h.logger.Warnf("Failed to download media for message %s: %v", parsedMsg.MessageID, err)
			// Continue processing - we still want to emit the event even if media download fails
		}
	}

	// Handle reactions separately if reaction handler is available
	if parsedMsg.MessageType == ParsedMessageTypeReaction && h.reactionHandler != nil {
		if err := h.reactionHandler.HandleIncomingReaction(ctx, sessionID, parsedMsg); err != nil {
			h.logger.Warnf("Failed to handle reaction: %v", err)
		}
		// Don't return a message.received event for reactions
		// The reaction handler publishes message.reaction events
		return nil, nil
	}

	// Create the event
	event, err := entity.NewEventWithPayload(
		generateEventID(),
		entity.EventTypeMessageReceived,
		sessionID,
		parsedMsg,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// isMediaMessage checks if the parsed message contains media
func (h *MessageHandler) isMediaMessage(msg *ParsedMessage) bool {
	switch msg.MessageType {
	case ParsedMessageTypeImage, ParsedMessageTypeVideo, ParsedMessageTypeAudio,
		ParsedMessageTypeDocument, ParsedMessageTypeSticker:
		return true
	default:
		return false
	}
}

// downloadAndStoreMedia downloads media from WhatsApp and stores it locally
func (h *MessageHandler) downloadAndStoreMedia(
	ctx context.Context,
	sessionID string,
	client *whatsmeow.Client,
	msg *events.Message,
	parsedMsg *ParsedMessage,
) error {
	var filePath, publicURL string
	var err error

	// Use the appropriate download method based on message type
	switch parsedMsg.MessageType {
	case ParsedMessageTypeImage:
		if msg.Message.GetImageMessage() != nil {
			filePath, publicURL, err = h.mediaDownloader.DownloadAndStoreImage(
				ctx, client, sessionID, parsedMsg.MessageID, msg.Message.GetImageMessage(),
			)
		}
	case ParsedMessageTypeVideo:
		if msg.Message.GetVideoMessage() != nil {
			filePath, publicURL, err = h.mediaDownloader.DownloadAndStoreVideo(
				ctx, client, sessionID, parsedMsg.MessageID, msg.Message.GetVideoMessage(),
			)
		}
	case ParsedMessageTypeAudio:
		if msg.Message.GetAudioMessage() != nil {
			filePath, publicURL, err = h.mediaDownloader.DownloadAndStoreAudio(
				ctx, client, sessionID, parsedMsg.MessageID, msg.Message.GetAudioMessage(),
			)
		}
	case ParsedMessageTypeDocument:
		if msg.Message.GetDocumentMessage() != nil {
			filePath, publicURL, err = h.mediaDownloader.DownloadAndStoreDocument(
				ctx, client, sessionID, parsedMsg.MessageID, msg.Message.GetDocumentMessage(),
			)
		}
	case ParsedMessageTypeSticker:
		if msg.Message.GetStickerMessage() != nil {
			filePath, publicURL, err = h.mediaDownloader.DownloadAndStoreSticker(
				ctx, client, sessionID, parsedMsg.MessageID, msg.Message.GetStickerMessage(),
			)
		}
	default:
		return fmt.Errorf("unsupported media type: %s", parsedMsg.MessageType)
	}

	if err != nil {
		return fmt.Errorf("failed to download and store media: %w", err)
	}

	// Update the parsed message with the media URL
	parsedMsg.MediaURL = &publicURL

	h.logger.Infof("Downloaded and stored media for message %s: %s (local: %s)",
		parsedMsg.MessageID, publicURL, filePath)
	return nil
}

// SetSessionConnected updates the connection status for a session
func (h *MessageHandler) SetSessionConnected(sessionID string, connected bool) {
	h.sessionConnections[sessionID] = connected

	// If session reconnected, flush queued events
	if connected {
		h.eventQueue.FlushSession(sessionID)
	}
}

// QueueEvent queues an event for a disconnected session
func (h *MessageHandler) QueueEvent(event *entity.Event) {
	h.eventQueue.Enqueue(event)
}

// GetQueuedEvents returns all queued events for a session
func (h *MessageHandler) GetQueuedEvents(sessionID string) []*entity.Event {
	return h.eventQueue.GetSessionEvents(sessionID)
}

// IsSessionConnected checks if a session is connected
func (h *MessageHandler) IsSessionConnected(sessionID string) bool {
	return h.sessionConnections[sessionID]
}
