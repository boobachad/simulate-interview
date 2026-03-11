package main

import (
	"log"
	"os"

	"github.com/boobachad/simulate-interview/backend/config"
	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	c := cors.DefaultConfig()
	c.AllowOrigins = []string{
		"http://localhost:3000",
		"http://simulate-interview.localhost:1355",
	}
	c.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	c.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(c))

	// API routes
	api := router.Group("/api")
	{
		// Focus areas
		api.GET("/focus-areas", handlers.GetFocusAreas)

		// Problems
		api.GET("/problems", handlers.GetProblems)
		api.GET("/problems/:id", handlers.GetProblem)
		api.POST("/problems/generate", handlers.GenerateProblem)
		api.POST("/problems/generate-stream", handlers.StreamGenerateProblem)

		// Code execution
		api.POST("/execute", handlers.ExecuteCode)

		// User stats
		api.POST("/stats", handlers.GetUserStats)
	}

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
