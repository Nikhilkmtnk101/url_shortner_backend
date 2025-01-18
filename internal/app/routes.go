package app

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/handler"
	"github.com/nikhil/url-shortner-backend/internal/middleware"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"gorm.io/gorm"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (a *App) setupRoutes(db *gorm.DB) {
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	urlRepo := repository.NewURLRepository(db)

	authService := service.NewAuthService(userRepo, sessionRepo, loginAttemptRepo, a.cfg.AccessJWTSecret, a.cfg.RefreshJWTSecret)
	userService := service.NewUserService(userRepo)
	urlService := service.NewURLService(urlRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	urlHandler := handler.NewURLHandler(urlService)

	// Router Groups
	a.router.Use(CORSMiddleware())
	routerGroup := a.router.Group("/api/v1")

	// Auth routes
	authRouterGroup := routerGroup.Group("/auth")
	{
		authRouterGroup.POST("/signup", authHandler.SignUp)
		authRouterGroup.POST("/login", authHandler.Login)
	}

	// URL redirect route (public)
	urlRouterGroup := routerGroup.Group("/url")
	{
		urlRouterGroup.GET("/s/:shortCode", urlHandler.RedirectToLongURL)
	}

	// Protected routes - authentication middleware
	protectedRouterGroup := routerGroup.Group("")
	protectedRouterGroup.Use(middleware.AuthMiddleware(a.cfg.AccessJWTSecret))
	{
		// User routes
		userRouterGroup := protectedRouterGroup.Group("/user")
		{
			userRouterGroup.GET("", userHandler.GetUsers)
			userRouterGroup.GET("/:id", userHandler.GetUser)
		}

		// URL management routes
		urlRouterGroup := protectedRouterGroup.Group("/url")
		{
			urlRouterGroup.POST("", urlHandler.CreateShortURL)
			urlRouterGroup.GET("", urlHandler.GetUserURLs)
		}
	}
}
