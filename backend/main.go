package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weekly-dashboard/config"
	"weekly-dashboard/database"
	"weekly-dashboard/handlers"
	"weekly-dashboard/middleware"
	"weekly-dashboard/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.Load()

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database
	if err := database.Seed(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	// Load settings from database (overrides .env values)
	handlers.LoadSettingsFromDB()

	// Initialize services
	authService := services.NewAuthService()
	sheetsService := services.NewSheetsService(authService)
	dashboardService := services.NewDashboardService(sheetsService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, sheetsService)
	screenshotHandler := handlers.NewScreenshotHandler()
	settingsHandler := handlers.NewSettingsHandler()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", handlers.HealthCheck)

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.GET("/google", authHandler.GoogleLogin)
			auth.GET("/callback", authHandler.GoogleCallback)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth())
		{
			// Auth
			protected.GET("/auth/me", authHandler.GetCurrentUser)

			// Dashboard
			protected.GET("/dashboard", dashboardHandler.GetDashboard)
			protected.GET("/months", dashboardHandler.GetAvailableMonths)
			protected.GET("/dashboard/compare", dashboardHandler.CompareDashboard)
			protected.POST("/dashboard/snapshot", dashboardHandler.SaveSnapshot)
			protected.GET("/dashboard/snapshots", dashboardHandler.GetSnapshotsByMonth)
			protected.DELETE("/dashboard/snapshot", dashboardHandler.DeleteSnapshot)

			// Screenshots
			protected.POST("/dashboard/screenshot", screenshotHandler.UploadScreenshot)
			protected.GET("/dashboard/screenshots", screenshotHandler.GetScreenshots)
			protected.GET("/dashboard/screenshot/:id", screenshotHandler.GetScreenshotImage)

			// Settings
			protected.GET("/settings/spreadsheet", settingsHandler.GetSpreadsheetSettings)
			protected.PUT("/settings/spreadsheet", settingsHandler.UpdateSpreadsheetSettings)
		}

		// Public screenshot image endpoint (no auth required for image viewing)
		api.GET("/screenshot/image/:id", screenshotHandler.ServeScreenshotImage)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + config.AppConfig.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", config.AppConfig.Port)
		log.Printf("Frontend URL: %s", config.AppConfig.FrontendURL)
		log.Printf("Google OAuth redirect: %s", config.AppConfig.GoogleRedirectURI)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
