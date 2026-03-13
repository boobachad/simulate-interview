package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SessionHandler struct {
	sessionService services.SessionService
}

func NewSessionHandler(sessionService services.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

type CreateSessionRequest struct {
	FocusMode    string   `json:"focus_mode" binding:"required"`
	FocusTopic   string   `json:"focus_topic"`
	FocusTopics  []string `json:"focus_topics"`
	ProblemCount int      `json:"problem_count" binding:"required,min=1,max=10"`
}

type CreateSessionResponse struct {
	SessionID    string                 `json:"session_id"`
	FirstProblem map[string]interface{} `json:"first_problem"`
}

type NextProblemResponse struct {
	Ready   bool                   `json:"ready"`
	Problem map[string]interface{} `json:"problem,omitempty"`
}

// CreateSession creates a new interview session
func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.FocusMode != "all" && req.FocusMode != "single" && req.FocusMode != "multiple" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Focus mode must be 'all', 'single', or 'multiple'"})
		return
	}

	if req.FocusMode == "single" && req.FocusTopic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Focus topic required for single mode"})
		return
	}

	if req.FocusMode == "multiple" {
		// Trim whitespace and validate topics
		cleanedTopics := make([]string, 0, len(req.FocusTopics))
		seen := make(map[string]bool)
		
		for _, topic := range req.FocusTopics {
			trimmed := strings.TrimSpace(topic)
			if trimmed == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Empty topics not allowed"})
				return
			}
			if seen[trimmed] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate topics not allowed"})
				return
			}
			seen[trimmed] = true
			cleanedTopics = append(cleanedTopics, trimmed)
		}
		
		if len(cleanedTopics) < 2 || len(cleanedTopics) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Multiple mode requires 2-10 unique topics"})
			return
		}
		
		// Use cleaned topics
		req.FocusTopics = cleanedTopics
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	sessionData, firstProblem, err := h.sessionService.CreateSession(
		c.Request.Context(),
		uid,
		req.ProblemCount,
		req.FocusMode,
		req.FocusTopic,
		req.FocusTopics,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	if len(sessionData.Problems) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session created without problems"})
		return
	}

	var createdAt string
	if sessionData.Problems[0].GeneratedAt != nil {
		createdAt = sessionData.Problems[0].GeneratedAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		createdAt = ""
	}

	problemMap := map[string]interface{}{
		"id":           sessionData.Problems[0].ID.String(),
		"title":        firstProblem.Title,
		"description":  firstProblem.Description,
		"focus_area":   firstProblem.FocusArea,
		"rating":       firstProblem.Rating,
		"sample_cases": firstProblem.SampleCases,
		"created_at":   createdAt,
	}

	c.JSON(http.StatusOK, CreateSessionResponse{
		SessionID:    sessionData.Session.ID.String(),
		FirstProblem: problemMap,
	})
}

// GetSession retrieves session with all problems
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	sessionData, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if sessionData.Session.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Transform session problems to include parsed problem data
	type SessionProblemResponse struct {
		ProblemNumber int                     `json:"problem_number"`
		Status        string                  `json:"status"`
		Problem       *map[string]interface{} `json:"problem,omitempty"`
		ErrorMessage  *string                 `json:"error_message,omitempty"`
	}

	problems := make([]SessionProblemResponse, len(sessionData.Problems))
	for i, sp := range sessionData.Problems {
		problems[i] = SessionProblemResponse{
			ProblemNumber: sp.ProblemNumber,
			Status:        sp.Status,
			ErrorMessage:  sp.ErrorMessage,
		}

		if sp.Status == "ready" && len(sp.ProblemData) > 0 {
			var problemResponse models.ProblemGenerationResponse
			if err := json.Unmarshal(sp.ProblemData, &problemResponse); err != nil {
				log.Printf("Failed to unmarshal session problem ID=%s GeneratedAt=%v: %v", sp.ID, sp.GeneratedAt, err)
			} else {
				var createdAt string
				if sp.GeneratedAt != nil {
					createdAt = sp.GeneratedAt.Format("2006-01-02T15:04:05Z07:00")
				}

				problemMap := map[string]interface{}{
					"id":           sp.ID.String(),
					"title":        problemResponse.Title,
					"description":  problemResponse.Description,
					"focus_area":   problemResponse.FocusArea,
					"rating":       problemResponse.Rating,
					"sample_cases": problemResponse.SampleCases,
					"created_at":   createdAt,
				}
				problems[i].Problem = &problemMap
			}
		}
	}

	response := gin.H{
		"id":                    sessionData.Session.ID.String(),
		"problem_count":         sessionData.Session.ProblemCount,
		"focus_mode":            sessionData.Session.FocusMode,
		"focus_topic":           sessionData.Session.FocusTopic,
		"focus_topics":          sessionData.Session.FocusTopics,
		"current_problem_number": sessionData.Session.CurrentProblemNumber,
		"status":                sessionData.Session.Status,
		"problems":              problems,
	}

	c.JSON(http.StatusOK, response)
}

// GetNextProblem checks if next problem is ready and returns it
func (h *SessionHandler) GetNextProblem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	currentNumberStr := c.Param("current_number")
	currentNumber, err := strconv.Atoi(currentNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current number"})
		return
	}

	sessionData, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if sessionData.Session.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	problem, err := h.sessionService.GetNextProblem(c.Request.Context(), sessionID, currentNumber)
	if err != nil {
		c.JSON(http.StatusOK, NextProblemResponse{
			Ready: false,
		})
		return
	}

	problemMap := map[string]interface{}{
		"title":        problem.Title,
		"description":  problem.Description,
		"focus_area":   problem.FocusArea,
		"sample_cases": problem.SampleCases,
	}

	c.JSON(http.StatusOK, NextProblemResponse{
		Ready:   true,
		Problem: problemMap,
	})
}

// CompleteSession marks session as completed
func (h *SessionHandler) CompleteSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	sessionData, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if sessionData.Session.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Validate all problems are ready before allowing completion
	readyCount := 0
	for _, p := range sessionData.Problems {
		if p.Status == "ready" {
			readyCount++
		}
	}

	if readyCount < sessionData.Session.ProblemCount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot complete session: not all problems are ready",
			"ready": readyCount,
			"total": sessionData.Session.ProblemCount,
		})
		return
	}

	if err := h.sessionService.CompleteSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session completed successfully"})
}

// ListActiveSessions returns user's active sessions
func (h *SessionHandler) ListActiveSessions(c *gin.Context) {
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

	sessions, err := h.sessionService.ListActiveSessions(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}
