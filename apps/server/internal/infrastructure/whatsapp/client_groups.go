package whatsapp

import (
	"context"
	"strings"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// GetJoinedGroups fetches all groups the session is a member of from WhatsApp
func (c *WhatsmeowClient) GetJoinedGroups(ctx context.Context, sessionID string) ([]*entity.Group, error) {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return nil, errors.ErrDisconnected
	}

	// Fetch joined groups from WhatsApp
	waGroups, err := client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, errors.ErrInternal.WithCause(err).WithMessage("failed to fetch groups from WhatsApp")
	}

	groups := make([]*entity.Group, 0, len(waGroups))
	now := time.Now()

	for _, waGroup := range waGroups {
		// Fetch chat settings (archived/muted) from whatsmeow store
		isArchived := false
		isMuted := false
		var mutedUntil *time.Time

		if client.Store != nil && client.Store.ChatSettings != nil {
			chatSettings, err := client.Store.ChatSettings.GetChatSettings(ctx, waGroup.JID)
			if err == nil {
				isArchived = chatSettings.Archived
				isMuted = chatSettings.MutedUntil != time.Time{}
				if isMuted {
					mutedUntil = &chatSettings.MutedUntil
				}
			}
		}

		group := &entity.Group{
			JID:         waGroup.JID.String(),
			Name:        waGroup.Name,
			SessionID:   sessionID,
			MemberCount: len(waGroup.Participants),
			IsAnnounce:  waGroup.IsAnnounce,
			IsLocked:    waGroup.IsLocked,
			IsArchived:  isArchived,
			IsMuted:     isMuted,
			MutedUntil:  mutedUntil,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Set optional fields
		if waGroup.Topic != "" {
			group.Description = &waGroup.Topic
		}

		// Resolve owner JID from LID if needed
		if !waGroup.OwnerJID.IsEmpty() {
			ownerJID := c.resolveJID(ctx, client, waGroup.OwnerJID)
			group.OwnerJID = &ownerJID
		}

		if !waGroup.GroupCreated.IsZero() {
			group.GroupCreatedAt = &waGroup.GroupCreated
		}

		// Convert participants
		participants := make([]entity.Participant, 0, len(waGroup.Participants))
		for _, waParticipant := range waGroup.Participants {
			role := convertParticipantRole(waParticipant)

			// Resolve participant JID from LID if needed
			resolvedJID := c.resolveJID(ctx, client, waParticipant.JID)

			participant := entity.Participant{
				JID:       resolvedJID,
				Role:      role,
				JoinedAt:  now,
				UpdatedAt: now,
			}

			// Try to get display name from multiple sources
			displayName := c.getParticipantDisplayName(ctx, client, waParticipant)
			if displayName != "" {
				participant.DisplayName = &displayName
			}

			participants = append(participants, participant)
		}
		group.Participants = participants

		groups = append(groups, group)
	}

	return groups, nil
}

// resolveJID resolves a JID, converting LID to phone number if needed
func (c *WhatsmeowClient) resolveJID(ctx context.Context, client *whatsmeow.Client, jid types.JID) string {
	// If it's a LID (Linked ID), resolve it to a phone number
	if jid.Server == types.DefaultUserServer || jid.Server == "lid" {
		// Try to resolve LID to phone number
		if client.Store != nil && client.Store.LIDs != nil {
			pn, err := client.Store.LIDs.GetPNForLID(ctx, jid)
			if err == nil && !pn.IsEmpty() {
				return pn.String()
			}
		}
	}

	// Return the original JID string if resolution fails or not needed
	return jid.String()
}

// convertParticipantRole converts whatsmeow participant to domain role
func convertParticipantRole(p types.GroupParticipant) entity.ParticipantRole {
	if p.IsSuperAdmin {
		return entity.ParticipantRoleSuperAdmin
	}
	if p.IsAdmin {
		return entity.ParticipantRoleAdmin
	}
	return entity.ParticipantRoleMember
}

// getParticipantDisplayName tries to get the display name from multiple sources:
// 1. GroupParticipant.DisplayName (from group metadata)
// 2. Contact store (pushname from previous interactions)
// 3. Returns empty string if not found
func (c *WhatsmeowClient) getParticipantDisplayName(ctx context.Context, client *whatsmeow.Client, p types.GroupParticipant) string {
	// First, check if DisplayName is set in the group participant data
	if p.DisplayName != "" {
		return p.DisplayName
	}

	// Try to get pushname from the contact store
	if client.Store != nil && client.Store.Contacts != nil {
		contact, err := client.Store.Contacts.GetContact(ctx, p.JID)
		if err == nil && contact.PushName != "" {
			return contact.PushName
		}
		// Also check FullName or BusinessName as fallbacks
		if err == nil && contact.FullName != "" {
			return contact.FullName
		}
		if err == nil && contact.BusinessName != "" {
			return contact.BusinessName
		}
	}

	return ""
}

