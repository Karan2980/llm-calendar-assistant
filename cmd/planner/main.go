package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Karan2980/llm-planner-golang-project/internal/auth"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/internal/planner"
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
		fmt.Println("❌ GitHub token not configured!")
		fmt.Println("💡 Please add your GitHub token to .env file:")
		fmt.Println("   GITHUB_TOKEN=your_actual_github_token")
		fmt.Println("")
		fmt.Println("🔗 Get your GitHub token from:")
		fmt.Println("   https://github.com/settings/tokens")
		fmt.Println("   Required scopes: No special scopes needed for GitHub Models")
		return
	}

	// Create Google Calendar service
	fmt.Println("🔐 Setting up Google Calendar access...")
	calendarService, err := auth.GetCalendarServiceFromEnv(ctx, config.Google)
	if err != nil {
		log.Fatalf("❌ Unable to create Calendar service: %v", err)
	}

	// Create scheduler
	scheduler := planner.NewScheduler(calendarService, config.AI)

	// Run the scheduler
	if err := scheduler.Run(ctx); err != nil {
		log.Fatalf("❌ Scheduler failed: %v", err)
	}
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	// Try to load .env from current directory or project root
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

	// If we're in cmd/planner, go up 2 levels
	if filepath.Base(wd) == "planner" && filepath.Base(filepath.Dir(wd)) == "cmd" {
		projectRoot := filepath.Dir(filepath.Dir(wd))
		return os.Chdir(projectRoot)
	}

	// If we're in cmd, go up 1 level
	if filepath.Base(wd) == "cmd" {
		projectRoot := filepath.Dir(wd)
		return os.Chdir(projectRoot)
	}

	// Already in project root
	return nil
}

// loadConfig loads configuration from environment variables
func loadConfig() models.Config {
	return models.Config{
		AI: models.AIConfig{
			GitHubToken:       os.Getenv("GITHUB_TOKEN"),
			OpenAIKey:         "", // Not used
			HuggingFaceToken:  "", // Not used
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
