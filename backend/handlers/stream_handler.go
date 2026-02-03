package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StreamGenerateProblem generates a new problem using LLM with streaming
func StreamGenerateProblem(c *gin.Context) {
	var request models.ProblemGenerationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if len(request.FocusAreas) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one focus area must be specified",
		})
		return
	}

	// Create LLM provider
	llm_provider, err := services.NewLLMProvider()
	if err != nil {
		log.Printf("Error creating LLM provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize LLM provider",
		})
		return
	}

	// Set headers for Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Create channel for streaming
	stream_chan := make(chan string, 10)
	done_chan := make(chan error, 1)

	// Start streaming in goroutine
	ctx := context.Background()
	go func() {
		defer close(stream_chan)
		err := llm_provider.GenerateProblemStream(ctx, request.FocusAreas, stream_chan)
		done_chan <- err
	}()

	// Collect full response while streaming to client
	var full_response strings.Builder

	// Stream chunks to client
	for {
		select {
		case chunk, ok := <-stream_chan:
			if !ok {
				// Channel closed, streaming done
				goto process_response
			}
			full_response.WriteString(chunk)

			// Send chunk as SSE
			event := fmt.Sprintf("data: %s\n\n", chunk)
			c.Writer.WriteString(event)
			c.Writer.Flush()

		case err := <-done_chan:
			if err != nil {
				log.Printf("Streaming error: %v", err)
				error_event := fmt.Sprintf("event: error\ndata: %s\n\n", err.Error())
				c.Writer.WriteString(error_event)
				c.Writer.Flush()
				return
			}
			goto process_response
		}
	}

process_response:
	// Wait for any remaining error
	err = <-done_chan
	if err != nil {
		log.Printf("Streaming error: %v", err)
		error_event := fmt.Sprintf("event: error\ndata: %s\n\n", err.Error())
		c.Writer.WriteString(error_event)
		c.Writer.Flush()
		return
	}

	// Parse the complete response
	content := full_response.String()
	content = utils.ExtractJSON(content)

	var problem_response models.ProblemGenerationResponse
	err = json.Unmarshal([]byte(content), &problem_response)
	if err != nil {
		log.Printf("Failed to parse streamed response: %s", content)
		c.Writer.WriteString("event: error\ndata: Failed to parse response\n\n")
		c.Writer.Flush()
		return
	}

	// Find or create focus area
	var focus_area models.FocusArea
	result := database.DB.Where("slug = ?", utils.Slugify(problem_response.FocusArea)).First(&focus_area)
	if result.Error != nil {
		// If focus area doesn't exist, use the first requested focus area
		database.DB.Where("slug = ?", request.FocusAreas[0]).First(&focus_area)
	}

	// Check if a problem with the same title already exists (Deduplication)
	var existingProblem models.Problem
	if result := database.DB.Where("title = ?", problem_response.Title).First(&existingProblem); result.Error == nil {
		log.Printf("Found existing problem with title '%s', reusing ID: %s", existingProblem.Title, existingProblem.ID)
		database.DB.Preload("FocusArea").First(&existingProblem, "id = ?", existingProblem.ID)

		problem_json, _ := json.Marshal(existingProblem)
		completion_event := fmt.Sprintf("event: complete\ndata: %s\n\n", string(problem_json))
		c.Writer.WriteString(completion_event)
		c.Writer.Flush()
		return
	}

	// Save problem to database
	problem := models.Problem{
		ID:          uuid.New(),
		Title:       problem_response.Title,
		Description: problem_response.Description,
		FocusAreaID: focus_area.ID,
		SampleCases: problem_response.SampleCases,
		HiddenCases: problem_response.HiddenCases,
	}

	result = database.DB.Create(&problem)
	if result.Error != nil {
		log.Printf("Error saving problem: %v", result.Error)
		c.Writer.WriteString("event: error\ndata: Failed to save problem\n\n")
		c.Writer.Flush()
		return
	}

	// Load the focus area relationship
	database.DB.Preload("FocusArea").First(&problem, "id = ?", problem.ID)

	log.Printf("Problem generated successfully: %s", problem.Title)

	// Send completion event with problem data
	problem_json, _ := json.Marshal(problem)
	completion_event := fmt.Sprintf("event: complete\ndata: %s\n\n", string(problem_json))
	c.Writer.WriteString(completion_event)
	c.Writer.Flush()
}
