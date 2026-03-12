package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

// GenerateFirstProblem generates the first problem synchronously
func (s *generationService) GenerateFirstProblem(ctx context.Context, sessionID uuid.UUID, contextStr string) (*models.ProblemGenerationResponse, error) {
	var session models.InterviewSession
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	focusAreas := []string{}
	if session.FocusMode == "single" && session.FocusTopic != nil {
		focusAreas = []string{*session.FocusTopic}
	}

	problemResponse, err := s.llmProvider.GenerateProblem(ctx, focusAreas, contextStr)
	if err != nil {
		return nil, fmt.Errorf("generate problem: %w", err)
	}

	return problemResponse, nil
}

// StartBackgroundQueue launches goroutine to generate remaining problems
func (s *generationService) StartBackgroundQueue(ctx context.Context, sessionID uuid.UUID, problemCount int, contextStr string) {
	go func() {
		for problemNumber := 2; problemNumber <= problemCount; problemNumber++ {
			select {
			case <-ctx.Done():
				log.Printf("Background queue cancelled for session %s", sessionID)
				return
			default:
				if err := s.GenerateProblemWithRetry(ctx, sessionID, problemNumber, contextStr); err != nil {
					log.Printf("Failed to generate problem %d for session %s: %v", problemNumber, sessionID, err)
					s.markProblemFailed(ctx, sessionID, problemNumber, err.Error())
				}
			}
		}
		log.Printf("Background queue completed for session %s", sessionID)
	}()
}

// GenerateProblemWithRetry generates a problem with rate limiting and retry logic
func (s *generationService) GenerateProblemWithRetry(ctx context.Context, sessionID uuid.UUID, problemNumber int, contextStr string) error {
	var session models.InterviewSession
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("query session: %w", err)
	}

	focusAreas := []string{}
	if session.FocusMode == "single" && session.FocusTopic != nil {
		focusAreas = []string{*session.FocusTopic}
	}

	var problemResponse *models.ProblemGenerationResponse
	var generateErr error

	err := s.rateLimiter.ExecuteWithBackoff(ctx, func() error {
		problemResponse, generateErr = s.llmProvider.GenerateProblem(ctx, focusAreas, contextStr)
		return generateErr
	})

	if err != nil {
		return fmt.Errorf("generate with backoff: %w", err)
	}

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