// CheckPhoneNumber checks if a phone number is registered on WhatsApp
func (c *WhatsmeowClient) CheckPhoneNumber(ctx context.Context, sessionID, phone string) (*entity.Contact, error) {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return nil, errors.ErrDisconnected
	}

	// Clean phone number (remove + and spaces)
	cleanPhone := strings.ReplaceAll(strings.TrimPrefix(phone, "+"), " ", "")

	// Check if number is on WhatsApp
	isOnWhatsApp, err := client.IsOnWhatsApp(ctx, []string{cleanPhone})
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to check phone number").WithCause(err)
	}

	if len(isOnWhatsApp) == 0 {
		return nil, errors.ErrInternal.WithMessage("no result from WhatsApp")
	}

	result := isOnWhatsApp[0]
	contact := entity.NewContact(result.JID.String(), phone, result.IsIn)

	return contact, nil
}

// GetUserProfile retrieves the profile information for a user
func (c *WhatsmeowClient) GetUserProfile(ctx context.Context, sessionID, jid string) (*entity.Contact, error) {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return nil, errors.ErrDisconnected
	}

	// Parse JID
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, errors.ErrInvalidInput.WithMessage("invalid JID").WithCause(err)
	}

	// Get profile picture
	var avatarURL string
	pic, err := client.GetProfilePictureInfo(ctx, parsedJID, &whatsmeow.GetProfilePictureParams{})
	if err == nil && pic != nil {
		avatarURL = pic.URL
	}

	// Get status/about
	var status string
	// Note: whatsmeow doesn't have a direct method to get user status
	// We'll leave it empty for now

	// Get display name from contact store
	contactInfo, err := client.Store.Contacts.GetContact(ctx, parsedJID)
	displayName := jid
	if err == nil && contactInfo.FullName != "" {
		displayName = contactInfo.FullName
	} else if err == nil && contactInfo.PushName != "" {
		displayName = contactInfo.PushName
	}

	contact := entity.NewContact(jid, displayName, true)
	if avatarURL != "" {
		contact.SetAvatar(avatarURL)
	}
	if status != "" {
		contact.SetStatus(status)
	}

	return contact, nil
}

// ListContacts retrieves all contacts for a session
func (c *WhatsmeowClient) ListContacts(ctx context.Context, sessionID string) ([]*entity.Contact, error) {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return nil, errors.ErrDisconnected
	}

	// Get all contacts from the store
	allContacts, err := client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to get contacts").WithCause(err)
	}

	contacts := make([]*entity.Contact, 0, len(allContacts))
	for jid, contactInfo := range allContacts {
		// Skip group JIDs
		if jid.Server == types.GroupServer || jid.Server == types.BroadcastServer {
			continue
		}

		displayName := contactInfo.FullName
		if displayName == "" {
			displayName = contactInfo.PushName
		}
		if displayName == "" {
			displayName = jid.User
		}

		contact := entity.NewContact(jid.String(), displayName, true)
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// ListChats retrieves all chats for a session
func (c *WhatsmeowClient) ListChats(ctx context.Context, sessionID string) ([]*entity.Chat, error) {
	c.mu.RLock()
	client, exists := c.clients[sessionID]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.ErrSessionNotFound
	}

	if !client.IsConnected() {
		return nil, errors.ErrDisconnected
	}

	// Get all contacts and create chats from them
	allContacts, err := client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to get chats").WithCause(err)
	}

	chats := make([]*entity.Chat, 0, len(allContacts))
	for jid, contactInfo := range allContacts {
		// Determine if it's a group
		isGroup := jid.Server == types.GroupServer

		// Get display name
		displayName := contactInfo.FullName
		if displayName == "" {
			displayName = contactInfo.PushName
		}
		if displayName == "" {
			displayName = jid.User
		}

		// For groups, try to get group info
		if isGroup {
			groupInfo, err := client.GetGroupInfo(ctx, jid)
			if err == nil && groupInfo != nil {
				displayName = groupInfo.Name
			}
		}

		chat := entity.NewChat(jid.String(), displayName, isGroup)

		// Note: ChatSettings API varies by whatsmeow version
		// We'll set basic defaults for now
		chat.SetUnreadCount(0)

		chats = append(chats, chat)
	}

	return chats, nil
}
