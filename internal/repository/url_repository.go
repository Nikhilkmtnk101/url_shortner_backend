package repository

import (
	"errors"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type URLRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(url *model.URL) error {
	return r.db.Create(url).Error
}

func (r *URLRepository) CreateBulk(urls []*model.URL) error {
	return r.db.Create(urls).Error
}

func (r *URLRepository) FindByShortCode(shortCode string) (*model.URL, error) {
	var url model.URL
	err := r.db.Where("short_code = ?", shortCode).First(&url).Error
	return &url, err
}

func (r *URLRepository) FindByUserID(userID uint) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.Where("user_id = ?", userID).Find(&urls).Error
	return urls, err
}

func (r *URLRepository) IncrementClicks(shortCode string) error {
	// Start a transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		var url model.URL

		// Lock the row for update
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("short_code = ?", shortCode).
			First(&url).Error; err != nil {
			return err
		}

		// Perform the increment
		result := tx.Model(&url).
			Where("short_code = ?", shortCode).
			Updates(map[string]interface{}{
				"clicks":     gorm.Expr("clicks + ?", 1),
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("failed to increment clicks: no rows affected")
		}

		return nil
	})
}
