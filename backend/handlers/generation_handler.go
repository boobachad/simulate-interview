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
	llm_provider, err := services.NewLLMProvider()
	if err != nil {
		log.Printf("Error creating LLM provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize LLM provider",
		})
		return
	}

	// Check if using mock provider
	is_mock := services.IsUsingMockProvider(llm_provider)

	// Generate problem
	log.Printf("Generating problem for focus areas: %v", request.FocusAreas)
	problem_response, err := llm_provider.GenerateProblem(request.FocusAreas)
	if err != nil {
		log.Printf("Error generating problem: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate problem",
		})
		return
	}

	// If using mock provider, return mock problem with "testing" ID
	if is_mock {
		// Find the focus area for the mock
		var focus_area models.FocusArea
		database.DB.Where("slug = ?", problem_response.FocusArea).First(&focus_area)

		// Return mock problem with special ID "testing"
		mock_problem := models.Problem{
			Title:       problem_response.Title,
			Description: problem_response.Description,
			SampleCases: problem_response.SampleCases,
			HiddenCases: problem_response.HiddenCases,
			FocusArea:   focus_area,
		}

		// Return with custom id "testing"
		c.JSON(http.StatusOK, gin.H{
			"id":           "testing",
			"title":        mock_problem.Title,
			"description":  mock_problem.Description,
			"focus_area":   mock_problem.FocusArea,
			"sample_cases": mock_problem.SampleCases,
			"hidden_cases": mock_problem.HiddenCases,
			"created_at":   nil,
		})
		return
	}

	// Find or create focus area
	var focus_area models.FocusArea
	result := database.DB.Where("slug = ?", utils.Slugify(problem_response.FocusArea)).First(&focus_area)
	if result.Error != nil {
		// If focus area doesn't exist, use the first requested focus area
		database.DB.Where("slug = ?", request.FocusAreas[0]).First(&focus_area)
	}

	// Format description markdown
	problem_response.Description = utils.FormatMarkdownDescription(problem_response.Description)

	// Check if a problem with the same title already exists (Deduplication)
	var existingProblem models.Problem
	if result := database.DB.Where("title = ?", problem_response.Title).First(&existingProblem); result.Error == nil {
		log.Printf("Found existing problem with title '%s', reusing ID: %s", existingProblem.Title, existingProblem.ID)
		database.DB.Preload("FocusArea").First(&existingProblem, "id = ?", existingProblem.ID)
		c.JSON(http.StatusOK, existingProblem)
		return
	}

	// Save problem to database
	problem := models.Problem{
		ID:          uuid.New(),
		Title:       problem_response.Title,
		Description: problem_response.Description,
		FocusAreaID: focus_area.ID,
		SampleCases: problem_response.SampleCases,
		HiddenCases: problem_response.HiddenCases,
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
