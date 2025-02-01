package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"github.com/nikhil/url-shortner-backend/internal/service/otp_service"
	"net/http"
	"os"
)

type AuthHandler struct {
	authService *service.AuthService
	otpService  otp_service.IOTPService
}

func NewAuthHandler(authService *service.AuthService, otpService otp_service.IOTPService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		otpService:  otpService,
	}
}

func (h *AuthHandler) setSecureCookie(ctx *gin.Context, name, value string, maxAge int, env string) {
	// Determine secure and domain settings based on environment
	var secure, httpOnly bool
	var domain string

	switch env {
	case "prod":
		secure = true
		httpOnly = true
		domain = "yourproductiondomain.com" // Replace with your actual domain
	case "stage", "uat":
		secure = true
		httpOnly = true
		domain = "staging.yourdomain.com" // Replace with actual staging domain
	default: // "local" or others
		secure = false
		httpOnly = true
		domain = ""
	}

	ctx.SetCookie(
		name,     // Cookie name
		value,    // Cookie value
		maxAge,   // Max age in seconds
		"/",      // Path
		domain,   // Domain
		secure,   // Secure flag
		httpOnly, // HTTPOnly flag
	)
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var signUpReq dto.SignupRequest
	if err := c.ShouldBindJSON(&signUpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user *model.User
	var err error
	if user, err = h.authService.SignUp(c, &signUpReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func (h *AuthHandler) VerifyRegistrationOTP(c *gin.Context) {
	var verifyRegistrationOTPRequest dto.VerifyRegistrationOTPRequest
	if err := c.ShouldBindJSON(&verifyRegistrationOTPRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.RegisterUser(c, verifyRegistrationOTPRequest.Email, verifyRegistrationOTPRequest.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var loginRequest dto.LoginRequest
	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	h.setSecureCookie(ctx, "refresh_token", token.RefreshToken, 7*24*60*60, os.Getenv("ENV"))
	ctx.JSON(http.StatusOK, gin.H{"data": map[string]string{"access_token": token.AccessToken}})
}

func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	// Read refresh token from cookie
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}
	refreshTokenResponse, err := h.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": *refreshTokenResponse})
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	// Read refresh token from cookie
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}
	err = h.authService.Logout(ctx, refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Something went wrong"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (h *AuthHandler) ForgotPassword(ctx *gin.Context) {
	var forgotPasswordRequest dto.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	err := h.authService.ForgotPassword(ctx, forgotPasswordRequest.Email)
	if err != nil {
		ctx.JSON(http.StatusInsufficientStorage, gin.H{"error": "Something went wrong"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "otp successfully sent"})
}

func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var resetPasswordRequest dto.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&resetPasswordRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.authService.ResetPassword(
		ctx,
		resetPasswordRequest.Email,
		resetPasswordRequest.OTP,
		resetPasswordRequest.NewPassword,
	)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Something went wrong"})
		return
	}
}
