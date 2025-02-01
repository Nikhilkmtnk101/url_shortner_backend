package repository

import (
	"fmt"
	"github.com/gin-gonic/gin"
	common_constants "github.com/nikhil/url-shortner-backend/constants"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/pkg/redis"
)

// SessionRepository handles all database operations related to sessions
type SessionRepository struct {
	cache redis.CacheClient
}

// NewSessionRepository creates a new instance of SessionRepository
func NewSessionRepository(cache redis.CacheClient) *SessionRepository {
	return &SessionRepository{
		cache: cache,
	}
}

func (s *SessionRepository) getCacheKey(userID uint) string {
	return fmt.Sprintf("session:%d", userID)
}

func (s *SessionRepository) IsUserSessionIsValid(ctx *gin.Context, userID uint) error {
	var session model.Session
	return s.cache.GetWithUnmarshal(ctx, s.getCacheKey(userID), &session)
}

func (s *SessionRepository) GetUserSession(ctx *gin.Context, userID uint) (*model.Session, error) {
	var session model.Session
	err := s.cache.GetWithUnmarshal(ctx, s.getCacheKey(userID), &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionRepository) UpdateUserSession(ctx *gin.Context, userID uint, session *model.Session) error {
	return s.cache.Set(ctx, s.getCacheKey(userID), session, common_constants.UserSessionTimeout)
}

func (s *SessionRepository) DeleteUserSession(ctx *gin.Context, userID uint) error {
	return s.cache.Delete(ctx, s.getCacheKey(userID))
}
