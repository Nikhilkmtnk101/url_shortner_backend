// internal/service/url_service.go
package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/middleware/logger"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"time"
)

type URLService struct {
	urlRepo *repository.URLRepository
}

func NewURLService(urlRepo *repository.URLRepository) *URLService {
	return &URLService{
		urlRepo: urlRepo,
	}
}

func (s *URLService) CreateShortURL(userID uint, longURL string, expiresDays int) (*model.URL, error) {
	var expiresAt *time.Time
	if expiresDays > 0 {
		t := time.Now().AddDate(0, 0, expiresDays)
		expiresAt = &t
	}

	url := &model.URL{
		UserID:    userID,
		LongURL:   longURL,
		ExpiresAt: expiresAt,
	}

	createdURL, err := s.urlRepo.Create(url)
	if err != nil {
		return nil, err
	}

	return createdURL, nil
}

func (s *URLService) GetLongURL(ctx *gin.Context, shortCode string) (string, error) {
	log := logger.GetLogger(ctx)
	url, err := s.urlRepo.FindByShortCode(shortCode)
	if err != nil {
		return "", err
	}

	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return "", errors.New("url has expired")
	}

	err = s.urlRepo.IncrementClicks(shortCode)
	if err != nil {
		log.Warnf("increment clicks failed: %v", err)
		return "", err
	}

	return url.LongURL, nil
}

func (s *URLService) GetUserURLs(userID uint) ([]model.URL, error) {
	return s.urlRepo.FindByUserID(userID)
}

func generateShortCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)[:6], nil
}
