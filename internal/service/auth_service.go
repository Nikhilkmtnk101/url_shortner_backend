package service

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nikhil/url-shortner-backend/constants"
	"github.com/nikhil/url-shortner-backend/internal/dto"
	"github.com/nikhil/url-shortner-backend/internal/models"
	"github.com/nikhil/url-shortner-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type AuthService struct {
	userRepo          *repository.UserRepository
	tokenRepo         *repository.SessionRepository
	loginAttemptRepo  *repository.LoginAttemptRepository
	accessSecret      string
	refreshSecret     string
	maxFailedAttempts int
	lockoutDuration   time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	loginAttemptRepo *repository.LoginAttemptRepository,
	accessSecret string,
	refreshSecret string,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		tokenRepo:         sessionRepo,
		loginAttemptRepo:  loginAttemptRepo,
		accessSecret:      accessSecret,
		refreshSecret:     refreshSecret,
		maxFailedAttempts: 5,
		lockoutDuration:   15 * time.Minute,
	}
}

func (s *AuthService) SignUp(req *dto.SignupRequest) (*models.User, error) {
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}
	// Generate hash with explicit cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		UserRole: common_constants.UserRoleUser,
	}

	// Attempt to create user
	if err := s.userRepo.Create(user); err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(email, password, userAgent, ip string) (*dto.TokenResponse, error) {
	// Check if IP is blocked due to too many failed attempts
	if s.isIPBlocked(ip) {
		return nil, errors.New("too many failed attempts, please try again later")
	}

	// Find user
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.recordLoginAttempt(0, ip, userAgent, "failure", "user not found")
		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if s.isAccountLocked(user.ID) {
		return nil, errors.New("account is temporarily locked due to too many failed attempts")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.recordLoginAttempt(user.ID, ip, userAgent, "failure", "invalid password")
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Create new session
	session := &models.Session{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IP:           ip,
		UserAgent:    userAgent,
		LastUsedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		IsActive:     true,
	}

	if err := s.tokenRepo.CreateSession(session); err != nil {
		return nil, err
	}

	// Record successful login
	s.recordLoginAttempt(user.ID, ip, userAgent, "success", "")

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *AuthService) generateTokenPair(user *models.User) (string, string, error) {
	// Generate access token
	accessToken, err := s.createAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.createRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) createAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

func (s *AuthService) createRefreshToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

func (s *AuthService) isIPBlocked(ip string) bool {
	attempts, err := s.loginAttemptRepo.GetRecentFailedAttempts(ip, time.Now().Add(-s.lockoutDuration))
	if err != nil {
		return false
	}
	return len(attempts) >= s.maxFailedAttempts
}

func (s *AuthService) isAccountLocked(userID uint) bool {
	attempts, err := s.loginAttemptRepo.GetUserRecentFailedAttempts(userID, time.Now().Add(-s.lockoutDuration))
	if err != nil {
		return false
	}
	return len(attempts) >= s.maxFailedAttempts
}

func (s *AuthService) recordLoginAttempt(userID uint, ip, userAgent, status, reason string) {
	attempt := &models.LoginAttempt{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		Status:    status,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
	s.loginAttemptRepo.Create(attempt)
}

func (s *AuthService) GetActiveSessions(userID uint) ([]models.Session, error) {
	return s.tokenRepo.GetActiveSessions(userID)
}

func (s *AuthService) RevokeSession(sessionID uint) error {
	return s.tokenRepo.DeactivateSession(sessionID)
}

func (s *AuthService) RevokeAllSessions(userID uint, exceptSessionID uint) error {
	return s.tokenRepo.DeactivateAllUserSessions(userID, exceptSessionID)
}

func (s *AuthService) RefreshToken(refreshToken, userAgent, ip string) (*dto.TokenResponse, error) {
	// Verify refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.refreshSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Get user ID from claims
	userID := uint(claims["user_id"].(float64))

	// Find session
	session, err := s.tokenRepo.FindSessionByRefreshToken(refreshToken)
	if err != nil || !session.IsActive {
		return nil, errors.New("invalid session")
	}

	// Verify session belongs to the same client
	if session.UserAgent != userAgent || session.IP != ip {
		s.recordLoginAttempt(userID, ip, userAgent, "failure", "session mismatch")
		return nil, errors.New("invalid session")
	}

	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new token pair
	accessToken, newRefreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Update session
	session.AccessToken = accessToken
	session.RefreshToken = newRefreshToken
	session.LastUsedAt = time.Now()
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)

	if err := s.tokenRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
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
