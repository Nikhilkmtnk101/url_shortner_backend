package database

import (
	"fmt"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	fmt.Println("Running database migrations...")

	// Add migrations here
	err := db.AutoMigrate(
		&model.User{},
		&model.URL{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}
