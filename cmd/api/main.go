package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Karan2980/llm-planner-golang-project/internal/api"
	"github.com/Karan2980/llm-planner-golang-project/internal/auth"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Change to project root directory
	if err := changeToProjectRoot(); err != nil {
		log.Fatalf("❌ Failed to change to project root: %v", err)
	}

	ctx := context.Background()

	// Load configuration from environment
	config := loadConfig()

	// Check if GitHub token is configured
	if config.AI.GitHubToken == "" || config.AI.GitHubToken == "your_github_token_here" {
		log.Fatal("❌ GitHub token not configured! Please add GITHUB_TOKEN to your .env file")
	}

	// Create Google Calendar service
	fmt.Println("🔐 Setting up Google Calendar access...")
	calendarService, err := auth.GetCalendarServiceFromEnv(ctx, config.Google)
	if err != nil {
		log.Fatalf("❌ Unable to create Calendar service: %v", err)
	}

	// Create API server
	server := api.NewServer(calendarService, config.AI, config.Calendar.TimeZone)

	// Start server
	port := getEnvOrDefault("PORT", "8080")
	fmt.Printf("🚀 Starting API server on port %s...\n", port)
	fmt.Printf("📖 API Documentation:\n")
	fmt.Printf("   POST /api/schedule - Schedule new events\n")
	fmt.Printf("   POST /api/query - Ask questions about calendar\n")
	fmt.Printf("   GET  /api/events/today - Get today's events\n")
	fmt.Printf("   GET  /api/events/upcoming - Get upcoming events\n")
	fmt.Printf("   POST /api/search - Search calendar events\n")
	fmt.Printf("   GET  /api/stats - Get calendar statistics\n")
	fmt.Printf("   GET  /health - Health check\n")
	fmt.Printf("\n💡 Use Postman to test the API endpoints!\n")

	log.Fatal(http.ListenAndServe(":"+port, server.Router()))
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	envPaths := []string{".env", "../.env", "../../.env"}
	
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("✅ Loaded environment from: %s\n", path)
			return nil
		}
	}
	
	return fmt.Errorf("no .env file found")
}

// changeToProjectRoot changes the working directory to the project root
func changeToProjectRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// If we're in cmd/api, go up 2 levels
	if filepath.Base(wd) == "api" && filepath.Base(filepath.Dir(wd)) == "cmd" {
		projectRoot := filepath.Dir(filepath.Dir(wd))
		return os.Chdir(projectRoot)
	}

	// If we're in cmd, go up 1 level
	if filepath.Base(wd) == "cmd" {
		projectRoot := filepath.Dir(wd)
		return os.Chdir(projectRoot)
	}

	return nil
}

// loadConfig loads configuration from environment variables
func loadConfig() models.Config {
	return models.Config{
		AI: models.AIConfig{
			GitHubToken:       os.Getenv("GITHUB_TOKEN"),
			OpenAIKey:         "",
			HuggingFaceToken:  "",
		},
		Calendar: models.CalendarConfig{
			TimeZone: getEnvOrDefault("APP_TIMEZONE", "Asia/Kolkata"),
		},
		Google: models.GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  getEnvOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080"),
			AccessToken:  os.Getenv("GOOGLE_ACCESS_TOKEN"),
			RefreshToken: os.Getenv("GOOGLE_REFRESH_TOKEN"),
			TokenExpiry:  os.Getenv("GOOGLE_TOKEN_EXPIRY"),
		},
	}
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
