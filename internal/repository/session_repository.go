package repository

import (
	"errors"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"gorm.io/gorm"
	"time"
)

// SessionRepository handles all database operations related to sessions
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new instance of SessionRepository
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// CreateSession creates a new session in the database
func (r *SessionRepository) CreateSession(session *model.Session) error {
	// Begin transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Deactivate any existing sessions with the same refresh token
	if err := tx.Model(&model.Session{}).
		Where("refresh_token = ?", session.RefreshToken).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create new session
	if err := tx.Create(session).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// FindSessionByRefreshToken finds a session by refresh token
func (r *SessionRepository) FindSessionByRefreshToken(refreshToken string) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("refresh_token = ? AND is_active = ?", refreshToken, true).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// GetActiveSessions gets all active sessions for a user
func (r *SessionRepository) GetActiveSessions(userID uint) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Where("user_id = ? AND is_active = ? AND expires_at > ?",
		userID, true, time.Now()).Find(&sessions).Error
	return sessions, err
}

// UpdateSession updates an existing session
func (r *SessionRepository) UpdateSession(session *model.Session) error {
	return r.db.Model(session).Updates(map[string]interface{}{
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"last_used_at":  session.LastUsedAt,
		"expires_at":    session.ExpiresAt,
	}).Error
}

// DeactivateSession deactivates a specific session
func (r *SessionRepository) DeactivateSession(sessionID uint) error {
	return r.db.Model(&model.Session{}).
		Where("id = ?", sessionID).
		Update("is_active", false).Error
}

// DeactivateAllUserSessions deactivates all sessions for a user except the current one
func (r *SessionRepository) DeactivateAllUserSessions(userID uint, exceptSessionID uint) error {
	return r.db.Model(&model.Session{}).
		Where("user_id = ? AND id != ? AND is_active = ?", userID, exceptSessionID, true).
		Update("is_active", false).Error
}

// CleanupExpiredSessions removes expired sessions
func (r *SessionRepository) CleanupExpiredSessions() error {
	return r.db.Where("expires_at < ? OR is_active = ?", time.Now(), false).
		Delete(&model.Session{}).Error
}

// GetSessionsByIP gets all active sessions from a specific IP
func (r *SessionRepository) GetSessionsByIP(ip string) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Where("ip = ? AND is_active = ?", ip, true).Find(&sessions).Error
	return sessions, err
}

// GetUserLastSession gets the last active session for a user
func (r *SessionRepository) GetUserLastSession(userID uint) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used_at DESC").
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active session found")
		}
		return nil, err
	}
	return &session, nil
}
