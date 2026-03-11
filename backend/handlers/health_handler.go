package handlers

import (
	"net/http"
	"time"

	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	db := database.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"db":     "disconnected",
			"error":  err.Error(),
			"time":   time.Now(),
		})
		return
	}
	err = sqlDB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"db":     "disconnected",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"db":     "connected",
		"time":   time.Now(),
	})
}
