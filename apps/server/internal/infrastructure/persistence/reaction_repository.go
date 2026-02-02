package persistence

import (
	"context"
	"sync"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
)

// InMemoryReactionRepository implements ReactionRepository with in-memory storage
type InMemoryReactionRepository struct {
	reactions        map[string]*entity.Reaction // ID -> Reaction
	messageReactions map[string][]string         // MessageID -> []ReactionID
	sessionReactions map[string][]string         // SessionID -> []ReactionID
	mu               sync.RWMutex
}

// NewInMemoryReactionRepository creates a new in-memory reaction repository
func NewInMemoryReactionRepository() *InMemoryReactionRepository {
	return &InMemoryReactionRepository{
		reactions:        make(map[string]*entity.Reaction),
		messageReactions: make(map[string][]string),
		sessionReactions: make(map[string][]string),
	}
}

// Save stores a reaction in the repository
func (r *InMemoryReactionRepository) Save(ctx context.Context, reaction *entity.Reaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy to prevent external mutation
	reactionCopy := *reaction
	r.reactions[reaction.ID] = &reactionCopy

	// Index by message ID
	if _, exists := r.messageReactions[reaction.MessageID]; !exists {
		r.messageReactions[reaction.MessageID] = []string{}
	}
	// Check if already indexed
	found := false
	for _, id := range r.messageReactions[reaction.MessageID] {
		if id == reaction.ID {
			found = true
			break
		}
	}
	if !found {
		r.messageReactions[reaction.MessageID] = append(r.messageReactions[reaction.MessageID], reaction.ID)
	}

	// Index by session ID
	if _, exists := r.sessionReactions[reaction.SessionID]; !exists {
		r.sessionReactions[reaction.SessionID] = []string{}
	}
	// Check if already indexed
	found = false
	for _, id := range r.sessionReactions[reaction.SessionID] {
		if id == reaction.ID {
			found = true
			break
		}
	}
	if !found {
		r.sessionReactions[reaction.SessionID] = append(r.sessionReactions[reaction.SessionID], reaction.ID)
	}

	return nil
}

// FindByMessageID retrieves all reactions for a specific message
func (r *InMemoryReactionRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reactionIDs, exists := r.messageReactions[messageID]
	if !exists {
		return []*entity.Reaction{}, nil
	}

	reactions := make([]*entity.Reaction, 0, len(reactionIDs))
	for _, id := range reactionIDs {
		if reaction, exists := r.reactions[id]; exists {
			reactionCopy := *reaction
			reactions = append(reactions, &reactionCopy)
		}
	}

	return reactions, nil
}

// FindBySessionID retrieves all reactions for a specific session with pagination
func (r *InMemoryReactionRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reactionIDs, exists := r.sessionReactions[sessionID]
	if !exists {
		return []*entity.Reaction{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(reactionIDs) {
		return []*entity.Reaction{}, nil
	}

	end := start + limit
	if end > len(reactionIDs) {
		end = len(reactionIDs)
	}

	reactions := make([]*entity.Reaction, 0, end-start)
	for i := start; i < end; i++ {
		if reaction, exists := r.reactions[reactionIDs[i]]; exists {
			reactionCopy := *reaction
			reactions = append(reactions, &reactionCopy)
		}
	}

	return reactions, nil
}

// Delete removes a reaction by its ID
func (r *InMemoryReactionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reaction, exists := r.reactions[id]
	if !exists {
		return errors.ErrNotFound
	}

	// Remove from message index
	if reactionIDs, exists := r.messageReactions[reaction.MessageID]; exists {
		newIDs := make([]string, 0, len(reactionIDs))
		for _, rid := range reactionIDs {
			if rid != id {
				newIDs = append(newIDs, rid)
			}
		}
		r.messageReactions[reaction.MessageID] = newIDs
	}

	// Remove from session index
	if reactionIDs, exists := r.sessionReactions[reaction.SessionID]; exists {
		newIDs := make([]string, 0, len(reactionIDs))
		for _, rid := range reactionIDs {
			if rid != id {
				newIDs = append(newIDs, rid)
			}
		}
		r.sessionReactions[reaction.SessionID] = newIDs
	}

	// Remove from main storage
	delete(r.reactions, id)
	return nil
}

// DeleteByMessageIDAndFrom removes a reaction by message ID and sender
func (r *InMemoryReactionRepository) DeleteByMessageIDAndFrom(ctx context.Context, messageID, from string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reactionIDs, exists := r.messageReactions[messageID]
	if !exists {
		return nil // No reactions for this message
	}

	// Find and delete matching reactions
	for _, id := range reactionIDs {
		if reaction, exists := r.reactions[id]; exists && reaction.From == from {
			// Remove from message index
			newIDs := make([]string, 0, len(reactionIDs))
			for _, rid := range reactionIDs {
				if rid != id {
					newIDs = append(newIDs, rid)
				}
			}
			r.messageReactions[messageID] = newIDs

			// Remove from session index
			if sessionIDs, exists := r.sessionReactions[reaction.SessionID]; exists {
				newSessionIDs := make([]string, 0, len(sessionIDs))
				for _, rid := range sessionIDs {
					if rid != id {
						newSessionIDs = append(newSessionIDs, rid)
					}
				}
				r.sessionReactions[reaction.SessionID] = newSessionIDs
			}

			// Remove from main storage
			delete(r.reactions, id)
		}
	}

	return nil
}

// Clear removes all reactions (for testing)
func (r *InMemoryReactionRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reactions = make(map[string]*entity.Reaction)
	r.messageReactions = make(map[string][]string)
	r.sessionReactions = make(map[string][]string)
}
