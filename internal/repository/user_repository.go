package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/pkg/redis"
	"gorm.io/gorm"
	"time"
)

type UserRepository struct {
	db    *gorm.DB
	cache redis.CacheClient
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB, cache redis.CacheClient) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cache,
	}
}

func (r *UserRepository) getUserCacheKey(email string) string {
	return "user:" + email
}

func (r *UserRepository) SaveUserToCache(ctx *gin.Context, user *model.User, timeout time.Duration) error {
	err := r.cache.Set(ctx, r.getUserCacheKey(user.Email), user, timeout)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetUserFromCache(ctx *gin.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.cache.GetWithUnmarshal(ctx, r.getUserCacheKey(email), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) DeleteUserFromCache(ctx *gin.Context, email string) error {
	return r.cache.Delete(ctx, r.getUserCacheKey(email))
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Find(&users).Error
	return users, err
}
