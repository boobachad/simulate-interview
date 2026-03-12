package handlers

import (
	"net/http"

	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StatsHandler struct {
	statsService services.StatsService
}

func NewStatsHandler(statsService services.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetStats retrieves cached user statistics
func (h *StatsHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	uid := userID.(uuid.UUID)

	stats, err := h.statsService.GetStats(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
