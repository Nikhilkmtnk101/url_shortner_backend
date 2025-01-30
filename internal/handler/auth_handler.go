package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"github.com/nikhil/url-shortner-backend/internal/service/otp_service"
	"net/http"
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

func (h *AuthHandler) SignUp(c *gin.Context) {
	var signUpReq dto.SignupRequest
	if err := c.ShouldBindJSON(&signUpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user *model.User
	var err error
	if user, err = h.authService.SignUp(&signUpReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(credentials.Email, credentials.Password, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) SendOTP(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err := h.otpService.SendOTP(request.Email, "576")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "otp successfully sent"})
}
