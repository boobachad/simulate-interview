package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boobachad/simulate-interview/backend/config"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type generationService struct {
	db           *gorm.DB
	llmProvider  LLMProvider
	statsService *statsService
	rateLimiter  *utils.RateLimiter
}

// NewGenerationService creates a new GenerationService instance
func NewGenerationService(db *gorm.DB, llmProvider LLMProvider, statsService *statsService, rateLimiter *utils.RateLimiter) *generationService {
	return &generationService{
		db:           db,
		llmProvider:  llmProvider,
		statsService: statsService,
		rateLimiter:  rateLimiter,
	}
}

// getStrategy returns the configured problem generation strategy
func (s *generationService) getStrategy() string {
	return config.Config.ProblemGenerationStrategy
}

// normalizeRating clamps ratings to [800,3000] and logs invalid values
func normalizeRating(rating int, context string) int {
	if rating < 800 || rating > 3000 {
		log.Printf("Invalid rating %d for %s, defaulting to 1200", rating, context)
		return 1200
	}
	return rating
}

// GenerateFirstProblem generates the first problem synchronously
func (s *generationService) GenerateFirstProblem(ctx context.Context, sessionID uuid.UUID, contextStr string, strategy string) (*models.ProblemGenerationResponse, error) {
	var session models.InterviewSession
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	focusAreas := s.selectFocusAreas(&session, 1, strategy)

	problemResponse, err := s.llmProvider.GenerateProblem(ctx, focusAreas, contextStr, nil)
	if err != nil {
		return nil, fmt.Errorf("generate problem: %w", err)
	}

	problemResponse.Rating = normalizeRating(problemResponse.Rating, "first problem")
	return problemResponse, nil
}

// StartBackgroundQueue launches goroutine to generate remaining problems
func (s *generationService) StartBackgroundQueue(ctx context.Context, sessionID uuid.UUID, problemCount int, contextStr string, strategy string) {
	log.Printf("=== BACKGROUND QUEUE START ===")
	log.Printf("Session ID: %s", sessionID)
	log.Printf("Total problems to generate: %d (problems 2-%d)", problemCount-1, problemCount)
	log.Printf("Context deadline: %v", ctx.Err())
	
	go func() {
		log.Printf("GOROUTINE STARTED for session %s", sessionID)
		for problemNumber := 2; problemNumber <= problemCount; problemNumber++ {
			select {
			case <-ctx.Done():
				log.Printf("ERROR: Background queue cancelled for session %s (reason: %v)", sessionID, ctx.Err())
				return
			default:
				log.Printf("=== GENERATING PROBLEM #%d ===", problemNumber)
				log.Printf("Session: %s", sessionID)
				
				if err := s.GenerateProblemWithRetry(ctx, sessionID, problemNumber, contextStr, strategy); err != nil {
					log.Printf("ERROR: Failed to generate problem %d for session %s: %v", problemNumber, sessionID, err)
					s.markProblemFailed(ctx, sessionID, problemNumber, err.Error())
				} else {
					log.Printf("SUCCESS: Problem #%d generated and stored", problemNumber)
				}
			}
		}
		log.Printf("=== BACKGROUND QUEUE COMPLETE ===")
		log.Printf("Session %s: All %d problems generated", sessionID, problemCount-1)
	}()
	
	log.Printf("Background goroutine launched for session %s", sessionID)
}

// GenerateProblemWithRetry generates a problem with rate limiting and retry logic
func (s *generationService) GenerateProblemWithRetry(ctx context.Context, sessionID uuid.UUID, problemNumber int, contextStr string, strategy string) error {
	var session models.InterviewSession
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("query session: %w", err)
	}

	focusAreas := s.selectFocusAreas(&session, problemNumber, strategy)

	var problemResponse *models.ProblemGenerationResponse
	var generateErr error

	err := s.rateLimiter.ExecuteWithBackoff(ctx, func() error {
		problemResponse, generateErr = s.llmProvider.GenerateProblem(ctx, focusAreas, contextStr, nil)
		return generateErr
	})

	if err != nil {
		return fmt.Errorf("generate with backoff: %w", err)
	}

	problemResponse.Rating = normalizeRating(problemResponse.Rating, fmt.Sprintf("retry problem %d", problemNumber))

	problemData, err := json.Marshal(problemResponse)
	if err != nil {
		return fmt.Errorf("marshal problem data: %w", err)
	}

	now := time.Now()
	sessionProblem := models.SessionProblem{
		SessionID:     sessionID,
		ProblemNumber: problemNumber,
		Status:        "ready",
		ProblemData:   models.ProblemData(problemData),
		GeneratedAt:   &now,
	}

	if err := s.db.WithContext(ctx).
		Where("session_id = ? AND problem_number = ?", sessionID, problemNumber).
		Assign(sessionProblem).
		FirstOrCreate(&sessionProblem).Error; err != nil {
		return fmt.Errorf("store session problem: %w", err)
	}

	log.Printf("Generated problem %d for session %s", problemNumber, sessionID)
	return nil
}

// selectFocusAreas determines which topics to use based on strategy
func (s *generationService) selectFocusAreas(session *models.InterviewSession, problemNumber int, strategy string) []string {
	var allTopics []string

	if session.FocusMode == "single" && session.FocusTopic != nil {
		allTopics = []string{*session.FocusTopic}
	} else if session.FocusMode == "multiple" && len(session.FocusTopics) > 0 {
		allTopics = session.FocusTopics
	} else {
		return []string{}
	}

	if len(allTopics) == 0 {
		return []string{}
	}

	if len(allTopics) == 1 {
		return allTopics
	}

	switch strategy {
	case "rotate":
		index := (problemNumber - 1) % len(allTopics)
		return []string{allTopics[index]}
	case "combine":
		return allTopics
	case "mix":
		if problemNumber%2 == 0 {
			return allTopics
		}
		index := (problemNumber - 1) % len(allTopics)
		return []string{allTopics[index]}
	default:
		return allTopics
	}
}

// markProblemFailed updates session_problems status to failed
func (s *generationService) markProblemFailed(ctx context.Context, sessionID uuid.UUID, problemNumber int, errorMessage string) {
	if err := s.db.WithContext(ctx).
		Model(&models.SessionProblem{}).
		Where("session_id = ? AND problem_number = ?", sessionID, problemNumber).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errorMessage,
		}).Error; err != nil {
		log.Printf("Failed to mark problem as failed: %v", err)
	}
}
