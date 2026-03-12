package handlers

import (
	"net/http"

	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService    services.AuthService
	profileService services.ProfileService
}

func NewAuthHandler(authService services.AuthService, profileService services.ProfileService) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		profileService: profileService,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	HasProfile bool   `json:"has_profile"`
}

// Login authenticates user and returns session token
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	token, err := h.authService.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Validate session to get user
	user, err := h.authService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate session"})
		return
	}

	// Check if user has profile
	hasProfile, err := h.profileService.HasProfile(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check profile status"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:      token,
		HasProfile: hasProfile,
	})
}

// Logout invalidates session token
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.authService.InvalidateSession(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
