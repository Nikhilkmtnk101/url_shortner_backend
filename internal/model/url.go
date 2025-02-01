package model

import (
	"time"
)

type URL struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id" gorm:"not null"`
	LongURL   string     `json:"long_url" gorm:"not null;type:text"`
	ShortCode string     `json:"short_code" gorm:"uniqueIndex;not null;type:varchar(10)"`
	Clicks    int64      `json:"clicks" gorm:"default:0"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"` // Automatically set when created
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"` // Automatically updated on save
	User      User       `json:"-" gorm:"foreignKey:UserID"`
}
