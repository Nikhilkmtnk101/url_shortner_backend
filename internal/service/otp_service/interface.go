package otp_service

import "github.com/gin-gonic/gin"

type IOTPService interface {
	GenerateOTP(email string) string
	SendOTP(email string, otp string) error
	SaveOTP(ctx *gin.Context, email string, otp string) error
	VerifyOTP(ctx *gin.Context, email string, otp string) error
	DeleteOTP(ctx *gin.Context, email string) error
}
