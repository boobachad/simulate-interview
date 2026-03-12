package middleware

import (
	"net/http"
	"strings"

	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/gin-gonic/gin"
)

func AuthRequired(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Validate session token
		user, err := authService.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again"})
			c.Abort()
			return
		}

		// Attach user ID to context
		c.Set("user_id", user.ID)
		c.Next()
	}
}
