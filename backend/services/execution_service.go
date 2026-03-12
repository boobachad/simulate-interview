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

// Execute compiles and runs code against test cases
func (s *ExecutionService) Execute(code string, testCases []models.TestCase, language string) ([]models.ExecutionResult, error) {
	// Default to C++ if no language specified
	if language == "" {
		language = "cpp"
	}

	// Validate language
	switch language {
	case "cpp", "python", "java", "javascript":
		// Valid languages
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Generate unique identifiers
	executionID := uuid.New().String()

	var results []models.ExecutionResult
	var err error

	switch language {
	case "cpp":
		results, err = s.executeCpp(code, testCases, executionID)
	case "python":
		results, err = s.executePython(code, testCases, executionID)
	case "java":
		results, err = s.executeJava(code, testCases, executionID)
	case "javascript":
		results, err = s.executeJavaScript(code, testCases, executionID)
	}

	return results, err
}

// runTestCase executes a single test case
func (s *ExecutionService) runTestCase(binaryFile string, testCase models.TestCase, caseNumber int) models.ExecutionResult {
	result := models.ExecutionResult{
		CaseNumber:     caseNumber,
		Input:          testCase.Input,
		ExpectedOutput: testCase.ExpectedOutput,
	}

	// Create command
	cmd := exec.Command(binaryFile)

	// Setup stdin
	cmd.Stdin = strings.NewReader(testCase.Input)

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
	actualOutput := strings.TrimSpace(stdout.String())
	expectedOutput := strings.TrimSpace(testCase.ExpectedOutput)

	result.ActualOutput = actualOutput

	// If expected output is empty (e.g. custom test case without expectation),
	// treat it as passed provided there was no runtime error (which is handled above).
	// This allows playground execution to be "green" just by running successfully.
	if expectedOutput == "" {
		result.Passed = true
	} else {
		result.Passed = actualOutput == expectedOutput
	}

	return result
}

// ValidateCode performs basic validation on code
func (s *ExecutionService) ValidateCode(code string, language string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("code cannot be empty")
	}

	// Language-specific validation
	switch language {
	case "cpp":
		if !strings.Contains(code, "int main") {
			return fmt.Errorf("C++ code must contain a main function")
		}
	case "python":
		// Python doesn't require specific structure
	case "java":
		if !strings.Contains(code, "public static void main") {
			return fmt.Errorf("Java code must contain a main method")
		}
	case "javascript":
		// JavaScript doesn't require specific structure
	}

	return nil
}

// executeCpp compiles and runs C++ code
func (s *ExecutionService) executeCpp(code string, testCases []models.TestCase, executionID string) ([]models.ExecutionResult, error) {
	sourceFile := fmt.Sprintf("/tmp/temp_%s.cpp", executionID)
	binaryFile := fmt.Sprintf("/tmp/bin_%s", executionID)

	err := os.WriteFile(sourceFile, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(sourceFile)
	defer os.Remove(binaryFile)

	log.Printf("Compiling C++ code with execution ID: %s", executionID)
	compileCmd := exec.Command("g++", "-O3", sourceFile, "-o", binaryFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %s", string(compileOutput))
	}

	results := make([]models.ExecutionResult, 0, len(testCases))
	for i, testCase := range testCases {
		result := s.runTestCase(binaryFile, testCase, i+1)
		results = append(results, result)
	}

	return results, nil
}

// executePython runs Python code
func (s *ExecutionService) executePython(code string, testCases []models.TestCase, executionID string) ([]models.ExecutionResult, error) {
	sourceFile := fmt.Sprintf("/tmp/temp_%s.py", executionID)

	err := os.WriteFile(sourceFile, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(sourceFile)

	log.Printf("Running Python code with execution ID: %s", executionID)

	results := make([]models.ExecutionResult, 0, len(testCases))
	for i, testCase := range testCases {
		result := s.runTestCaseWithCommand("python3", []string{sourceFile}, testCase, i+1)
		results = append(results, result)
	}

	return results, nil
}

// executeJava compiles and runs Java code
func (s *ExecutionService) executeJava(code string, testCases []models.TestCase, executionID string) ([]models.ExecutionResult, error) {
	// Extract class name from code
	className := "Main"
	if strings.Contains(code, "public class") {
		parts := strings.Split(code, "public class")
		if len(parts) > 1 {
			classPart := strings.TrimSpace(parts[1])
			endIdx := strings.IndexAny(classPart, " {")
			if endIdx > 0 {
				className = classPart[:endIdx]
			}
		}
	}

	sourceFile := fmt.Sprintf("/tmp/%s_%s.java", className, executionID)
	classDir := fmt.Sprintf("/tmp/java_%s", executionID)

	err := os.MkdirAll(classDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create class directory: %w", err)
	}

	err = os.WriteFile(sourceFile, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(sourceFile)
	defer os.RemoveAll(classDir)

	log.Printf("Compiling Java code with execution ID: %s", executionID)
	compileCmd := exec.Command("javac", "-d", classDir, sourceFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %s", string(compileOutput))
	}

	results := make([]models.ExecutionResult, 0, len(testCases))
	for i, testCase := range testCases {
		result := s.runTestCaseWithCommand("java", []string{"-cp", classDir, className}, testCase, i+1)
		results = append(results, result)
	}

	return results, nil
}

// executeJavaScript runs JavaScript code
func (s *ExecutionService) executeJavaScript(code string, testCases []models.TestCase, executionID string) ([]models.ExecutionResult, error) {
	sourceFile := fmt.Sprintf("/tmp/temp_%s.js", executionID)

	err := os.WriteFile(sourceFile, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(sourceFile)

	log.Printf("Running JavaScript code with execution ID: %s", executionID)

	results := make([]models.ExecutionResult, 0, len(testCases))
	for i, testCase := range testCases {
		result := s.runTestCaseWithCommand("node", []string{sourceFile}, testCase, i+1)
		results = append(results, result)
	}

	return results, nil
}

// runTestCaseWithCommand executes a test case with a custom command
func (s *ExecutionService) runTestCaseWithCommand(command string, args []string, testCase models.TestCase, caseNumber int) models.ExecutionResult {
	result := models.ExecutionResult{
		CaseNumber:     caseNumber,
		Input:          testCase.Input,
		ExpectedOutput: testCase.ExpectedOutput,
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(testCase.Input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

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

	actualOutput := strings.TrimSpace(stdout.String())
	expectedOutput := strings.TrimSpace(testCase.ExpectedOutput)

	result.ActualOutput = actualOutput

	if expectedOutput == "" {
		result.Passed = true
	} else {
		result.Passed = actualOutput == expectedOutput
	}

	return result
}
