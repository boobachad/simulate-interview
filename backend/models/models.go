package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FocusArea represents a topic category for interview problems
type FocusArea struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string    `gorm:"type:varchar(255);unique;not null" json:"name"`
	Slug           string    `gorm:"type:varchar(255);unique;not null" json:"slug"`
	PromptGuidance string    `gorm:"type:text" json:"prompt_guidance"`
	CreatedAt      time.Time `json:"created_at"`
}

// TestCase represents a single test case
type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	Explanation    string `json:"explanation,omitempty"`
}

// TestCaseList wrapper for []TestCase to implement Scanner and Valuer interfaces
type TestCaseList []TestCase

// Value implementation for driver.Valuer
func (t TestCaseList) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implementation for sql.Scanner
func (t *TestCaseList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, t)
}

// Problem represents a coding interview problem
type Problem struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string       `gorm:"type:varchar(500);not null" json:"title"`
	Description string       `gorm:"type:text;not null" json:"description"`
	FocusAreaID uuid.UUID    `gorm:"type:uuid;not null" json:"focus_area_id"`
	FocusArea   FocusArea    `gorm:"foreignKey:FocusAreaID" json:"focus_area"`
	SampleCases TestCaseList `gorm:"type:jsonb;not null" json:"sample_cases"`
	HiddenCases TestCaseList `gorm:"type:jsonb;not null" json:"hidden_cases"`
	CreatedAt   time.Time    `gorm:"index" json:"created_at"`
}

// BeforeCreate sets UUID before creating record
func (f *FocusArea) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets UUID before creating record
func (p *Problem) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ProblemGenerationRequest represents the request to generate a problem
type ProblemGenerationRequest struct {
	FocusAreas []string `json:"focus_areas" binding:"required"`
}

// ProblemGenerationResponse represents the LLM response format
type ProblemGenerationResponse struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	FocusArea   string       `json:"focus_area"`
	SampleCases TestCaseList `json:"sample_cases"`
	HiddenCases TestCaseList `json:"hidden_cases"`
}

// ExecutionRequest represents a code execution request
type ExecutionRequest struct {
	Code        string     `json:"code" binding:"required"`
	ProblemID   string     `json:"problem_id" binding:"required"`
	Language    string     `json:"language"` // "cpp", "python", "java", "javascript"
	CustomCases []TestCase `json:"custom_cases"`
	Mode        string     `json:"mode"` // "run" or "submit"
}

// ExecutionResult represents the result of a test case execution
type ExecutionResult struct {
	CaseNumber     int    `json:"case_number"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
	Passed         bool   `json:"passed"`
	Error          string `json:"error,omitempty"`
}

// ExecutionResponse represents the response from code execution
type ExecutionResponse struct {
	Success     bool              `json:"success"`
	Results     []ExecutionResult `json:"results"`
	TotalPassed int               `json:"total_passed"`
	TotalCases  int               `json:"total_cases"`
}

// ============================================================================
// Personalized Interview System Models
// ============================================================================

// UserProfile represents user authentication and profile
type UserProfile struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(255);unique;not null" json:"name"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// BeforeCreate sets UUID before creating record
func (u *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// CodingProfile represents coding platform credentials
type CodingProfile struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Platform   string     `gorm:"type:varchar(50);not null" json:"platform"`
	Username   string     `gorm:"type:varchar(255);not null" json:"username"`
	LastSynced *time.Time `json:"last_synced"`
}

// BeforeCreate sets UUID before creating record
func (c *CodingProfile) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// StatsJSON wrapper for json.RawMessage to implement Scanner and Valuer interfaces
type StatsJSON json.RawMessage

// Value implementation for driver.Valuer
func (s StatsJSON) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.RawMessage(s).MarshalJSON()
}

// Scan implementation for sql.Scanner
func (s *StatsJSON) Scan(value interface{}) error {
	if value == nil {
		*s = StatsJSON("{}")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*s = StatsJSON(bytes)
	return nil
}

// UserStats represents cached user statistics
type UserStats struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Platform  string    `gorm:"type:varchar(50);not null" json:"platform"`
	StatsJSON StatsJSON `gorm:"type:jsonb;not null" json:"stats_json"`
	SyncedAt  time.Time `json:"synced_at"`
}

// BeforeCreate sets UUID before creating record
func (u *UserStats) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserFocusProgress represents user progress per focus area
type UserFocusProgress struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Platform    string    `gorm:"type:varchar(50);not null" json:"platform"`
	Topic       string    `gorm:"type:varchar(255);not null" json:"topic"`
	SolvedCount int       `gorm:"default:0" json:"solved_count"`
}

// BeforeCreate sets UUID before creating record
func (u *UserFocusProgress) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// FocusAreaDynamic represents dynamic focus areas from platform APIs
type FocusAreaDynamic struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Platform     string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_platform_topic" json:"platform"`
	Topic        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_platform_topic" json:"topic"`
	ProblemCount int       `gorm:"not null" json:"problem_count"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BeforeCreate sets UUID before creating record
func (f *FocusAreaDynamic) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// InterviewSession represents an interview session
type InterviewSession struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID               uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ProblemCount         int       `gorm:"not null" json:"problem_count"`
	FocusMode            string    `gorm:"type:varchar(20);not null" json:"focus_mode"`
	FocusTopic           *string   `gorm:"type:varchar(255)" json:"focus_topic"`
	CurrentProblemNumber int       `gorm:"default:1" json:"current_problem_number"`
	Status               string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt            time.Time `json:"created_at"`
}

// BeforeCreate sets UUID before creating record
func (i *InterviewSession) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// ProblemData wrapper for json.RawMessage to implement Scanner and Valuer interfaces
type ProblemData json.RawMessage

// Value implementation for driver.Valuer
func (p ProblemData) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.RawMessage(p).MarshalJSON()
}

// Scan implementation for sql.Scanner
func (p *ProblemData) Scan(value interface{}) error {
	if value == nil {
		*p = ProblemData("{}")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*p = ProblemData(bytes)
	return nil
}

// SessionProblem represents a problem in a session with generation status
type SessionProblem struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID     uuid.UUID   `gorm:"type:uuid;not null" json:"session_id"`
	ProblemNumber int         `gorm:"not null" json:"problem_number"`
	Status        string      `gorm:"type:varchar(20);not null" json:"status"`
	ProblemID     *uuid.UUID  `gorm:"type:uuid" json:"problem_id"`
	ProblemData   ProblemData `gorm:"type:jsonb" json:"problem_data"`
	GeneratedAt   *time.Time  `json:"generated_at"`
	ErrorMessage  *string     `gorm:"type:text" json:"error_message"`
}

// BeforeCreate sets UUID before creating record
func (s *SessionProblem) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// SessionToken represents authentication session token
type SessionToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Token     string    `gorm:"type:varchar(255);unique;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate sets UUID before creating record
func (s *SessionToken) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
