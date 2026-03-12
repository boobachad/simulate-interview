package handlers

import (
	"log"
	"net/http"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerateProblem generates a new problem using LLM
func GenerateProblem(c *gin.Context) {
	var request models.ProblemGenerationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// If no focus areas specified, pick a random one
	if len(request.FocusAreas) == 0 {
		var randomFocusArea models.FocusArea
		// Use RANDOM() for PostgreSQL/SQLite
		result := database.DB.Order("RANDOM()").First(&randomFocusArea)
		if result.Error != nil {
			log.Printf("Error fetching random focus area: %v", result.Error)
			// Fallback if DB fetch fails
			request.FocusAreas = []string{"dynamic-programming"}
		} else {
			request.FocusAreas = []string{randomFocusArea.Slug}
			log.Printf("No focus area specified. Selected random: %s", randomFocusArea.Name)
		}
	}

	// Create LLM provider
	llmProvider, err := services.NewLLMProvider()
	if err != nil {
		log.Printf("Error creating LLM provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize LLM provider",
		})
		return
	}

	// Check if using mock provider
	isMock := services.IsUsingMockProvider(llmProvider)

	// Generate problem with context
	ctx := c.Request.Context()
	log.Printf("Generating problem for focus areas: %v", request.FocusAreas)
	problemResponse, err := llmProvider.GenerateProblem(ctx, request.FocusAreas, "")
	if err != nil {
		log.Printf("Error generating problem: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate problem",
		})
		return
	}

	// If using mock provider, return mock problem with "testing" ID
	if isMock {
		// Find the focus area for the mock
		var focusArea models.FocusArea
		database.DB.Where("slug = ?", problemResponse.FocusArea).First(&focusArea)

		// Return mock problem with special ID "testing"
		mockProblem := models.Problem{
			Title:       problemResponse.Title,
			Description: problemResponse.Description,
			SampleCases: problemResponse.SampleCases,
			HiddenCases: problemResponse.HiddenCases,
			FocusArea:   focusArea,
		}

		// Return with custom id "testing"
		c.JSON(http.StatusOK, gin.H{
			"id":           "testing",
			"title":        mockProblem.Title,
			"description":  mockProblem.Description,
			"focus_area":   mockProblem.FocusArea,
			"sample_cases": mockProblem.SampleCases,
			"hidden_cases": mockProblem.HiddenCases,
			"created_at":   nil,
		})
		return
	}

	// Find or create focus area
	var focusArea models.FocusArea
	result := database.DB.Where("slug = ?", utils.Slugify(problemResponse.FocusArea)).First(&focusArea)
	if result.Error != nil {
		// If focus area doesn't exist, use the first requested focus area
		database.DB.Where("slug = ?", request.FocusAreas[0]).First(&focusArea)
	}

	// Format description markdown
	problemResponse.Description = utils.FormatMarkdownDescription(problemResponse.Description)

	// Check if a problem with the same title already exists (Deduplication)
	var existingProblem models.Problem
	if result := database.DB.Where("title = ?", problemResponse.Title).First(&existingProblem); result.Error == nil {
		log.Printf("Found existing problem with title '%s', reusing ID: %s", existingProblem.Title, existingProblem.ID)
		database.DB.Preload("FocusArea").First(&existingProblem, "id = ?", existingProblem.ID)
		c.JSON(http.StatusOK, existingProblem)
		return
	}

	// Save problem to database
	problem := models.Problem{
		ID:          uuid.New(),
		Title:       problemResponse.Title,
		Description: problemResponse.Description,
		FocusAreaID: focusArea.ID,
		SampleCases: problemResponse.SampleCases,
		HiddenCases: problemResponse.HiddenCases,
	}

	result = database.DB.Create(&problem)
	if result.Error != nil {
		log.Printf("Error saving problem: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save problem",
		})
		return
	}

	// Load the focus area relationship
	database.DB.Preload("FocusArea").First(&problem, "id = ?", problem.ID)

	log.Printf("Problem generated successfully: %s", problem.Title)
	c.JSON(http.StatusOK, problem)
}
