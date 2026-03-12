package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FocusAreasHandler struct{}

func NewFocusAreasHandler() *FocusAreasHandler {
	return &FocusAreasHandler{}
}

type FocusAreaResponse struct {
	Platform     string `json:"platform"`
	Topic        string `json:"topic"`
	ProblemCount int    `json:"problem_count"`
	UserSolved   *int   `json:"user_solved,omitempty"`
}

// GetFocusAreas retrieves focus areas sorted by problem count
func (h *FocusAreasHandler) GetFocusAreas(c *gin.Context) {
	db := database.GetDB()

	page := 1
	pageSize := 100

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := c.Query("page_size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	offset := (page - 1) * pageSize

	var focusAreas []models.FocusAreaDynamic
	if err := db.WithContext(c.Request.Context()).
		Order("problem_count DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&focusAreas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve focus areas"})
		return
	}

	userID, authenticated := c.Get("user_id")
	var userProgress map[string]int

	if authenticated {
		uid, ok := userID.(uuid.UUID)
		if ok {
			var progressRecords []models.UserFocusProgress
			if err := db.WithContext(c.Request.Context()).
				Where("user_id = ?", uid).
				Find(&progressRecords).Error; err != nil {
				log.Printf("Failed to query user focus progress for user %s: %v", uid, err)
			} else {
				userProgress = make(map[string]int)
				for _, p := range progressRecords {
					key := p.Platform + ":" + p.Topic
					userProgress[key] = p.SolvedCount
				}
			}
		}
	}

	response := make([]FocusAreaResponse, len(focusAreas))
	for i, fa := range focusAreas {
		response[i] = FocusAreaResponse{
			Platform:     fa.Platform,
			Topic:        fa.Topic,
			ProblemCount: fa.ProblemCount,
		}

		if userProgress != nil {
			key := fa.Platform + ":" + fa.Topic
			if solved, exists := userProgress[key]; exists {
				response[i].UserSolved = &solved
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
