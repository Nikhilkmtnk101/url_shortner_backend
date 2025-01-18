package model

import (
	"github.com/nikhil/url-shortner-backend/constants"
	"time"
)

type User struct {
	ID        uint                      `json:"id" gorm:"primaryKey"`
	Email     string                    `json:"email" gorm:"unique;not null"`
	Password  string                    `json:"-" gorm:"not null"` // '-' prevents password from being shown in JSON
	Name      string                    `json:"name" gorm:"not null"`
	UserRole  common_constants.UserRole `json:"user_role" gorm:"not null"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}
