package services

import (
	"context"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/google/uuid"
)

type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (string, *models.UserProfile, error)
	ValidateSession(ctx context.Context, token string) (*models.UserProfile, error)
	CreateSession(ctx context.Context, userID uuid.UUID) (string, error)
	InvalidateSession(ctx context.Context, token string) error
	CreateUser(ctx context.Context, username, password string) (*models.UserProfile, error)
}

type ProfileService interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, leetcodeUsername, codeforcesUsername string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfileData, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, leetcodeUsername, codeforcesUsername string) error
	HasProfile(ctx context.Context, userID uuid.UUID) (bool, error)
}

type StatsService interface {
	SyncStats(ctx context.Context, userID uuid.UUID) error
	GetStats(ctx context.Context, userID uuid.UUID) (*CombinedStats, error)
	BuildPersonalizationContext(ctx context.Context, userID uuid.UUID, focusMode string, focusTopic string, focusTopics []string) (string, error)
}

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, problemCount int, focusMode, focusTopic string, focusTopics []string) (*SessionData, *models.ProblemGenerationResponse, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionData, error)
	GetNextProblem(ctx context.Context, sessionID uuid.UUID, currentNumber int) (*models.ProblemGenerationResponse, error)
	IsNextProblemReady(ctx context.Context, sessionID uuid.UUID, currentNumber int) (bool, error)
	UpdateCurrentProblem(ctx context.Context, sessionID uuid.UUID, problemNumber int) error
	CompleteSession(ctx context.Context, sessionID uuid.UUID) error
	ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]ActiveSessionSummary, error)
}
