package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Query problems table (no user_id filter - problems are global)
	var problems []models.Problem
	query := database.DB.Order("created_at DESC")

	focusArea := c.Query("focus_area")
	if focusArea != "" {
		query = query.Where("focus_area_topic = ?", focusArea)
	}

	if err := query.Find(&problems).Error; err != nil {
		log.Printf("Error fetching problems: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch problems"})
		return
	}

	// Query session_problems table for ready problems (user-specific)
	var sessionProblems []models.SessionProblem
	sessionQuery := database.DB.
		Joins("JOIN interview_sessions ON interview_sessions.id = session_problems.session_id").
		Where("interview_sessions.user_id = ? AND session_problems.status = ?", uid, "ready").
		Order("session_problems.generated_at DESC")

	if err := sessionQuery.Find(&sessionProblems).Error; err != nil {
		log.Printf("Error fetching session problems: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch session problems"})
		return
	}

	// Batch-fetch FocusArea records to avoid N+1 queries
	focusAreaIDs := make([]uuid.UUID, 0)
	for _, p := range problems {
		if p.FocusAreaID != nil && (p.FocusAreaTopic == nil || *p.FocusAreaTopic == "") {
			focusAreaIDs = append(focusAreaIDs, *p.FocusAreaID)
		}
	}

	focusAreaMap := make(map[uuid.UUID]string)
	if len(focusAreaIDs) > 0 {
		var focusAreas []models.FocusArea
		if err := database.DB.Where("id IN ?", focusAreaIDs).Find(&focusAreas).Error; err == nil {
			for _, fa := range focusAreas {
				focusAreaMap[fa.ID] = fa.Slug
			}
		}
	}

	// Transform problems table results
	response := make([]gin.H, 0, len(problems)+len(sessionProblems))
	for _, p := range problems {
		focusAreaValue := ""
		if p.FocusAreaTopic != nil && *p.FocusAreaTopic != "" {
			focusAreaValue = *p.FocusAreaTopic
		} else if p.FocusAreaID != nil {
			if slug, exists := focusAreaMap[*p.FocusAreaID]; exists {
				focusAreaValue = slug
			}
		}

		response = append(response, gin.H{
			"id":          p.ID,
			"title":       p.Title,
			"description": p.Description,
			"rating":      p.Rating,
			"focus_area":  focusAreaValue,
			"created_at":  p.CreatedAt,
			"source":      "playground",
		})
	}

	// Transform session_problems results
	for _, sp := range sessionProblems {
		if len(sp.ProblemData) == 0 {
			continue
		}

		var problemResponse models.ProblemGenerationResponse
		if err := json.Unmarshal(sp.ProblemData, &problemResponse); err != nil {
			log.Printf("Failed to unmarshal session problem: %v", err)
			continue
		}

		// Apply focus area filter if specified
		if focusArea != "" && problemResponse.FocusArea != focusArea {
			continue
		}

		createdAt := time.Now()
		if sp.GeneratedAt != nil {
			createdAt = *sp.GeneratedAt
		}

		response = append(response, gin.H{
			"id":          sp.ID,
			"title":       problemResponse.Title,
			"description": problemResponse.Description,
			"rating":      problemResponse.Rating,
			"focus_area":  problemResponse.FocusArea,
			"created_at":  createdAt,
			"source":      "session",
		})
	}

	// Sort merged results by created_at DESC (latest first)
	sort.Slice(response, func(i, j int) bool {
		tI := response[i]["created_at"].(time.Time)
		tJ := response[j]["created_at"].(time.Time)
		return tI.After(tJ)
	})

	c.JSON(http.StatusOK, response)
}

// GetProblem returns a single problem by ID
// Special handling for "testing" ID which returns the mock problem
func GetProblem(c *gin.Context) {
	problemID := c.Param("id")

	// Handle mock/testing problem
	if problemID == "testing" {
		mockProblem, err := services.LoadMockProblem()
		if err != nil {
			log.Printf("Error loading mock problem: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load mock problem",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":           "testing",
			"title":        mockProblem.Title,
			"description":  utils.FormatMarkdownDescription(mockProblem.Description),
			"focus_area":   mockProblem.FocusArea,
			"rating":       mockProblem.Rating,
			"sample_cases": mockProblem.SampleCases,
			"hidden_cases": mockProblem.HiddenCases,
			"created_at":   nil,
		})
		return
	}

	var problem models.Problem
	result := database.DB.First(&problem, "id = ?", problemID)

	// If not found in problems table, check session_problems table
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			var sessionProblem models.SessionProblem
			if err := database.DB.First(&sessionProblem, "id = ?", problemID).Error; err == nil {
				// Found in session_problems, unmarshal and return
				var problemResponse models.ProblemGenerationResponse
				if err := json.Unmarshal(sessionProblem.ProblemData, &problemResponse); err != nil {
					log.Printf("Failed to unmarshal session problem: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse problem data"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"id":           sessionProblem.ID,
					"title":        problemResponse.Title,
					"description":  utils.FormatMarkdownDescription(problemResponse.Description),
					"rating":       problemResponse.Rating,
					"focus_area":   problemResponse.FocusArea,
					"sample_cases": problemResponse.SampleCases,
					"hidden_cases": problemResponse.HiddenCases,
					"created_at":   sessionProblem.GeneratedAt,
				})
				return
			}
		}

		log.Printf("Error fetching problem: %v", result.Error)
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}

	// Format description for display
	problem.Description = utils.FormatMarkdownDescription(problem.Description)

	// Determine focus_area value (use topic if available)
	focusAreaValue := ""
	if problem.FocusAreaTopic != nil && *problem.FocusAreaTopic != "" {
		focusAreaValue = *problem.FocusAreaTopic
	} else if problem.FocusAreaID != nil {
		var focusArea models.FocusArea
		if err := database.DB.First(&focusArea, "id = ?", problem.FocusAreaID).Error; err == nil {
			focusAreaValue = focusArea.Slug
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           problem.ID,
		"title":        problem.Title,
		"description":  problem.Description,
		"rating":       problem.Rating,
		"focus_area":   focusAreaValue,
		"sample_cases": problem.SampleCases,
		"hidden_cases": problem.HiddenCases,
		"created_at":   problem.CreatedAt,
	})
}

// GetProblemSession returns the most recent session containing the problem
func GetProblemSession(c *gin.Context) {
	problemID := c.Param("id")

	// Handle mock/testing problem
	if problemID == "testing" {
		c.JSON(http.StatusOK, gin.H{
			"session_id": nil,
		})
		return
	}

	// Validate UUID format
	if !utils.IsValidUUID(problemID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid problem ID format",
		})
		return
	}

	var sessionProblem models.SessionProblem
	result := database.DB.
		Where("problem_id = ?", problemID).
		Order("generated_at DESC").
		First(&sessionProblem)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Problem not in any session
			c.JSON(http.StatusOK, gin.H{
				"session_id": nil,
			})
			return
		}
		// Actual database error
		log.Printf("Error fetching session for problem %s: %v", problemID, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionProblem.SessionID,
	})
}
