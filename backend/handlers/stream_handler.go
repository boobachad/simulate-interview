package handlers

import (
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
	llmProvider, err := services.NewLLMProvider()
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
	streamChan := make(chan string, 10)
	doneChan := make(chan error, 1)

	// Start streaming in goroutine
	ctx := c.Request.Context()
	go func() {
		defer close(streamChan)
		err := llmProvider.GenerateProblemStream(ctx, request.FocusAreas, streamChan)
		doneChan <- err
	}()

	// Collect full response while streaming to client
	var fullResponse strings.Builder

	// Stream chunks to client
	for {
		select {
		case chunk, ok := <-streamChan:
			if !ok {
				// Channel closed, streaming done
				goto process_response
			}
			fullResponse.WriteString(chunk)

			// Send chunk as SSE
			event := fmt.Sprintf("data: %s\n\n", chunk)
			c.Writer.WriteString(event)
			c.Writer.Flush()

		case err := <-doneChan:
			if err != nil {
				log.Printf("Streaming error: %v", err)
				errorEvent := fmt.Sprintf("event: error\ndata: %s\n\n", err.Error())
				c.Writer.WriteString(errorEvent)
				c.Writer.Flush()
				return
			}
			goto process_response
		}
	}

process_response:
	// Wait for any remaining error
	err = <-doneChan
	if err != nil {
		log.Printf("Streaming error: %v", err)
		errorEvent := fmt.Sprintf("event: error\ndata: %s\n\n", err.Error())
		c.Writer.WriteString(errorEvent)
		c.Writer.Flush()
		return
	}

	// Parse the complete response
	content := fullResponse.String()
	content = utils.ExtractJSON(content)

	var problemResponse models.ProblemGenerationResponse
	err = json.Unmarshal([]byte(content), &problemResponse)
	if err != nil {
		log.Printf("Failed to parse streamed response: %s", content)
		c.Writer.WriteString("event: error\ndata: Failed to parse response\n\n")
		c.Writer.Flush()
		return
	}

	// Find or create focus area
	var focusArea models.FocusArea
	result := database.DB.Where("slug = ?", utils.Slugify(problemResponse.FocusArea)).First(&focusArea)
	if result.Error != nil {
		// If focus area doesn't exist, use the first requested focus area
		database.DB.Where("slug = ?", request.FocusAreas[0]).First(&focusArea)
	}

	// Check if a problem with the same title already exists (Deduplication)
	var existingProblem models.Problem
	if result := database.DB.Where("title = ?", problemResponse.Title).First(&existingProblem); result.Error == nil {
		log.Printf("Found existing problem with title '%s', reusing ID: %s", existingProblem.Title, existingProblem.ID)
		database.DB.Preload("FocusArea").First(&existingProblem, "id = ?", existingProblem.ID)

		problemJSON, _ := json.Marshal(existingProblem)
		completionEvent := fmt.Sprintf("event: complete\ndata: %s\n\n", string(problemJSON))
		c.Writer.WriteString(completionEvent)
		c.Writer.Flush()
		return
	}

	// Save problem to database
	problem := models.Problem{
		ID:          uuid.New(),
		Title:       problemResponse.Title,
		Description: problemResponse.Description,
		FocusAreaID: focusArea.ID,
		SampleCases: problemResponse.SampleCases,
		HiddenCases: problemResponse.HiddenCases,
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
	problemJSON, _ := json.Marshal(problem)
	completionEvent := fmt.Sprintf("event: complete\ndata: %s\n\n", string(problemJSON))
	c.Writer.WriteString(completionEvent)
	c.Writer.Flush()
}
