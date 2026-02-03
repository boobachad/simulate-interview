package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/boobachad/simulate-interview/backend/config"
	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// LLMProvider interface for problem generation
type LLMProvider interface {
	GenerateProblem(focus_areas []string) (*models.ProblemGenerationResponse, error)
	GenerateProblemStream(ctx context.Context, focus_areas []string, stream_chan chan string) error
}

// GeminiProvider implements LLMProvider for Google Gemini
type GeminiProvider struct {
	api_key string
	model   string
}

// OpenRouterProvider implements LLMProvider for OpenRouter
type OpenRouterProvider struct {
	api_key string
	model   string
}

// NewLLMProvider creates an LLM provider based on configuration
// Returns a MockProvider if API keys are not configured
func NewLLMProvider() (LLMProvider, error) {
	provider := config.Config.ActiveProvider

	switch provider {
	case "gemini":
		api_key := os.Getenv("GEMINI_API_KEY")
		if api_key == "" || api_key == "demo_key" {
			log.Println("GEMINI_API_KEY not configured, using mock provider")
			return &MockProvider{}, nil
		}
		return &GeminiProvider{
			api_key: api_key,
			model:   config.Config.Gemini.Model,
		}, nil
	case "openrouter":
		api_key := os.Getenv("OPENROUTER_API_KEY")
		if api_key == "" || api_key == "demo_key" {
			log.Println("OPENROUTER_API_KEY not configured, using mock provider")
			return &MockProvider{}, nil
		}
		return &OpenRouterProvider{
			api_key: api_key,
			model:   config.Config.OpenRouter.Model,
		}, nil
	default:
		log.Println("Unknown provider, using mock provider")
		return &MockProvider{}, nil
	}
}

// GenerateProblem generates a coding problem using Gemini API
func (g *GeminiProvider) GenerateProblem(focus_areas []string) (*models.ProblemGenerationResponse, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.api_key))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(g.model)
	model.SetTemperature(0.9)

	// Fetch focus area guidance from DB
	guidance := ""
	if len(focus_areas) > 0 {
		var fa models.FocusArea
		result := database.DB.Where("slug = ?", utils.Slugify(focus_areas[0])).First(&fa)
		if result.Error == nil && fa.PromptGuidance != "" {
			guidance = fa.PromptGuidance
		}
	}

	prompt := buildPrompt(focus_areas, guidance)

	log.Printf("Generating problem with Gemini for focus areas: %v", focus_areas)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	content := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Extract JSON from markdown code blocks if present
	content = utils.ExtractJSON(content)

	var problem_response models.ProblemGenerationResponse
	err = json.Unmarshal([]byte(content), &problem_response)
	if err != nil {
		log.Printf("Failed to parse Gemini response: %s", content)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &problem_response, nil
}

// GenerateProblemStream generates a coding problem using Gemini API with streaming
func (g *GeminiProvider) GenerateProblemStream(ctx context.Context, focus_areas []string, stream_chan chan string) error {
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.api_key))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(g.model)
	model.SetTemperature(0.9)

	// Fetch focus area guidance from DB
	guidance := ""
	if len(focus_areas) > 0 {
		var fa models.FocusArea
		result := database.DB.Where("slug = ?", utils.Slugify(focus_areas[0])).First(&fa)
		if result.Error == nil && fa.PromptGuidance != "" {
			guidance = fa.PromptGuidance
		}
	}

	prompt := buildPrompt(focus_areas, guidance)

	log.Printf("Streaming problem generation with Gemini for focus areas: %v", focus_areas)

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))

	for {
		resp, err := iter.Next()
		if err != nil {
			if err.Error() == "iterator exhausted" || err.Error() == "no more items in iterator" {
				break
			}
			return fmt.Errorf("error during streaming: %w", err)
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			chunk := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
			stream_chan <- chunk
		}
	}

	return nil
}

