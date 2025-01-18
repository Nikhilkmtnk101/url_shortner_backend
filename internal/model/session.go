package model

import "time"

type Session struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null;index"` // Index for faster queries
	AccessToken  string    `gorm:"type:text;not null"`
	RefreshToken string    `gorm:"type:text;not null;uniqueIndex"` // Unique index for faster lookups
	IP           string    `gorm:"not null"`
	UserAgent    string    `gorm:"type:text;not null"`
	LastUsedAt   time.Time `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null;index"`              // Index for cleanup queries
	IsActive     bool      `gorm:"not null;default:true;index"` // Index for active session queries
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}
