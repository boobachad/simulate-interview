package services

import (
	"context"
	"fmt"
	"time"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authService struct {
	db *gorm.DB
}

// NewAuthService creates a new AuthService instance
func NewAuthService(db *gorm.DB) *authService {
	return &authService{db: db}
}

// Authenticate validates username and password, returns session token
func (s *authService) Authenticate(ctx context.Context, username, password string) (string, error) {
	// Early return for empty credentials
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required")
	}

	// Query user by username with context
	var user models.UserProfile
	if err := s.db.WithContext(ctx).Where("name = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("user %s not found: %w", username, err)
		}
		return "", fmt.Errorf("query user %s: %w", username, err)
	}

	// Validate password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password for user %s: %w", username, err)
	}

	// Create session token
	token, err := s.CreateSession(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("create session for user %s: %w", user.ID, err)
	}

	return token, nil
}

// CreateSession generates UUID token with 24h expiration
func (s *authService) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate UUID token
	token := uuid.New().String()

	// Set expiration to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create session token record
	sessionToken := models.SessionToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := s.db.WithContext(ctx).Create(&sessionToken).Error; err != nil {
		return "", fmt.Errorf("create session token: %w", err)
	}

	return token, nil
}

// ValidateSession checks token expiration and returns associated user
func (s *authService) ValidateSession(ctx context.Context, token string) (*models.UserProfile, error) {
	// Early return for empty token
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Query session token with context
	var sessionToken models.SessionToken
	if err := s.db.WithContext(ctx).Where("token = ?", token).First(&sessionToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session token not found: %w", err)
		}
		return nil, fmt.Errorf("query session token: %w", err)
	}

	// Check expiration
	if time.Now().After(sessionToken.ExpiresAt) {
		return nil, fmt.Errorf("session token expired")
	}

	// Query user profile
	var user models.UserProfile
	if err := s.db.WithContext(ctx).Where("id = ?", sessionToken.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found for session: %w", err)
		}
		return nil, fmt.Errorf("query user for session: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user with bcrypt hashed password
func (s *authService) CreateUser(ctx context.Context, username, password string) (*models.UserProfile, error) {
	// Early return for empty credentials
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user profile
	user := models.UserProfile{
		Name:         username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user %s: %w", username, err)
	}

	return &user, nil
}

// InvalidateSession removes session token (logout)
func (s *authService) InvalidateSession(ctx context.Context, token string) error {
	// Early return for empty token
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Delete session token
	if err := s.db.WithContext(ctx).Where("token = ?", token).Delete(&models.SessionToken{}).Error; err != nil {
		return fmt.Errorf("delete session token: %w", err)
	}

	return nil
}
