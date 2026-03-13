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

type GenerationHandler struct {
	statsService services.StatsService
}

func NewGenerationHandler(statsService services.StatsService) *GenerationHandler {
	return &GenerationHandler{
		statsService: statsService,
	}
}

// GenerateProblem generates a new problem using LLM
func (h *GenerationHandler) GenerateProblem(c *gin.Context) {
	var request models.ProblemGenerationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	var userUUID uuid.UUID
	if exists {
		var ok bool
		userUUID, ok = userID.(uuid.UUID)
		if !ok {
			log.Printf("Warning: Invalid user_id type in context, generating without personalization")
			exists = false
		}
	} else {
		// If no user context, generate without personalization
		log.Printf("Warning: No user context found, generating without personalization")
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

	// Build personalization context
	ctx := c.Request.Context()
	personalizationContext := ""

	if exists {
		// Determine focus mode based on number of focus areas
		focusMode := "all"
		focusTopic := ""
		focusTopics := []string{}

		if len(request.FocusAreas) == 1 {
			focusMode = "single"
			focusTopic = request.FocusAreas[0]
		} else if len(request.FocusAreas) > 1 {
			focusMode = "multiple"
			focusTopics = request.FocusAreas
		}

		var err error
		personalizationContext, err = h.statsService.BuildPersonalizationContext(ctx, userUUID, focusMode, focusTopic, focusTopics)
		if err != nil {
			log.Printf("Failed to build personalization context: %v", err)
			personalizationContext = ""
		} else {
			log.Printf("=== PERSONALIZATION CONTEXT ===")
			log.Printf("Length: %d characters", len(personalizationContext))
			log.Printf("Content:\n%s", personalizationContext)
			log.Printf("=== END PERSONALIZATION CONTEXT ===")
		}
	}

	// Generate problem with context
	log.Printf("Generating problem for focus areas: %v", request.FocusAreas)
	if request.TargetRating != nil {
		log.Printf("Target rating: %d", *request.TargetRating)
	}
	problemResponse, err := llmProvider.GenerateProblem(ctx, request.FocusAreas, personalizationContext, request.TargetRating)
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
			Rating:      problemResponse.Rating,
			SampleCases: problemResponse.SampleCases,
			HiddenCases: problemResponse.HiddenCases,
			FocusArea:   focusArea,
		}

		// Return with custom id "testing"
		c.JSON(http.StatusOK, gin.H{
			"id":           "testing",
			"title":        mockProblem.Title,
			"description":  mockProblem.Description,
			"rating":       mockProblem.Rating,
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

		focusAreaValue := ""
		if existingProblem.FocusAreaTopic != nil && *existingProblem.FocusAreaTopic != "" {
			focusAreaValue = *existingProblem.FocusAreaTopic
		} else if existingProblem.FocusAreaID != nil {
			var fa models.FocusArea
			if err := database.DB.First(&fa, "id = ?", existingProblem.FocusAreaID).Error; err == nil {
				focusAreaValue = fa.Slug
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"id":           existingProblem.ID,
			"title":        existingProblem.Title,
			"description":  existingProblem.Description,
			"rating":       existingProblem.Rating,
			"focus_area":   focusAreaValue,
			"sample_cases": existingProblem.SampleCases,
			"created_at":   existingProblem.CreatedAt,
		})
		return
	}

	// Save problem to database with focus_area_topic
	focusAreaTopic := problemResponse.FocusArea
	problem := models.Problem{
		ID:             uuid.New(),
		Title:          problemResponse.Title,
		Description:    problemResponse.Description,
		Rating:         problemResponse.Rating,
		FocusAreaTopic: &focusAreaTopic,
		SampleCases:    problemResponse.SampleCases,
		HiddenCases:    problemResponse.HiddenCases,
	}

	result = database.DB.Create(&problem)
	if result.Error != nil {
		log.Printf("Error saving problem: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save problem",
		})
		return
	}

	log.Printf("Problem generated successfully: %s", problem.Title)
	c.JSON(http.StatusOK, gin.H{
		"id":           problem.ID,
		"title":        problem.Title,
		"description":  problem.Description,
		"rating":       problem.Rating,
		"focus_area":   focusAreaTopic,
		"sample_cases": problem.SampleCases,
		"created_at":   problem.CreatedAt,
	})
}
