package app

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/handler"
	"github.com/nikhil/url-shortner-backend/internal/middleware"
	"github.com/nikhil/url-shortner-backend/internal/middleware/logger"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"github.com/nikhil/url-shortner-backend/internal/service/email_service"
	"github.com/nikhil/url-shortner-backend/internal/service/otp_service"
	"github.com/nikhil/url-shortner-backend/pkg/redis"
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

func (a *App) setupRoutes(db *gorm.DB, cache redis.CacheClient) {
	userRepo := repository.NewUserRepository(db, cache)
	otpRepo := repository.NewOTPRepository(cache)
	sessionRepo := repository.NewSessionRepository(cache)
	urlRepo := repository.NewURLRepository(db)

	emailService := email_service.GetSMTPEmailService(a.cfg.EmailConfig)
	otpService := otp_service.NewOTPService(emailService, otpRepo)

	authService := service.NewAuthService(userRepo, sessionRepo, otpService, a.cfg.AccessJWTSecret, a.cfg.RefreshJWTSecret)
	urlService := service.NewURLService(urlRepo)

	authHandler := handler.NewAuthHandler(authService, otpService)
	urlHandler := handler.NewURLHandler(urlService)

	// Router Groups
	a.router.Use(CORSMiddleware())
	a.router.Use(gin.Recovery())
	a.router.Use(logger.LoggerMiddleware(a.cfg.Env, a.cfg.Component))
	routerGroup := a.router.Group("/api/v1")

	// Auth routes
	authRouterGroup := routerGroup.Group("/auth")
	{
		authRouterGroup.POST("/signup", authHandler.SignUp)
		authRouterGroup.POST("/verify-registration-otp", authHandler.VerifyRegistrationOTP)
		authRouterGroup.POST("/login", authHandler.Login)
		authRouterGroup.POST("/refresh-token", authHandler.RefreshToken)
		authRouterGroup.POST("/logout", authHandler.Logout)
		authRouterGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authRouterGroup.POST("/reset-password", authHandler.ResetPassword)
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
		// URL management routes
		protectedURLRouterGroup := protectedRouterGroup.Group("/url")
		{
			protectedURLRouterGroup.POST("", urlHandler.CreateShortURL)
			protectedURLRouterGroup.GET("", urlHandler.GetUserURLs)
		}
	}
}
