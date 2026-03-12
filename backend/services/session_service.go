package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionService struct {
	db                *gorm.DB
	generationService *generationService
	statsService      *statsService
}

// NewSessionService creates a new SessionService instance
func NewSessionService(db *gorm.DB, generationService *generationService, statsService *statsService) *sessionService {
	return &sessionService{
		db:                db,
		generationService: generationService,
		statsService:      statsService,
	}
}

// SessionData represents complete session information
type SessionData struct {
	Session  models.InterviewSession `json:"session"`
	Problems []models.SessionProblem `json:"problems"`
}

// CreateSession creates a new interview session and generates first problem
func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, problemCount int, focusMode, focusTopic string) (*SessionData, *models.ProblemGenerationResponse, error) {
	var focusTopicPtr *string
	if focusTopic != "" {
		focusTopicPtr = &focusTopic
	}

	session := models.InterviewSession{
		UserID:       userID,
		ProblemCount: problemCount,
		FocusMode:    focusMode,
		FocusTopic:   focusTopicPtr,
		Status:       "active",
	}

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	// Build personalization context
	contextStr, err := s.statsService.BuildPersonalizationContext(ctx, userID, focusMode, focusTopic)
	if err != nil {
		contextStr = ""
	}

	// Generate first problem synchronously
	firstProblem, err := s.generationService.GenerateFirstProblem(ctx, session.ID, contextStr)
	if err != nil {
		return nil, nil, fmt.Errorf("generate first problem: %w", err)
	}

	// Store first problem
	problemData, err := json.Marshal(firstProblem)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal problem data: %w", err)
	}

	sessionProblem := models.SessionProblem{
		SessionID:     session.ID,
		ProblemNumber: 1,
		Status:        "ready",
		ProblemData:   models.ProblemData(problemData),
	}

	if err := s.db.WithContext(ctx).Create(&sessionProblem).Error; err != nil {
		return nil, nil, fmt.Errorf("store first problem: %w", err)
	}

	// Create placeholder records for remaining problems
	for i := 2; i <= problemCount; i++ {
		placeholder := models.SessionProblem{
			SessionID:     session.ID,
			ProblemNumber: i,
			Status:        "generating",
		}
		if err := s.db.WithContext(ctx).Create(&placeholder).Error; err != nil {
			return nil, nil, fmt.Errorf("create placeholder: %w", err)
		}
	}

	// Start background queue for remaining problems
	if problemCount > 1 {
		s.generationService.StartBackgroundQueue(ctx, session.ID, problemCount, contextStr)
	}

	sessionData := &SessionData{
		Session:  session,
		Problems: []models.SessionProblem{sessionProblem},
	}

	return sessionData, firstProblem, nil
}

// GetSession retrieves session with all problems
func (s *sessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionData, error) {
	var session models.InterviewSession
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	var problems []models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("problem_number ASC").
		Find(&problems).Error; err != nil {
		return nil, fmt.Errorf("query problems: %w", err)
	}

	return &SessionData{
		Session:  session,
		Problems: problems,
	}, nil
}

// GetNextProblem retrieves the next problem if ready
func (s *sessionService) GetNextProblem(ctx context.Context, sessionID uuid.UUID, currentNumber int) (*models.ProblemGenerationResponse, error) {
	nextNumber := currentNumber + 1

	var sessionProblem models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id = ? AND problem_number = ?", sessionID, nextNumber).
		First(&sessionProblem).Error; err != nil {
		return nil, fmt.Errorf("query next problem: %w", err)
	}

	if sessionProblem.Status != "ready" {
		return nil, fmt.Errorf("problem not ready: status=%s", sessionProblem.Status)
	}

	var problemResponse models.ProblemGenerationResponse
	if err := json.Unmarshal(sessionProblem.ProblemData, &problemResponse); err != nil {
		return nil, fmt.Errorf("unmarshal problem data: %w", err)
	}

	return &problemResponse, nil
}

// IsNextProblemReady checks if next problem is ready
func (s *sessionService) IsNextProblemReady(ctx context.Context, sessionID uuid.UUID, currentNumber int) (bool, error) {
	nextNumber := currentNumber + 1

	var sessionProblem models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id = ? AND problem_number = ?", sessionID, nextNumber).
		First(&sessionProblem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("query next problem: %w", err)
	}

	return sessionProblem.Status == "ready", nil
}

// UpdateCurrentProblem updates the current problem number
func (s *sessionService) UpdateCurrentProblem(ctx context.Context, sessionID uuid.UUID, problemNumber int) error {
	if err := s.db.WithContext(ctx).
		Model(&models.InterviewSession{}).
		Where("id = ?", sessionID).
		Update("current_problem_number", problemNumber).Error; err != nil {
		return fmt.Errorf("update current problem: %w", err)
	}

	return nil
}

// CompleteSession marks session as completed
func (s *sessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.db.WithContext(ctx).
		Model(&models.InterviewSession{}).
		Where("id = ?", sessionID).
		Update("status", "completed").Error; err != nil {
		return fmt.Errorf("complete session: %w", err)
	}

	return nil
}
