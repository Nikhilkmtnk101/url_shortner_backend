package models

import "time"

type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	IP        string    `gorm:"not null"`
	UserAgent string    `gorm:"not null"`
	Status    string    `gorm:"not null"` // success or failure
	Reason    string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
}
