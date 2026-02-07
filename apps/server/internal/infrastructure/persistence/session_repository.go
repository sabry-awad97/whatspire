package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// SessionRepository implements SessionRepository with GORM
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new GORM session repository
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session in the repository
func (r *SessionRepository) Create(ctx context.Context, session *entity.Session) error {
	model := &models.Session{
		ID:        session.ID,
		Name:      session.Name,
		JID:       session.JID,
		Status:    session.Status.String(),
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return domainErrors.ErrSessionExists
		}
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return nil
}

// GetByID retrieves a session by its ID
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	var model models.Session

	result := r.db.WithContext(ctx).First(&model, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrSessionNotFound
		}
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert model to domain entity
	session := &entity.Session{
		ID:        model.ID,
		Name:      model.Name,
		JID:       model.JID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	session.SetStatus(entity.Status(model.Status))

	return session, nil
}

// GetAll retrieves all sessions
func (r *SessionRepository) GetAll(ctx context.Context) ([]*entity.Session, error) {
	var models []models.Session

	result := r.db.WithContext(ctx).Find(&models)
	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	sessions := make([]*entity.Session, 0, len(models))
	for _, model := range models {
		session := &entity.Session{
			ID:        model.ID,
			Name:      model.Name,
			JID:       model.JID,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		}
		session.SetStatus(entity.Status(model.Status))
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update updates an existing session
func (r *SessionRepository) Update(ctx context.Context, session *entity.Session) error {
	updates := map[string]interface{}{
		"name":       session.Name,
		"jid":        session.JID,
		"status":     session.Status.String(),
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", session.ID).
		Updates(updates)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrSessionNotFound
	}

	return nil
}

// Delete removes a session by its ID
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrSessionNotFound
	}

	return nil
}

// UpdateStatus updates only the status of a session
func (r *SessionRepository) UpdateStatus(ctx context.Context, id string, status entity.Status) error {
	updates := map[string]interface{}{
		"status":     status.String(),
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrSessionNotFound
	}

	return nil
}

// isUniqueConstraintError checks if the error is a SQLite unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "constraint failed")
}
