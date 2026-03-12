package handlers

import (
	"log"
	"net/http"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/gin-gonic/gin"
)

// GetFocusAreas returns all available focus areas
func GetFocusAreas(c *gin.Context) {
	var focusAreas []models.FocusArea

	result := database.DB.Order("name ASC").Find(&focusAreas)
	if result.Error != nil {
		log.Printf("Error fetching focus areas: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch focus areas",
		})
		return
	}

	c.JSON(http.StatusOK, focusAreas)
}

// GetProblems returns problems, optionally filtered by focus area
func GetProblems(c *gin.Context) {
	var problems []models.Problem

	query := database.DB.Preload("FocusArea").Order("created_at DESC")

	// Optional filter by focus area slug
	focusArea := c.Query("focus_area")
	if focusArea != "" {
		query = query.Joins("JOIN focus_areas ON focus_areas.id = problems.focus_area_id").
			Where("focus_areas.slug = ?", focusArea)
	}

	result := query.Find(&problems)
	if result.Error != nil {
		log.Printf("Error fetching problems: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch problems",
		})
		return
	}

	c.JSON(http.StatusOK, problems)
}

// GetProblem returns a single problem by ID
// Special handling for "testing" ID which returns the mock problem
func GetProblem(c *gin.Context) {
	problemID := c.Param("id")

	// Handle mock/testing problem
	if problemID == "testing" {
		// Return the mock problem from file
		mockProblem, err := services.LoadMockProblem()
		if err != nil {
			log.Printf("Error loading mock problem: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load mock problem",
			})
			return
		}

		// Find the focus area
		var focusArea models.FocusArea
		database.DB.Where("slug = ?", mockProblem.FocusArea).First(&focusArea)

		// Return mock problem with special handling
		c.JSON(http.StatusOK, gin.H{
			"id":           "testing",
			"title":        mockProblem.Title,
			"description":  utils.FormatMarkdownDescription(mockProblem.Description),
			"focus_area":   focusArea,
			"sample_cases": mockProblem.SampleCases,
			"hidden_cases": mockProblem.HiddenCases,
			"created_at":   nil,
		})
		return
	}

	var problem models.Problem
	result := database.DB.Preload("FocusArea").First(&problem, "id = ?", problemID)
	if result.Error != nil {
		log.Printf("Error fetching problem: %v", result.Error)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Problem not found",
		})
		return
	}

	// Format description for display
	problem.Description = utils.FormatMarkdownDescription(problem.Description)

	c.JSON(http.StatusOK, problem)
}
