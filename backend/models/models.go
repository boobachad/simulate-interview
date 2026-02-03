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
