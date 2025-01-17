package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/config"
	"github.com/nikhil/url-shortner-backend/internal/database"
	"github.com/nikhil/url-shortner-backend/internal/handler"
	"github.com/nikhil/url-shortner-backend/internal/middleware"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/service"
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

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)

	authService := service.NewAuthService(userRepo, sessionRepo, loginAttemptRepo, a.cfg.AccessJWTSecret, a.cfg.RefreshJWTSecret)
	userService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Public routes
	routerGroup := a.router.Group("/api/v1")
	routerGroup.POST("/signup", authHandler.SignUp)
	routerGroup.POST("/login", authHandler.Login)

	// Protected routes
	protected := a.router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(a.cfg.AccessJWTSecret))
	{
		protected.GET("/users", userHandler.GetUsers)
		protected.GET("/user/:id", userHandler.GetUser)
	}

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
