package repository

import (
	"errors"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/internal/utils"
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

func (r *URLRepository) Create(url *model.URL) (*model.URL, error) {
	var createdURL model.URL

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// First create the URL to get the ID
		if err := tx.Create(url).Error; err != nil {
			return err
		}

		// Generate short code from the ID
		shortCode := utils.GenerateURLID(url.ID, 6)

		// Update the URL with the generated short code
		if err := tx.Model(url).Update("short_code", shortCode).Error; err != nil {
			return err
		}

		// Fetch the complete model
		if err := tx.First(&createdURL, url.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &createdURL, nil
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
