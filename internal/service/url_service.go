// internal/service/url_service.go
package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
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

func (s *URLService) CreateShortURL(ctx *gin.Context, userID uint, longURL string, expiresDays int) (*model.URL, error) {
	log := logger.GetLogger(ctx)
	var expiresAt *time.Time
	t := time.Now().AddDate(0, 0, expiresDays)
	expiresAt = &t

	url := &model.URL{
		UserID:    userID,
		LongURL:   longURL,
		ExpiresAt: expiresAt,
	}

	createdURL, err := s.urlRepo.Create(url)
	if err != nil {
		log.Errorf("CreateShortURL err: %v", err)
		return nil, err
	}

	return createdURL, nil
}

func (s *URLService) CreateShortURLs(
	ctx *gin.Context, userID uint, createBulkShortURLsRequest []dto.CreateShortURLRequest,
) ([]*model.URL, error) {
	log := logger.GetLogger(ctx)
	var urls []*model.URL
	for _, request := range createBulkShortURLsRequest {
		var expiresAt *time.Time
		t := time.Now().AddDate(0, 0, request.ExpiresDays)
		expiresAt = &t
		urls = append(urls, &model.URL{
			UserID:    userID,
			LongURL:   request.LongURL,
			ExpiresAt: expiresAt,
		})
	}
	createdURLs, err := s.urlRepo.CreateBulk(urls)
	if err != nil {
		log.Errorf("CreateShortURLs err: %v", err)
		return nil, err
	}
	return createdURLs, nil
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