// GenerateProblem generates a coding problem using OpenRouter API
func (o *OpenRouterProvider) GenerateProblem(focus_areas []string) (*models.ProblemGenerationResponse, error) {
	// Fetch focus area guidance from DB
	guidance := ""
	if len(focus_areas) > 0 {
		var fa models.FocusArea
		result := database.DB.Where("slug = ?", utils.Slugify(focus_areas[0])).First(&fa)
		if result.Error == nil && fa.PromptGuidance != "" {
			guidance = fa.PromptGuidance
		}
	}

	prompt := buildPrompt(focus_areas, guidance)

	log.Printf("Generating problem with OpenRouter for focus areas: %v", focus_areas)

	request_body := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	json_data, err := json.Marshal(request_body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(json_data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.api_key)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			log.Printf("OpenRouter API returned 401 Unauthorized. Falling back to MockProvider.")
			mockProvider := &MockProvider{}
			return mockProvider.GenerateProblem(focus_areas)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var api_response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &api_response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if len(api_response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenRouter")
	}

	content := api_response.Choices[0].Message.Content
	content = utils.ExtractJSON(content)

	var problem_response models.ProblemGenerationResponse
	err = json.Unmarshal([]byte(content), &problem_response)
	if err != nil {
		log.Printf("Failed to parse OpenRouter response: %s", content)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &problem_response, nil
}

// GenerateProblemStream generates a coding problem using OpenRouter API with streaming
func (o *OpenRouterProvider) GenerateProblemStream(ctx context.Context, focus_areas []string, stream_chan chan string) error {
	// Fetch focus area guidance from DB
	guidance := ""
	if len(focus_areas) > 0 {
		var fa models.FocusArea
		result := database.DB.Where("slug = ?", utils.Slugify(focus_areas[0])).First(&fa)
		if result.Error == nil && fa.PromptGuidance != "" {
			guidance = fa.PromptGuidance
		}
	}

	prompt := buildPrompt(focus_areas, guidance)

	log.Printf("Streaming problem generation with OpenRouter for focus areas: %v", focus_areas)

	request_body := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": true,
	}

	json_data, err := json.Marshal(request_body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.api_key)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			log.Printf("OpenRouter API returned %d. Falling back to MockProvider.", resp.StatusCode)
			mockProvider := &MockProvider{}
			return mockProvider.GenerateProblemStream(ctx, focus_areas, stream_chan)
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read streaming response
	reader := resp.Body
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		chunk := string(buffer[:n])
		// Parse SSE format and extract content
		lines := strings.Split(chunk, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return nil
				}

				var stream_resp struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &stream_resp); err == nil {
					if len(stream_resp.Choices) > 0 && stream_resp.Choices[0].Delta.Content != "" {
						stream_chan <- stream_resp.Choices[0].Delta.Content
					}
				}
			}
		}
	}

	return nil
}

