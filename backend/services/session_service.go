package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, problemCount int, focusMode, focusTopic string, focusTopics []string) (*SessionData, *models.ProblemGenerationResponse, error) {
	log.Printf("=== SESSION CREATION START ===")
	log.Printf("User ID: %s", userID)
	log.Printf("Problem Count: %d", problemCount)
	log.Printf("Focus Mode: %s", focusMode)
	log.Printf("Focus Topic: %s", focusTopic)
	log.Printf("Focus Topics: %v", focusTopics)

	// Validate focus mode and topics
	if focusMode == "multiple" {
		if len(focusTopics) < 2 || len(focusTopics) > 10 {
			log.Printf("ERROR: Invalid topic count for multiple mode: %d", len(focusTopics))
			return nil, nil, fmt.Errorf("multiple mode requires 2-10 topics, got %d", len(focusTopics))
		}
	}

	var focusTopicPtr *string
	if focusTopic != "" {
		focusTopicPtr = &focusTopic
	}

	log.Printf("Creating session record in database...")
	session := models.InterviewSession{
		UserID:       userID,
		ProblemCount: problemCount,
		FocusMode:    focusMode,
		FocusTopic:   focusTopicPtr,
		FocusTopics:  focusTopics,
		Status:       "active",
	}

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		log.Printf("ERROR: Failed to create session: %v", err)
		return nil, nil, fmt.Errorf("create session: %w", err)
	}
	log.Printf("Session created with ID: %s", session.ID)

	// Build personalization context
	log.Printf("Building personalization context...")
	contextStr, err := s.statsService.BuildPersonalizationContext(ctx, userID, focusMode, focusTopic, focusTopics)
	if err != nil {
		log.Printf("WARNING: Failed to build personalization context: %v", err)
		contextStr = ""
	} else {
		log.Printf("Personalization context built (%d characters)", len(contextStr))
	}

	// Get strategy from config
	strategy := s.generationService.getStrategy()
	log.Printf("Using generation strategy: %s", strategy)

	// Generate first problem synchronously
	log.Printf("Generating first problem (synchronous)...")
	firstProblem, err := s.generationService.GenerateFirstProblem(ctx, session.ID, contextStr, strategy)
	if err != nil {
		log.Printf("ERROR: Failed to generate first problem: %v", err)
		return nil, nil, fmt.Errorf("generate first problem: %w", err)
	}
	log.Printf("First problem generated: %s (rating: %d)", firstProblem.Title, firstProblem.Rating)

	// Store first problem
	log.Printf("Storing first problem in database...")
	problemData, err := json.Marshal(firstProblem)
	if err != nil {
		log.Printf("ERROR: Failed to marshal problem data: %v", err)
		return nil, nil, fmt.Errorf("marshal problem data: %w", err)
	}

	sessionProblem := models.SessionProblem{
		SessionID:     session.ID,
		ProblemNumber: 1,
		Status:        "ready",
		ProblemData:   models.ProblemData(problemData),
	}

	if err := s.db.WithContext(ctx).Create(&sessionProblem).Error; err != nil {
		log.Printf("ERROR: Failed to store first problem: %v", err)
		return nil, nil, fmt.Errorf("store first problem: %w", err)
	}
	log.Printf("First problem stored with ID: %s", sessionProblem.ID)

	// Create placeholder records for remaining problems
	log.Printf("Creating placeholders for %d remaining problems...", problemCount-1)
	for i := 2; i <= problemCount; i++ {
		placeholder := models.SessionProblem{
			SessionID:     session.ID,
			ProblemNumber: i,
			Status:        "generating",
		}
		if err := s.db.WithContext(ctx).Create(&placeholder).Error; err != nil {
			log.Printf("ERROR: Failed to create placeholder for problem %d: %v", i, err)
			return nil, nil, fmt.Errorf("create placeholder: %w", err)
		}
	}
	log.Printf("Placeholders created successfully")

	// Start background queue for remaining problems
	if problemCount > 1 {
		log.Printf("Starting background queue for %d problems...", problemCount-1)
		bgCtx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
		s.generationService.StartBackgroundQueue(bgCtx, session.ID, problemCount, contextStr, strategy)
		log.Printf("Background queue started (will continue independently)")
	}

	sessionData := &SessionData{
		Session:  session,
		Problems: []models.SessionProblem{sessionProblem},
	}

	log.Printf("=== SESSION CREATION COMPLETE ===")
	log.Printf("Session ID: %s", session.ID)
	log.Printf("First Problem ID: %s", sessionProblem.ID)
	log.Printf("Total Problems: %d (1 ready, %d generating)", problemCount, problemCount-1)

	return sessionData, firstProblem, nil
}

