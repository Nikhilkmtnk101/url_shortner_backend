package otp_service

type IOTPService interface {
	GenerateOTP(email string) string
	SendOTP(email string, otp string) error
	VerifyOTP(email string, otp string) error
}
