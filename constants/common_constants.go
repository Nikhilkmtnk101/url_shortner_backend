package common_constants

import "time"

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

const (
	OTPCacheTimeOut        time.Duration = 5 * time.Minute
	UserSignupCacheTimeout time.Duration = 5 * time.Minute
	UserSessionTimeout     time.Duration = 7 * 24 * time.Hour
)
