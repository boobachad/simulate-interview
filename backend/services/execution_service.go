package services

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/google/uuid"
)

const (
	EXECUTION_TIMEOUT = 2 * time.Second
	COMPILE_TIMEOUT   = 10 * time.Second
)

// ExecutionService handles C++ code compilation and execution
type ExecutionService struct{}

// NewExecutionService creates a new execution service
func NewExecutionService() *ExecutionService {
	return &ExecutionService{}
}

// Execute compiles and runs C++ code against test cases
func (s *ExecutionService) Execute(code string, test_cases []models.TestCase) ([]models.ExecutionResult, error) {
	// Generate unique identifiers
	execution_id := uuid.New().String()
	source_file := fmt.Sprintf("/tmp/temp_%s.cpp", execution_id)
	binary_file := fmt.Sprintf("/tmp/bin_%s", execution_id)

	// Write source code to file
	err := os.WriteFile(source_file, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	// Cleanup files after execution
	defer os.Remove(source_file)
	defer os.Remove(binary_file)

	// Compile the code
	log.Printf("Compiling code with execution ID: %s", execution_id)
	compile_cmd := exec.Command("g++", "-O3", source_file, "-o", binary_file)
	compile_output, err := compile_cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %s", string(compile_output))
	}

	// Run test cases
	results := make([]models.ExecutionResult, 0, len(test_cases))
	for i, test_case := range test_cases {
		result := s.runTestCase(binary_file, test_case, i+1)
		results = append(results, result)
	}

	return results, nil
}

// runTestCase executes a single test case
func (s *ExecutionService) runTestCase(binary_file string, test_case models.TestCase, case_number int) models.ExecutionResult {
	result := models.ExecutionResult{
		CaseNumber:     case_number,
		Input:          test_case.Input,
		ExpectedOutput: test_case.ExpectedOutput,
	}

	// Create command
	cmd := exec.Command(binary_file)

	// Setup stdin
	cmd.Stdin = strings.NewReader(test_case.Input)

	// Setup stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create a channel to handle timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait with timeout
	select {
	case err := <-done:
		if err != nil {
			result.Error = fmt.Sprintf("Runtime error: %s", stderr.String())
			result.Passed = false
			return result
		}
	case <-time.After(EXECUTION_TIMEOUT):
		cmd.Process.Kill()
		result.Error = "Execution timeout (2s limit exceeded)"
		result.Passed = false
		return result
	}

	// Get output and compare
	actual_output := strings.TrimSpace(stdout.String())
	expected_output := strings.TrimSpace(test_case.ExpectedOutput)

	result.ActualOutput = actual_output

	// If expected output is empty (e.g. custom test case without expectation),
	// treat it as passed provided there was no runtime error (which is handled above).
	// This allows playground execution to be "green" just by running successfully.
	if expected_output == "" {
		result.Passed = true
	} else {
		result.Passed = actual_output == expected_output
	}

	return result
}

// ValidateCode performs basic validation on C++ code
func (s *ExecutionService) ValidateCode(code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("code cannot be empty")
	}

	// Check for main function
	if !strings.Contains(code, "int main") {
		return fmt.Errorf("code must contain a main function")
	}

	return nil
}
