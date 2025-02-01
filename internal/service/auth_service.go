package service

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/nikhil/url-shortner-backend/constants"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/middleware/logger"
	"github.com/nikhil/url-shortner-backend/internal/model"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"github.com/nikhil/url-shortner-backend/internal/service/otp_service"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token format")
	ErrNoUserID     = errors.New("user ID not found in token claims")
)

type AuthService struct {
	userRepo          *repository.UserRepository
	tokenRepo         *repository.SessionRepository
	otpService        otp_service.IOTPService
	accessSecret      string
	refreshSecret     string
	maxFailedAttempts int
	lockoutDuration   time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	otpService otp_service.IOTPService,
	accessSecret string,
	refreshSecret string,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		tokenRepo:         sessionRepo,
		otpService:        otpService,
		accessSecret:      accessSecret,
		refreshSecret:     refreshSecret,
		maxFailedAttempts: 5,
		lockoutDuration:   15 * time.Minute,
	}
}

func (s *AuthService) createAccessToken(userId uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

func (s *AuthService) createRefreshToken(userId uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

func (s *AuthService) generateTokenPair(userID uint) (string, string, error) {
	// Generate access token
	accessToken, err := s.createAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.createRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) SignUp(ctx *gin.Context, req *dto.SignupRequest) (*model.User, error) {
	log := logger.GetLogger(ctx)
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}
	// Generate hash with explicit cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("Failed to hash password: %v", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		UserRole: common_constants.UserRoleUser,
	}

	otp := s.otpService.GenerateOTP(req.Email)
	err = s.otpService.SaveOTP(ctx, user.Email, otp)
	if err != nil {
		log.Errorf("Failed to save OTP: %v", err)
		return nil, err
	}
	err = s.userRepo.SaveUserToCache(ctx, user, common_constants.UserSignupCacheTimeout)
	if err != nil {
		log.Errorf("Failed to save user to cache: %v", err)
		return nil, err

	}
	err = s.otpService.SendOTP(user.Email, otp)
	if err != nil {
		log.Errorf("Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) RegisterUser(ctx *gin.Context, email string, otp string) error {
	log := logger.GetLogger(ctx)
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		log.Errorf("User already exists with email: %s", email)
		return errors.New("email already exists")
	}
	user, err := s.userRepo.GetUserFromCache(ctx, email)
	if user == nil {
		log.Errorf("something went wrong please register again")
		return errors.New("something went wrong please register again")
	}
	if err != nil {
		log.Errorf("Failed to get user from cache: %v", err)
		return err
	}
	err = s.otpService.VerifyOTP(ctx, email, otp)
	if err != nil {
		log.Errorf("Failed to verify OTP: %v", err)
		return errors.New("failed to verify OTP")
	}
	err = s.userRepo.Create(user)
	if err != nil {
		log.Errorf("Failed to create user: %v", err)
		return fmt.Errorf("failed to create user")
	}
	err = s.userRepo.DeleteUserFromCache(ctx, email)
	if err != nil {
		log.Errorf("Failed to delete user from cache: %v", err)
	}

	err = s.otpService.DeleteOTP(ctx, email)
	if err != nil {
		log.Errorf("Failed to delete OTP: %v", err)
	}
	return nil
}

func (s *AuthService) Login(ctx *gin.Context, email, password string) (*dto.LoginResponse, error) {
	// Find user
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	// Create new session
	session := &model.Session{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	if err = s.tokenRepo.UpdateUserSession(ctx, user.ID, session); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserIDFromRefreshToken extracts the user ID from a JWT refresh token
// Returns the user ID as uint and error if any occurs
func (s *AuthService) GetUserIDFromRefreshToken(refreshToken string, secretKey []byte) (uint, error) {
	// Parse the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	// Validate token and extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract user ID from claims
		if userID, exists := claims["user_id"]; exists {
			// Handle different numeric types that could come from JSON
			switch v := userID.(type) {
			case float64:
				return uint(v), nil
			case float32:
				return uint(v), nil
			case int:
				return uint(v), nil
			case int64:
				return uint(v), nil
			case uint:
				return v, nil
			case string:
				// If stored as string, try to convert to uint
				parsed, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return 0, ErrInvalidToken
				}
				return uint(parsed), nil
			default:
				return 0, ErrInvalidToken
			}
		}
		return 0, ErrNoUserID
	}

	return 0, ErrInvalidToken
}

func (s *AuthService) RefreshToken(ctx *gin.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	log := logger.GetLogger(ctx)
	userID, err := s.GetUserIDFromRefreshToken(refreshToken, []byte(s.refreshSecret))
	if err != nil {
		log.Errorf("Failed to get user ID: %v", err)
		return nil, err
	}
	err = s.tokenRepo.IsUserSessionIsValid(ctx, userID)
	if err != nil {
		log.Errorf("Failed to get user session: %v", err)
		return nil, err
	}
	accessToken, err := s.createAccessToken(userID)
	if err != nil {
		return nil, err
	}
	return &dto.RefreshTokenResponse{AccessToken: accessToken}, nil
}

func (s *AuthService) Logout(ctx *gin.Context, refreshToken string) error {
	log := logger.GetLogger(ctx)
	userID, err := s.GetUserIDFromRefreshToken(refreshToken, []byte(s.refreshSecret))
	if err != nil {
		log.Errorf("Failed to get user ID: %v", err)
		return err
	}
	return s.tokenRepo.DeleteUserSession(ctx, userID)
}

func (s *AuthService) ForgotPassword(ctx *gin.Context, email string) error {
	log := logger.GetLogger(ctx)
	_, err := s.userRepo.FindByEmail(email)
	if err != nil {
		log.Errorf("failed to get email id %s, err:  %v", email, err)
		return fmt.Errorf("invalid email")
	}
	otp := s.otpService.GenerateOTP(email)
	err = s.otpService.SaveOTP(ctx, email, otp)
	if err != nil {
		log.Errorf("failed to save OTP: %v", err)
		return err
	}
	err = s.otpService.SendOTP(email, otp)
	if err != nil {
		log.Errorf("failed to send OTP: %v", err)
		deleteOTPErr := s.otpService.DeleteOTP(ctx, email)
		if deleteOTPErr != nil {
			log.Errorf("failed to delete OTP: %v", deleteOTPErr)
		}
		return err
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx *gin.Context, email, otp, newPassword string) error {
	log := logger.GetLogger(ctx)
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		log.Errorf("failed to get user id %s, err: %v", email, err)
		return fmt.Errorf("invalid email")
	}
	err = s.otpService.VerifyOTP(ctx, email, otp)
	if err != nil {
		log.Errorf("failed to verify OTP: %v", err)
		return fmt.Errorf("invalid validation failed")
	}
	err = s.otpService.DeleteOTP(ctx, email)
	if err != nil {
		log.Errorf("failed to delete OTP: %v", err)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("failed to hash password: %v", err)
		return fmt.Errorf("invalid password")
	}
	user.Password = string(hashedPassword)
	err = s.tokenRepo.DeleteUserSession(ctx, user.ID)
	if err != nil {
		log.Errorf("failed to delete user session: %v", err)
		return err
	}
	err = s.userRepo.Update(user)
	if err != nil {
		log.Errorf("failed to update user: %v", err)
		return err
	}
	return nil
}

func (s *AuthService) ValidateAccessToken(accessToken string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.accessSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return &claims, nil
}
