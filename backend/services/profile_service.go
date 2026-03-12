package services

import (
	"context"
	"fmt"
	"time"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type profileService struct {
	db *gorm.DB
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(db *gorm.DB) *profileService {
	return &profileService{db: db}
}

// UserProfileData represents complete user profile information
type UserProfileData struct {
	UserID              uuid.UUID  `json:"user_id"`
	LeetCodeUsername    string     `json:"leetcode_username"`
	CodeforcesUsername  string     `json:"codeforces_username"`
	LeetCodeLastSynced  *time.Time `json:"leetcode_last_synced"`
	CodeforcesLastSynced *time.Time `json:"codeforces_last_synced"`
}

// CreateProfile creates coding profiles for LeetCode and Codeforces
func (s *profileService) CreateProfile(ctx context.Context, userID uuid.UUID, leetcodeUsername, codeforcesUsername string) error {
	if leetcodeUsername == "" && codeforcesUsername == "" {
		return fmt.Errorf("at least one platform username is required")
	}

	// Create LeetCode profile if username provided
	if leetcodeUsername != "" {
		leetcodeProfile := models.CodingProfile{
			UserID:   userID,
			Platform: "leetcode",
			Username: leetcodeUsername,
		}
		if err := s.db.WithContext(ctx).Create(&leetcodeProfile).Error; err != nil {
			return fmt.Errorf("create leetcode profile: %w", err)
		}
	}

	// Create Codeforces profile if username provided
	if codeforcesUsername != "" {
		codeforcesProfile := models.CodingProfile{
			UserID:   userID,
			Platform: "codeforces",
			Username: codeforcesUsername,
		}
		if err := s.db.WithContext(ctx).Create(&codeforcesProfile).Error; err != nil {
			return fmt.Errorf("create codeforces profile: %w", err)
		}
	}

	return nil
}

// GetProfile retrieves user profile data from coding_profiles table
func (s *profileService) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfileData, error) {
	var profiles []models.CodingProfile
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("query coding profiles: %w", err)
	}

	profileData := &UserProfileData{
		UserID: userID,
	}

	// Extract platform-specific data
	for _, p := range profiles {
		if p.Platform == "leetcode" {
			profileData.LeetCodeUsername = p.Username
			profileData.LeetCodeLastSynced = p.LastSynced
		} else if p.Platform == "codeforces" {
			profileData.CodeforcesUsername = p.Username
			profileData.CodeforcesLastSynced = p.LastSynced
		}
	}

	return profileData, nil
}

// UpdateProfile updates coding profile usernames
func (s *profileService) UpdateProfile(ctx context.Context, userID uuid.UUID, leetcodeUsername, codeforcesUsername string) error {
	// Update LeetCode username if provided
	if leetcodeUsername != "" {
		if err := s.db.WithContext(ctx).
			Model(&models.CodingProfile{}).
			Where("user_id = ? AND platform = ?", userID, "leetcode").
			Update("username", leetcodeUsername).Error; err != nil {
			return fmt.Errorf("update leetcode username: %w", err)
		}
	}

	// Update Codeforces username if provided
	if codeforcesUsername != "" {
		if err := s.db.WithContext(ctx).
			Model(&models.CodingProfile{}).
			Where("user_id = ? AND platform = ?", userID, "codeforces").
			Update("username", codeforcesUsername).Error; err != nil {
			return fmt.Errorf("update codeforces username: %w", err)
		}
	}

	return nil
}

// HasProfile checks if user has any coding profiles
func (s *profileService) HasProfile(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&models.CodingProfile{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count coding profiles: %w", err)
	}

	return count > 0, nil
}
