package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/config"
	"github.com/nikhil/url-shortner-backend/internal/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	cfg    *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		router: gin.Default(),
		cfg:    cfg,
	}
}

func (a *App) Run() {
	db := a.setupDatabase()
	a.setupRoutes(db)
	a.router.Run(":" + a.cfg.ServerPort)
}

func (a *App) setupDatabase() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		a.cfg.DBHost,
		a.cfg.DBUser,
		a.cfg.DBPassword,
		a.cfg.DBName,
		a.cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	return db
}
