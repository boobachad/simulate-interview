package handlers

import (
	"log"
	"net/http"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
)

// ExecuteCode compiles and executes C++ code against test cases
func ExecuteCode(c *gin.Context) {
	var request models.ExecutionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get problem from database or mock
	var problem models.Problem

	if request.ProblemID == "testing" {
		mockProblem, err := services.LoadMockProblem()
		if err != nil {
			log.Printf("Error loading mock problem: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load mock problem",
			})
			return
		}
		// Convert mock problem struct to models.Problem (simplified mapping)
		problem = models.Problem{
			Title:       mockProblem.Title,
			SampleCases: mockProblem.SampleCases,
			HiddenCases: mockProblem.HiddenCases,
		}
	} else if request.ProblemID == "playground" {
		// Empty problem for playground/scratchpad mode
		// This allows code to run against custom cases only without failing mock problem cases
		problem = models.Problem{
			Title:       "Playground",
			SampleCases: []models.TestCase{},
			HiddenCases: []models.TestCase{},
		}
	} else {
		result := database.DB.First(&problem, "id = ?", request.ProblemID)
		if result.Error != nil {
			log.Printf("Error fetching problem: %v", result.Error)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Problem not found",
			})
			return
		}
	}

	// Create execution service
	execService := services.NewExecutionService()

	// Default to C++ if no language specified
	language := request.Language
	if language == "" {
		language = "cpp"
	}

	// Validate code
	if err := execService.ValidateCode(request.Code, language); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Determine test cases based on mode
	var allCases []models.TestCase

	if request.Mode == "submit" {
		// Submit: Run against hidden cases (and sample cases usually, but user asked specifically for hidden check differentiation)
		// Standard practice: Submit runs EVERYTHING to ensure it passes all constraints.
		allCases = append(problem.SampleCases, problem.HiddenCases...)
	} else {
		// Run: Run against Sample Cases + Custom Cases
		allCases = problem.SampleCases
		if len(request.CustomCases) > 0 {
			allCases = append(allCases, request.CustomCases...)
		}
	}

	// Execute code
	log.Printf("Executing %s code for problem: %s", language, problem.Title)
	results, err := execService.Execute(request.Code, allCases, language)
	if err != nil {
		log.Printf("Execution error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Count passed tests
	totalPassed := 0
	for _, result := range results {
		if result.Passed {
			totalPassed++
		}
	}

	response := models.ExecutionResponse{
		Success:     totalPassed == len(results),
		Results:     results,
		TotalPassed: totalPassed,
		TotalCases:  len(results),
	}

	c.JSON(http.StatusOK, response)
}
