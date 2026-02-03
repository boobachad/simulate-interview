package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
)

// GetUserStatsRequest represents the request to fetch user stats
type GetUserStatsRequest struct {
	Name              string `json:"name" binding:"required"`
	LeetCodeUsername  string `json:"leetcode_username"`
	CodeforcesUsername string `json:"codeforces_username"`
}

// GetUserStats fetches and combines stats from LeetCode and Codeforces
func GetUserStats(c *gin.Context) {
	var request GetUserStatsRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if request.LeetCodeUsername == "" && request.CodeforcesUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one platform username must be provided",
		})
		return
	}

	profile := services.UserProfile{
		Name: request.Name,
	}

	// Fetch LeetCode stats
	if request.LeetCodeUsername != "" {
		log.Printf("Fetching LeetCode stats for: %s", request.LeetCodeUsername)
		leetcodeStats, err := services.FetchLeetCodeStats(request.LeetCodeUsername)
		if err != nil {
			log.Printf("Error fetching LeetCode stats: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to fetch LeetCode stats: %v", err),
			})
			return
		}
		profile.LeetCodeStats = leetcodeStats
	}

	// Fetch Codeforces stats
	if request.CodeforcesUsername != "" {
		log.Printf("Fetching Codeforces stats for: %s", request.CodeforcesUsername)
		codeforcesStats, err := services.FetchCodeforcesStats(request.CodeforcesUsername)
		if err != nil {
			log.Printf("Error fetching Codeforces stats: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to fetch Codeforces stats: %v", err),
			})
			return
		}
		profile.CodeforcesStats = codeforcesStats
	}

	// Generate suggested focus areas based on stats
	profile.SuggestedAreas = services.GenerateSuggestedAreas(profile.LeetCodeStats, profile.CodeforcesStats)

	log.Printf("Successfully fetched stats for: %s", request.Name)
	c.JSON(http.StatusOK, profile)
}
