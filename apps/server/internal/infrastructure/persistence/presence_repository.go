package persistence

import (
	"context"
	"sync"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
)

// InMemoryPresenceRepository implements PresenceRepository with in-memory storage
type InMemoryPresenceRepository struct {
	presences       map[string]*entity.Presence // ID -> Presence
	sessionPresence map[string][]string         // SessionID -> []PresenceID
	userPresence    map[string][]string         // UserJID -> []PresenceID
	latestPresence  map[string]string           // UserJID -> PresenceID (latest)
	mu              sync.RWMutex
}

// NewInMemoryPresenceRepository creates a new in-memory presence repository
func NewInMemoryPresenceRepository() *InMemoryPresenceRepository {
	return &InMemoryPresenceRepository{
		presences:       make(map[string]*entity.Presence),
		sessionPresence: make(map[string][]string),
		userPresence:    make(map[string][]string),
		latestPresence:  make(map[string]string),
	}
}

// Save stores a presence update in the repository
func (r *InMemoryPresenceRepository) Save(ctx context.Context, presence *entity.Presence) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy to prevent external mutation
	presenceCopy := *presence
	r.presences[presence.ID] = &presenceCopy

	// Index by session ID
	if _, exists := r.sessionPresence[presence.SessionID]; !exists {
		r.sessionPresence[presence.SessionID] = []string{}
	}
	// Check if already indexed
	found := false
	for _, id := range r.sessionPresence[presence.SessionID] {
		if id == presence.ID {
			found = true
			break
		}
	}
	if !found {
		r.sessionPresence[presence.SessionID] = append(r.sessionPresence[presence.SessionID], presence.ID)
	}

	// Index by user JID
	if _, exists := r.userPresence[presence.UserJID]; !exists {
		r.userPresence[presence.UserJID] = []string{}
	}
	// Check if already indexed
	found = false
	for _, id := range r.userPresence[presence.UserJID] {
		if id == presence.ID {
			found = true
			break
		}
	}
	if !found {
		r.userPresence[presence.UserJID] = append(r.userPresence[presence.UserJID], presence.ID)
	}

	// Update latest presence for user
	if latestID, exists := r.latestPresence[presence.UserJID]; exists {
		if latestPresence, exists := r.presences[latestID]; exists {
			if presence.Timestamp.After(latestPresence.Timestamp) {
				r.latestPresence[presence.UserJID] = presence.ID
			}
		}
	} else {
		r.latestPresence[presence.UserJID] = presence.ID
	}

	return nil
}

// FindBySessionID retrieves all presence updates for a specific session with pagination
func (r *InMemoryPresenceRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Presence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	presenceIDs, exists := r.sessionPresence[sessionID]
	if !exists {
		return []*entity.Presence{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(presenceIDs) {
		return []*entity.Presence{}, nil
	}

	end := start + limit
	if end > len(presenceIDs) {
		end = len(presenceIDs)
	}

	presences := make([]*entity.Presence, 0, end-start)
	for i := start; i < end; i++ {
		if presence, exists := r.presences[presenceIDs[i]]; exists {
			presenceCopy := *presence
			presences = append(presences, &presenceCopy)
		}
	}

	return presences, nil
}

// FindByUserJID retrieves all presence updates for a specific user with pagination
func (r *InMemoryPresenceRepository) FindByUserJID(ctx context.Context, userJID string, limit, offset int) ([]*entity.Presence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	presenceIDs, exists := r.userPresence[userJID]
	if !exists {
		return []*entity.Presence{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(presenceIDs) {
		return []*entity.Presence{}, nil
	}

	end := start + limit
	if end > len(presenceIDs) {
		end = len(presenceIDs)
	}

	presences := make([]*entity.Presence, 0, end-start)
	for i := start; i < end; i++ {
		if presence, exists := r.presences[presenceIDs[i]]; exists {
			presenceCopy := *presence
			presences = append(presences, &presenceCopy)
		}
	}

	return presences, nil
}

// GetLatestByUserJID retrieves the most recent presence update for a user
func (r *InMemoryPresenceRepository) GetLatestByUserJID(ctx context.Context, userJID string) (*entity.Presence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	latestID, exists := r.latestPresence[userJID]
	if !exists {
		return nil, errors.ErrNotFound
	}

	presence, exists := r.presences[latestID]
	if !exists {
		return nil, errors.ErrNotFound
	}

	// Return a copy to prevent external mutation
	presenceCopy := *presence
	return &presenceCopy, nil
}

// Delete removes a presence update by its ID
func (r *InMemoryPresenceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	presence, exists := r.presences[id]
	if !exists {
		return errors.ErrNotFound
	}

	// Remove from session index
	if presenceIDs, exists := r.sessionPresence[presence.SessionID]; exists {
		newIDs := make([]string, 0, len(presenceIDs))
		for _, pid := range presenceIDs {
			if pid != id {
				newIDs = append(newIDs, pid)
			}
		}
		r.sessionPresence[presence.SessionID] = newIDs
	}

	// Remove from user index
	if presenceIDs, exists := r.userPresence[presence.UserJID]; exists {
		newIDs := make([]string, 0, len(presenceIDs))
		for _, pid := range presenceIDs {
			if pid != id {
				newIDs = append(newIDs, pid)
			}
		}
		r.userPresence[presence.UserJID] = newIDs
	}

	// Update latest presence if this was the latest
	if latestID, exists := r.latestPresence[presence.UserJID]; exists && latestID == id {
		// Find the new latest presence
		if presenceIDs, exists := r.userPresence[presence.UserJID]; exists && len(presenceIDs) > 0 {
			var latestTime time.Time
			var newLatestID string
			for _, pid := range presenceIDs {
				if p, exists := r.presences[pid]; exists {
					if p.Timestamp.After(latestTime) {
						latestTime = p.Timestamp
						newLatestID = pid
					}
				}
			}
			if newLatestID != "" {
				r.latestPresence[presence.UserJID] = newLatestID
			} else {
				delete(r.latestPresence, presence.UserJID)
			}
		} else {
			delete(r.latestPresence, presence.UserJID)
		}
	}

	// Remove from main storage
	delete(r.presences, id)
	return nil
}

// Clear removes all presence updates (for testing)
func (r *InMemoryPresenceRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.presences = make(map[string]*entity.Presence)
	r.sessionPresence = make(map[string][]string)
	r.userPresence = make(map[string][]string)
	r.latestPresence = make(map[string]string)
}