// buildPrompt creates the prompt for problem generation
func buildPrompt(focus_areas []string, guidance string) string {
	focus_str := strings.Join(focus_areas, ", ")

	// Use provided guidance if available
	focus_requirements := ""
	if guidance != "" {
		focus_requirements = fmt.Sprintf(`

FOCUS AREA REQUIREMENTS:
The problem you generate MUST satisfy the focus area requirements below:
%s`, guidance)
	} else if len(focus_areas) > 0 {
		// Fallback generic guidance
		focus_requirements = fmt.Sprintf(`

FOCUS AREA REQUIREMENTS:
The problem you generate MUST satisfy the focus area requirements below:
- Primary Topic: %s
- The problem must fundamentally require knowledge of these specific topics to solve efficienty.
- Do NOT generate a generic array/string problem unless that is the explicit focus.`, focus_str)
	}

	return fmt.Sprintf(`Generate a competitive programming problem.

%s

You must respond with ONLY valid JSON in the following exact format (no markdown, no code blocks, just raw JSON):

{
  "title": "Problem Title",
  "description": "# Problem Description\n\n[Provide a clear story and problem statement here]\n\n## Input Format\n\n[Describe input format]\n\n## Output Format\n\n[Describe output format]\n\n## Constraints\n\n[List constraints]\n\n## Example 1\n**Input:**\n`+"```"+`\n[Input 1]\n`+"```"+`\n**Output:**\n`+"```"+`\n[Output 1]\n`+"```"+`\n**Explanation:**\n[Explanation 1]\n\n## Example 2\n**Input:**\n`+"```"+`\n[Input 2]\n`+"```"+`\n**Output:**\n`+"```"+`\n[Output 2]\n`+"```"+`\n**Explanation:**\n[Explanation 2]",
  "focus_area": "%s",
  "sample_cases": [
    {
      "input": "sample input 1",
      "expected_output": "expected output 1",
      "explanation": "explanation for sample case 1"
    },
    {
      "input": "sample input 2",
      "expected_output": "expected output 2",
      "explanation": "explanation for sample case 2"
    }
  ],
  "hidden_cases": [
    { "input": "hidden input 1", "expected_output": "hidden output 1" },
    { "input": "hidden input 2", "expected_output": "hidden output 2" },
    { "input": "hidden input 3", "expected_output": "hidden output 3" },
    { "input": "hidden input 4", "expected_output": "hidden output 4" },
    { "input": "hidden input 5", "expected_output": "hidden output 5" }
  ]
}

- Must be solvable in C++
- Provide exactly 2 sample cases in the 'sample_cases' array.
- ALSO INCLUDE THESE SAME 2 SAMPLE CASES IN THE 'description' FIELD using the format specified above (## Example 1, ## Example 2).
- Provide exactly 5 hidden test cases in the 'hidden_cases' array.
- Use proper input/output format that can be read from stdin and written to stdout
- Make the problem challenging but solvable in 10-15 minutes
- Include clear constraints in the description
- **CRITICAL**: Append a '## Solution Hints' section at the very end of the 'description'. Checkpoints or algorithmic hints to help a stuck user, but DO NOT give the full code.`, focus_requirements, focus_areas[0])
}

// MockProvider provides a mock problem when API keys are not configured
type MockProvider struct{}

// GenerateProblem returns a mock problem from the mock_problem.json file
func (m *MockProvider) GenerateProblem(focus_areas []string) (*models.ProblemGenerationResponse, error) {
	log.Println("Using mock problem (API keys not configured)")

	// Read mock problem from file
	data, err := os.ReadFile("mock_problem.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read mock problem file: %w", err)
	}

	var mock_response models.ProblemGenerationResponse
	if err := json.Unmarshal(data, &mock_response); err != nil {
		return nil, fmt.Errorf("failed to parse mock problem: %w", err)
	}

	// If specific focus areas are requested, update the mock to reflect them
	if len(focus_areas) > 0 {
		mock_response.FocusArea = focus_areas[0]
		log.Printf("Mock problem adapted for focus area: %s", focus_areas[0])
	}

	return &mock_response, nil
}

// GenerateProblemStream returns a mock problem through streaming
func (m *MockProvider) GenerateProblemStream(ctx context.Context, focus_areas []string, stream_chan chan string) error {
	log.Println("Using mock problem stream (API keys not configured)")

	// Get the mock problem
	problem, err := m.GenerateProblem(focus_areas)
	if err != nil {
		return err
	}

	// Convert to JSON and send through stream
	json_data, err := json.MarshalIndent(problem, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mock problem: %w", err)
	}

	// Send the JSON string through the stream channel
	select {
	case stream_chan <- string(json_data):
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// IsUsingMockProvider checks if the provider is a mock provider
func IsUsingMockProvider(provider LLMProvider) bool {
	_, is_mock := provider.(*MockProvider)
	return is_mock
}

// LoadMockProblem loads the mock problem from file
func LoadMockProblem() (*models.ProblemGenerationResponse, error) {
	data, err := os.ReadFile("mock_problem.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read mock problem file: %w", err)
	}

	var mock_response models.ProblemGenerationResponse
	if err := json.Unmarshal(data, &mock_response); err != nil {
		return nil, fmt.Errorf("failed to parse mock problem: %w", err)
	}

	return &mock_response, nil
}
