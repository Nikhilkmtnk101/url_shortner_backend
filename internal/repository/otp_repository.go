package repository

import (
	"github.com/gin-gonic/gin"
	common_constants "github.com/nikhil/url-shortner-backend/constants"
	"github.com/nikhil/url-shortner-backend/internal/middleware/logger"
	"github.com/nikhil/url-shortner-backend/pkg/redis"
)

type IOTPRepository interface {
	SaveOTP(ctx *gin.Context, email string, otp string) error
	GetOTP(ctx *gin.Context, email string) (string, error)
	DeleteOTP(ctx *gin.Context, email string) error
}
type OTPRepository struct {
	cache redis.CacheClient
}

func NewOTPRepository(cache redis.CacheClient) IOTPRepository {
	return &OTPRepository{
		cache: cache,
	}
}

func getOTPCacheKey(email string) string {
	return "otp:" + email
}

func (o *OTPRepository) SaveOTP(ctx *gin.Context, email string, otp string) error {
	log := logger.GetLogger(ctx)
	cacheKey := getOTPCacheKey(email)
	err := o.cache.Set(ctx, cacheKey, otp, common_constants.OTPCacheTimeOut)
	if err != nil {
		log.Errorf("Failed to set cache key: %s, err: %v", cacheKey, err)
		return err
	}
	return nil
}

func (o *OTPRepository) GetOTP(ctx *gin.Context, email string) (string, error) {
	log := logger.GetLogger(ctx)
	cacheKey := getOTPCacheKey(email)
	value, err := o.cache.Get(ctx, cacheKey)
	if err != nil {
		log.Errorf("Failed to get cache key: %s, err: %v", cacheKey, err)
	}
	return value, nil
}

func (o *OTPRepository) DeleteOTP(ctx *gin.Context, email string) error {
	log := logger.GetLogger(ctx)
	cacheKey := getOTPCacheKey(email)
	err := o.cache.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf("Failed to delete cache key: %s, err: %v", cacheKey, err)
		return err
	}
	return nil
}
