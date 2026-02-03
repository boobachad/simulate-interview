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
		mock_problem, err := services.LoadMockProblem()
		if err != nil {
			log.Printf("Error loading mock problem: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load mock problem",
			})
			return
		}
		// Convert mock problem struct to models.Problem (simplified mapping)
		problem = models.Problem{
			Title:       mock_problem.Title,
			SampleCases: mock_problem.SampleCases,
			HiddenCases: mock_problem.HiddenCases,
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
	exec_service := services.NewExecutionService()

	// Validate code
	if err := exec_service.ValidateCode(request.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Determine test cases based on mode
	var all_cases []models.TestCase

	if request.Mode == "submit" {
		// Submit: Run against hidden cases (and sample cases usually, but user asked specifically for hidden check differentiation)
		// Standard practice: Submit runs EVERYTHING to ensure it passes all constraints.
		all_cases = append(problem.SampleCases, problem.HiddenCases...)
	} else {
		// Run: Run against Sample Cases + Custom Cases
		all_cases = problem.SampleCases
		if len(request.CustomCases) > 0 {
			all_cases = append(all_cases, request.CustomCases...)
		}
	}

	// Execute code
	log.Printf("Executing code for problem: %s", problem.Title)
	results, err := exec_service.Execute(request.Code, all_cases)
	if err != nil {
		log.Printf("Execution error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Count passed tests
	total_passed := 0
	for _, result := range results {
		if result.Passed {
			total_passed++
		}
	}

	response := models.ExecutionResponse{
		Success:     total_passed == len(results),
		Results:     results,
		TotalPassed: total_passed,
		TotalCases:  len(results),
	}

	c.JSON(http.StatusOK, response)
}