// GetSession retrieves session with all problems
func (s *sessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionData, error) {
	log.Printf("=== GET SESSION ===")
	log.Printf("Session ID: %s", sessionID)

	var session models.InterviewSession
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		log.Printf("ERROR: Session not found: %v", err)
		return nil, fmt.Errorf("query session: %w", err)
	}

	log.Printf("Session found - Status: %s, Problems: %d/%d, Current: %d", 
		session.Status, session.CurrentProblemNumber, session.ProblemCount, session.CurrentProblemNumber)

	var problems []models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("problem_number ASC").
		Find(&problems).Error; err != nil {
		log.Printf("ERROR: Failed to query problems: %v", err)
		return nil, fmt.Errorf("query problems: %w", err)
	}

	log.Printf("Found %d problem records:", len(problems))
	for _, p := range problems {
		log.Printf("  Problem #%d: Status=%s, ID=%s", p.ProblemNumber, p.Status, p.ID)
	}

	return &SessionData{
		Session:  session,
		Problems: problems,
	}, nil
}

// GetNextProblem retrieves the next problem if ready
func (s *sessionService) GetNextProblem(ctx context.Context, sessionID uuid.UUID, currentNumber int) (*models.ProblemGenerationResponse, error) {
	nextNumber := currentNumber + 1
	log.Printf("=== GET NEXT PROBLEM ===")
	log.Printf("Session ID: %s, Current: %d, Next: %d", sessionID, currentNumber, nextNumber)

	var sessionProblem models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id = ? AND problem_number = ?", sessionID, nextNumber).
		First(&sessionProblem).Error; err != nil {
		log.Printf("ERROR: Next problem not found: %v", err)
		return nil, fmt.Errorf("query next problem: %w", err)
	}

	log.Printf("Next problem status: %s", sessionProblem.Status)

	if sessionProblem.Status != "ready" {
		log.Printf("WARNING: Problem not ready yet")
		return nil, fmt.Errorf("problem not ready: status=%s", sessionProblem.Status)
	}

	var problemResponse models.ProblemGenerationResponse
	if err := json.Unmarshal(sessionProblem.ProblemData, &problemResponse); err != nil {
		log.Printf("ERROR: Failed to unmarshal problem data: %v", err)
		return nil, fmt.Errorf("unmarshal problem data: %w", err)
	}

	log.Printf("Next problem ready: %s (rating: %d)", problemResponse.Title, problemResponse.Rating)
	return &problemResponse, nil
}

// IsNextProblemReady checks if next problem is ready
func (s *sessionService) IsNextProblemReady(ctx context.Context, sessionID uuid.UUID, currentNumber int) (bool, error) {
	nextNumber := currentNumber + 1
	log.Printf("=== CHECK NEXT PROBLEM READY ===")
	log.Printf("Session ID: %s, Checking problem #%d", sessionID, nextNumber)

	var sessionProblem models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id = ? AND problem_number = ?", sessionID, nextNumber).
		First(&sessionProblem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Next problem record not found")
			return false, nil
		}
		log.Printf("ERROR: Query failed: %v", err)
		return false, fmt.Errorf("query next problem: %w", err)
	}

	ready := sessionProblem.Status == "ready"
	log.Printf("Problem #%d status: %s (ready: %v)", nextNumber, sessionProblem.Status, ready)
	return ready, nil
}

