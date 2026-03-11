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
func (s *ExecutionService) Execute(code string, test_cases []models.TestCase, language string) ([]models.ExecutionResult, error) {
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
	execution_id := uuid.New().String()

	var results []models.ExecutionResult
	var err error

	switch language {
	case "cpp":
		results, err = s.executeCpp(code, test_cases, execution_id)
	case "python":
		results, err = s.executePython(code, test_cases, execution_id)
	case "java":
		results, err = s.executeJava(code, test_cases, execution_id)
	case "javascript":
		results, err = s.executeJavaScript(code, test_cases, execution_id)
	}

	return results, err
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
func (s *ExecutionService) executeCpp(code string, test_cases []models.TestCase, execution_id string) ([]models.ExecutionResult, error) {
	source_file := fmt.Sprintf("/tmp/temp_%s.cpp", execution_id)
	binary_file := fmt.Sprintf("/tmp/bin_%s", execution_id)

	err := os.WriteFile(source_file, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(source_file)
	defer os.Remove(binary_file)

	log.Printf("Compiling C++ code with execution ID: %s", execution_id)
	compile_cmd := exec.Command("g++", "-O3", source_file, "-o", binary_file)
	compile_output, err := compile_cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %s", string(compile_output))
	}

	results := make([]models.ExecutionResult, 0, len(test_cases))
	for i, test_case := range test_cases {
		result := s.runTestCase(binary_file, test_case, i+1)
		results = append(results, result)
	}

	return results, nil
}

// executePython runs Python code
func (s *ExecutionService) executePython(code string, test_cases []models.TestCase, execution_id string) ([]models.ExecutionResult, error) {
	source_file := fmt.Sprintf("/tmp/temp_%s.py", execution_id)

	err := os.WriteFile(source_file, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(source_file)

	log.Printf("Running Python code with execution ID: %s", execution_id)

	results := make([]models.ExecutionResult, 0, len(test_cases))
	for i, test_case := range test_cases {
		result := s.runTestCaseWithCommand("python3", []string{source_file}, test_case, i+1)
		results = append(results, result)
	}

	return results, nil
}

// executeJava compiles and runs Java code
func (s *ExecutionService) executeJava(code string, test_cases []models.TestCase, execution_id string) ([]models.ExecutionResult, error) {
	// Extract class name from code
	class_name := "Main"
	if strings.Contains(code, "public class") {
		parts := strings.Split(code, "public class")
		if len(parts) > 1 {
			class_part := strings.TrimSpace(parts[1])
			end_idx := strings.IndexAny(class_part, " {")
			if end_idx > 0 {
				class_name = class_part[:end_idx]
			}
		}
	}

	source_file := fmt.Sprintf("/tmp/%s_%s.java", class_name, execution_id)
	class_dir := fmt.Sprintf("/tmp/java_%s", execution_id)

	err := os.MkdirAll(class_dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create class directory: %w", err)
	}

	err = os.WriteFile(source_file, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(source_file)
	defer os.RemoveAll(class_dir)

	log.Printf("Compiling Java code with execution ID: %s", execution_id)
	compile_cmd := exec.Command("javac", "-d", class_dir, source_file)
	compile_output, err := compile_cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %s", string(compile_output))
	}

	results := make([]models.ExecutionResult, 0, len(test_cases))
	for i, test_case := range test_cases {
		result := s.runTestCaseWithCommand("java", []string{"-cp", class_dir, class_name}, test_case, i+1)
		results = append(results, result)
	}

	return results, nil
}

// executeJavaScript runs JavaScript code
func (s *ExecutionService) executeJavaScript(code string, test_cases []models.TestCase, execution_id string) ([]models.ExecutionResult, error) {
	source_file := fmt.Sprintf("/tmp/temp_%s.js", execution_id)

	err := os.WriteFile(source_file, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	defer os.Remove(source_file)

	log.Printf("Running JavaScript code with execution ID: %s", execution_id)

	results := make([]models.ExecutionResult, 0, len(test_cases))
	for i, test_case := range test_cases {
		result := s.runTestCaseWithCommand("node", []string{source_file}, test_case, i+1)
		results = append(results, result)
	}

	return results, nil
}

// runTestCaseWithCommand executes a test case with a custom command
func (s *ExecutionService) runTestCaseWithCommand(command string, args []string, test_case models.TestCase, case_number int) models.ExecutionResult {
	result := models.ExecutionResult{
		CaseNumber:     case_number,
		Input:          test_case.Input,
		ExpectedOutput: test_case.ExpectedOutput,
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(test_case.Input)

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

	actual_output := strings.TrimSpace(stdout.String())
	expected_output := strings.TrimSpace(test_case.ExpectedOutput)

	result.ActualOutput = actual_output

	if expected_output == "" {
		result.Passed = true
	} else {
		result.Passed = actual_output == expected_output
	}

	return result
}
