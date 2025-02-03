// internal/service/url_service.go
package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/middleware/logger"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/utils"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
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

func (s *URLService) CreateShortURL(ctx *gin.Context, userID uint, req *dto.CreateShortURLRequest) (*model.URL, error) {
	log := logger.GetLogger(ctx)
	var expiresAt *time.Time
	t := time.Now().AddDate(0, 0, req.ExpiresDays)
	expiresAt = &t
	var err error
	if req.Alias == "" {
		req.Alias, err = utils.GenerateShortCode()
	}

	if err != nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("Failed to hash password: %v", err)
		return nil, err
	}

	url := &model.URL{
		UserID:    userID,
		LongURL:   req.LongURL,
		ExpiresAt: expiresAt,
		ShortCode: req.Alias,
		Password:  string(hashedPassword),
	}

	err = s.urlRepo.Create(url)
	if err != nil {
		log.Errorf("CreateShortURL err: %v", err)
		return nil, err
	}

	return url, nil
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
		var err error
		if request.Alias == "" {
			request.Alias, err = utils.GenerateShortCode()
		}
		if err != nil {
			return nil, err
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Errorf("Failed to hash password: %v", err)
			return nil, err
		}
		urls = append(urls, &model.URL{
			UserID:    userID,
			LongURL:   request.LongURL,
			ExpiresAt: expiresAt,
			ShortCode: request.Alias,
			Password:  string(hashedPassword),
		})
	}
	err := s.urlRepo.CreateBulk(urls)
	if err != nil {
		log.Errorf("CreateShortURLs err: %v", err)
		return nil, err
	}
	return urls, nil
}

func (s *URLService) GetLongURL(ctx *gin.Context, shortCode string) (*model.URL, error) {
	log := logger.GetLogger(ctx)
	url, err := s.urlRepo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, errors.New("url has expired")
	}

	err = s.urlRepo.IncrementClicks(shortCode)
	if err != nil {
		log.Errorf("increment clicks failed: %v", err)
		return nil, err
	}

	return url, nil
}

func (s *URLService) GetUserURLs(userID uint) ([]model.URL, error) {
	return s.urlRepo.FindByUserID(userID)
}

func (s *URLService) GenerateQRCodeBase64(ctx *gin.Context, shortCode string) (string, error) {
	log := logger.GetLogger(ctx)
	url := "https://localhost:8080/api/v1/url/" + shortCode
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		log.Errorf("GenerateQRCode err: %v", err)
		return "", err
	}
	encodedPNG := base64.StdEncoding.EncodeToString(png)

	return fmt.Sprintf("data:image/png;base64,%s", encodedPNG), nil

}