// UpdateCurrentProblem updates the current problem number
func (s *sessionService) UpdateCurrentProblem(ctx context.Context, sessionID uuid.UUID, problemNumber int) error {
	log.Printf("=== UPDATE CURRENT PROBLEM ===")
	log.Printf("Session ID: %s, New current problem: %d", sessionID, problemNumber)

	if err := s.db.WithContext(ctx).
		Model(&models.InterviewSession{}).
		Where("id = ?", sessionID).
		Update("current_problem_number", problemNumber).Error; err != nil {
		log.Printf("ERROR: Failed to update: %v", err)
		return fmt.Errorf("update current problem: %w", err)
	}

	log.Printf("Current problem updated successfully")
	return nil
}

// CompleteSession marks session as completed
func (s *sessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	log.Printf("=== COMPLETE SESSION ===")
	log.Printf("Session ID: %s", sessionID)

	if err := s.db.WithContext(ctx).
		Model(&models.InterviewSession{}).
		Where("id = ?", sessionID).
		Update("status", "completed").Error; err != nil {
		log.Printf("ERROR: Failed to complete session: %v", err)
		return fmt.Errorf("complete session: %w", err)
	}

	log.Printf("Session marked as completed successfully")
	return nil
}

// ActiveSessionSummary represents a summary of an active session
type ActiveSessionSummary struct {
	ID                    string `json:"id"`
	ProblemCount          int    `json:"problem_count"`
	CurrentProblemNumber  int    `json:"current_problem_number"`
	ReadyProblems         int    `json:"ready_problems"`
	FocusMode             string `json:"focus_mode"`
	FirstReadyProblemID   string `json:"first_ready_problem_id,omitempty"`
	CreatedAt             string `json:"created_at"`
}

// ListActiveSessions returns all active sessions for a user
func (s *sessionService) ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]ActiveSessionSummary, error) {
	var sessions []models.InterviewSession
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("query active sessions: %w", err)
	}

	if len(sessions) == 0 {
		return []ActiveSessionSummary{}, nil
	}

	// Batch query: load all session problems in one query
	sessionIDs := make([]uuid.UUID, len(sessions))
	for i, session := range sessions {
		sessionIDs[i] = session.ID
	}

	var allProblems []models.SessionProblem
	if err := s.db.WithContext(ctx).
		Where("session_id IN (?)", sessionIDs).
		Order("session_id ASC, problem_number ASC").
		Find(&allProblems).Error; err != nil {
		log.Printf("Error loading session problems: %v", err)
		return nil, fmt.Errorf("query session problems: %w", err)
	}

	// Group problems by session_id
	problemsBySession := make(map[uuid.UUID][]models.SessionProblem)
	for _, p := range allProblems {
		problemsBySession[p.SessionID] = append(problemsBySession[p.SessionID], p)
	}

	summaries := make([]ActiveSessionSummary, 0, len(sessions))
	for _, session := range sessions {
		problems := problemsBySession[session.ID]

		readyCount := 0
		var firstReadyProblemID string
		for _, p := range problems {
			if p.Status == "ready" {
				readyCount++
				if firstReadyProblemID == "" {
					// Extract problem ID from JSONB data
					var problemResponse models.ProblemGenerationResponse
					if err := json.Unmarshal(p.ProblemData, &problemResponse); err == nil {
						// Use session_problem ID as the problem identifier
						firstReadyProblemID = p.ID.String()
					}
				}
			}
		}

		summaries = append(summaries, ActiveSessionSummary{
			ID:                   session.ID.String(),
			ProblemCount:         session.ProblemCount,
			CurrentProblemNumber: session.CurrentProblemNumber,
			ReadyProblems:        readyCount,
			FocusMode:            session.FocusMode,
			FirstReadyProblemID:  firstReadyProblemID,
			CreatedAt:            session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return summaries, nil
}
