package repository

import (
	"github.com/nikhil/url-shortner-backend/internal/models"
	"gorm.io/gorm"
	"time"
)

// LoginAttemptRepository handles all database operations related to login attempts
type LoginAttemptRepository struct {
	db *gorm.DB
}

// NewLoginAttemptRepository creates a new instance of LoginAttemptRepository
func NewLoginAttemptRepository(db *gorm.DB) *LoginAttemptRepository {
	return &LoginAttemptRepository{
		db: db,
	}
}

// Create adds a new login attempt to the database
func (r *LoginAttemptRepository) Create(attempt *models.LoginAttempt) error {
	return r.db.Create(attempt).Error
}

// GetRecentFailedAttempts retrieves recent failed login attempts by IP
func (r *LoginAttemptRepository) GetRecentFailedAttempts(ip string, since time.Time) ([]models.LoginAttempt, error) {
	var attempts []models.LoginAttempt
	err := r.db.Where("ip = ? AND status = ? AND created_at >= ?", ip, "failure", since).
		Find(&attempts).Error
	return attempts, err
}

// GetUserRecentFailedAttempts retrieves recent failed login attempts for a specific user
func (r *LoginAttemptRepository) GetUserRecentFailedAttempts(userID uint, since time.Time) ([]models.LoginAttempt, error) {
	var attempts []models.LoginAttempt
	err := r.db.Where("user_id = ? AND status = ? AND created_at >= ?", userID, "failure", since).
		Find(&attempts).Error
	return attempts, err
}

// CleanupOldAttempts deletes old login attempts to keep the table size manageable
func (r *LoginAttemptRepository) CleanupOldAttempts(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&models.LoginAttempt{}).Error
}
