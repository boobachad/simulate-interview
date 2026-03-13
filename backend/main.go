package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boobachad/simulate-interview/backend/config"
	"github.com/boobachad/simulate-interview/backend/database"
	"github.com/boobachad/simulate-interview/backend/handlers"
	"github.com/boobachad/simulate-interview/backend/middleware"
	"github.com/boobachad/simulate-interview/backend/services"
	"github.com/boobachad/simulate-interview/backend/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database with context
	if err := database.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	db := database.GetDB()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	authService := services.NewAuthService(db)
	profileService := services.NewProfileService(db)
	statsService := services.NewStatsService(db, httpClient)
	rateLimiter := utils.NewRateLimiter()

	llmProvider, err := services.NewLLMProvider()
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	generationService := services.NewGenerationService(db, llmProvider, statsService, rateLimiter)
	sessionService := services.NewSessionService(db, generationService, statsService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, profileService)
	profileHandler := handlers.NewProfileHandler(profileService, statsService)
	statsHandler := handlers.NewStatsHandler(statsService)
	focusAreasHandler := handlers.NewFocusAreasHandler()
	sessionHandler := handlers.NewSessionHandler(sessionService)
	generationHandler := handlers.NewGenerationHandler(statsService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{
		"http://localhost:3000",
		"http://simulate-interview.localhost:1355",
		"http://api.simulate-interview.localhost:1355",
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes (public)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)

		// protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(authService))
		{
			// Focus areas
			protected.GET("/focus-areas", focusAreasHandler.GetFocusAreas)

			// Problems
			protected.GET("/problems", handlers.GetProblems)
			protected.GET("/problems/:id", handlers.GetProblem)
			protected.GET("/problems/:id/session", handlers.GetProblemSession)
			protected.POST("/problems/generate", generationHandler.GenerateProblem)
			protected.POST("/problems/generate-stream", handlers.StreamGenerateProblem)

			// Code execution
			protected.POST("/execute", handlers.ExecuteCode)

			// Profile routes
			protected.POST("/profile/setup", profileHandler.Setup)
			protected.GET("/profile", profileHandler.GetProfile)
			protected.PUT("/profile", profileHandler.UpdateProfile)
			protected.POST("/profile/sync", profileHandler.SyncStats)

			// Stats routes
			protected.GET("/stats", statsHandler.GetStats)

			// Session routes
			protected.POST("/sessions", sessionHandler.CreateSession)
			protected.GET("/sessions", sessionHandler.ListActiveSessions)
			protected.GET("/sessions/:session_id", sessionHandler.GetSession)
			protected.GET("/sessions/:session_id/next/:current_number", sessionHandler.GetNextProblem)
			protected.POST("/sessions/:session_id/complete", sessionHandler.CompleteSession)
		}
	}

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cancel all background goroutines first
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
