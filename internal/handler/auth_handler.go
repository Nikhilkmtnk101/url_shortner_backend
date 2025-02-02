package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/service"
	"github.com/nikhil/url-shortner-backend/internal/service/otp_service"
	"github.com/nikhil/url-shortner-backend/internal/utils"
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
	shouldSecure := os.Getenv("SHOULD_SECURE") == "true"
	shouldHTTPOnly := os.Getenv("SHOULD_HTTP_ONLY") == "true"
	domain := os.Getenv("DOMAIN")
	ctx.SetCookie(name, value, maxAge, "/", domain, shouldSecure, shouldHTTPOnly)
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var signUpReq dto.SignupRequest
	if err := c.ShouldBindJSON(&signUpReq); err != nil {
		utils.NewResponse().SetStatus(http.StatusBadRequest).SetMessage("Invalid request").SetErrorCode("BAD_REQUEST").Build(c)
		return
	}

	user, err := h.authService.SignUp(c, &signUpReq)
	if err != nil {
		utils.NewResponse().SetStatus(http.StatusInternalServerError).SetMessage(err.Error()).SetErrorCode("INTERNAL_ERROR").Build(c)
		return
	}

	utils.NewResponse().SetStatus(http.StatusCreated).SetMessage("User created successfully").SetData(user).Build(c)
}

func (h *AuthHandler) VerifyRegistrationOTP(c *gin.Context) {
	var verifyRegistrationOTPRequest dto.VerifyRegistrationOTPRequest
	if err := c.ShouldBindJSON(&verifyRegistrationOTPRequest); err != nil {
		utils.NewResponse().SetStatus(http.StatusBadRequest).SetMessage("Invalid request").SetErrorCode("BAD_REQUEST").Build(c)
		return
	}

	err := h.authService.RegisterUser(c, verifyRegistrationOTPRequest.Email, verifyRegistrationOTPRequest.OTP)
	if err != nil {
		utils.NewResponse().SetStatus(http.StatusInternalServerError).SetMessage(err.Error()).SetErrorCode("INTERNAL_ERROR").Build(c)
		return
	}
	utils.NewResponse().SetStatus(http.StatusCreated).SetMessage("User created successfully").Build(c)
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var loginRequest dto.LoginRequest
	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		utils.NewResponse().SetStatus(http.StatusBadRequest).SetMessage("Invalid request").SetErrorCode("BAD_REQUEST").Build(ctx)
		return
	}

	token, err := h.authService.Login(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		utils.NewResponse().SetStatus(http.StatusUnauthorized).SetMessage("Unauthorized").SetErrorCode("UNAUTHORIZED").Build(ctx)
		return
	}
	h.setSecureCookie(ctx, "refresh_token", token.RefreshToken, 7*24*60*60, os.Getenv("ENV"))
	utils.NewResponse().SetStatus(http.StatusOK).SetMessage("Login successful").SetData(map[string]string{"access_token": token.AccessToken}).Build(ctx)
}

func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		utils.NewResponse().SetStatus(http.StatusUnauthorized).SetMessage("Missing refresh token").SetErrorCode("UNAUTHORIZED").Build(ctx)
		return
	}
	refreshTokenResponse, err := h.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		utils.NewResponse().SetStatus(http.StatusInternalServerError).SetMessage("Something went wrong").SetErrorCode("INTERNAL_ERROR").Build(ctx)
		return
	}
	utils.NewResponse().SetStatus(http.StatusOK).SetMessage("Token refreshed successfully").SetData(refreshTokenResponse).Build(ctx)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		utils.NewResponse().SetStatus(http.StatusUnauthorized).SetMessage("Missing refresh token").SetErrorCode("UNAUTHORIZED").Build(ctx)
		return
	}
	if err = h.authService.Logout(ctx, refreshToken); err != nil {
		utils.NewResponse().SetStatus(http.StatusInternalServerError).SetMessage("Something went wrong").SetErrorCode("INTERNAL_ERROR").Build(ctx)
		return
	}
	utils.NewResponse().SetStatus(http.StatusOK).SetMessage("Successfully logged out").Build(ctx)
}

func (h *AuthHandler) ForgotPassword(ctx *gin.Context) {
	var forgotPasswordRequest dto.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		utils.NewResponse().SetStatus(http.StatusBadRequest).SetMessage("Invalid request").SetErrorCode("BAD_REQUEST").Build(ctx)
		return
	}
	if err := h.authService.ForgotPassword(ctx, forgotPasswordRequest.Email); err != nil {
		utils.NewResponse().SetStatus(http.StatusInternalServerError).SetMessage("Something went wrong").SetErrorCode("INTERNAL_ERROR").Build(ctx)
		return
	}
	utils.NewResponse().SetStatus(http.StatusOK).SetMessage("OTP successfully sent").Build(ctx)
}

func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var resetPasswordRequest dto.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&resetPasswordRequest); err != nil {
		utils.NewResponse().SetStatus(http.StatusBadRequest).SetMessage("Invalid request").SetErrorCode("BAD_REQUEST").Build(ctx)
		return
	}
	if err := h.authService.ResetPassword(ctx, resetPasswordRequest.Email, resetPasswordRequest.OTP, resetPasswordRequest.NewPassword); err != nil {
		utils.NewResponse().SetStatus(http.StatusUnauthorized).SetMessage("Something went wrong").SetErrorCode("UNAUTHORIZED").Build(ctx)
		return
	}
	utils.NewResponse().SetStatus(http.StatusOK).SetMessage("Password reset successful").Build(ctx)
}
