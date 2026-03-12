package handlers

import (
	"net/http"

	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	profileService services.ProfileService
	statsService   services.StatsService
}

func NewProfileHandler(profileService services.ProfileService, statsService services.StatsService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		statsService:   statsService,
	}
}

type ProfileSetupRequest struct {
	LeetCodeUsername   string `json:"leetcode_username"`
	CodeforcesUsername string `json:"codeforces_username"`
}

type ProfileUpdateRequest struct {
	LeetCodeUsername   string `json:"leetcode_username"`
	CodeforcesUsername string `json:"codeforces_username"`
}

// Setup creates user profile and triggers stats sync
func (h *ProfileHandler) Setup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req ProfileSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.LeetCodeUsername == "" && req.CodeforcesUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one platform username is required"})
		return
	}

	uid := userID.(uuid.UUID)

	// Create profile
	if err := h.profileService.CreateProfile(c.Request.Context(), uid, req.LeetCodeUsername, req.CodeforcesUsername); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	// Trigger stats sync
	syncStatus := "success"
	if err := h.statsService.SyncStats(c.Request.Context(), uid); err != nil {
		syncStatus = "partial"
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"sync_status": syncStatus,
	})
}

// GetProfile retrieves user profile data
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	uid := userID.(uuid.UUID)

	profile, err := h.profileService.GetProfile(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates user profile usernames
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	uid := userID.(uuid.UUID)

	if err := h.profileService.UpdateProfile(c.Request.Context(), uid, req.LeetCodeUsername, req.CodeforcesUsername); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SyncStats manually triggers stats synchronization
func (h *ProfileHandler) SyncStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	uid := userID.(uuid.UUID)

	if err := h.statsService.SyncStats(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to fetch stats from both platforms, please try again later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"synced_at": "now",
	})
}
