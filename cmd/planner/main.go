package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	// Parse command line arguments
	args := os.Args[1:]
	
	// Create enhanced scheduler
	scheduler := planner.NewEnhancedScheduler(calendarService, config.AI, config.Calendar.TimeZone)

	// Handle different modes based on command line arguments
	if len(args) == 0 {
		// Interactive mode (default)
		if err := scheduler.Run(ctx); err != nil {
			log.Fatalf("❌ Scheduler failed: %v", err)
		}
	} else {
		// Command mode
		command := strings.ToLower(args[0])
		commandArgs := args[1:]

		switch command {
		case "help", "-h", "--help":
			scheduler.ShowHelp()
		case "query", "q":
			if len(commandArgs) > 0 {
				// Quick query mode
				question := strings.Join(commandArgs, " ")
				if err := scheduler.RunQuickQuery(ctx, question); err != nil {
					log.Fatalf("❌ Query failed: %v", err)
				}
			} else {
				// Interactive query mode
				fmt.Println("🤖 Starting Interactive Query Mode...")
				if err := scheduler.RunInteractiveQuery(ctx); err != nil {
					log.Fatalf("❌ Interactive query failed: %v", err)
				}
			}
		case "schedule", "s":
			// Direct scheduling mode
			fmt.Println("📅 Starting Scheduling Mode...")
			if err := scheduler.HandleScheduling(ctx); err != nil {
				log.Fatalf("❌ Scheduling failed: %v", err)
			}
		case "search":
			// Search mode
			fmt.Println("🔍 Starting Search Mode...")
			if err := scheduler.HandleSearch(ctx); err != nil {
				log.Fatalf("❌ Search failed: %v", err)
			}
		case "stats":
			// Statistics mode
			fmt.Println("📊 Starting Statistics Mode...")
			if err := scheduler.HandleStats(ctx); err != nil {
				log.Fatalf("❌ Statistics failed: %v", err)
			}
		case "batch":
			// Batch query mode
			if len(commandArgs) == 0 {
				fmt.Println("❌ Batch mode requires questions as arguments")
				fmt.Println("Example: go run main.go batch \"when is my next meeting?\" \"what time is gym?\"")
				return
			}
			fmt.Println("🔄 Starting Batch Query Mode...")
			if err := scheduler.RunBatchQueries(ctx, commandArgs); err != nil {
				log.Fatalf("❌ Batch queries failed: %v", err)
			}
		case "demo":
			// Demo mode with sample queries
			runDemo(ctx, scheduler)
		default:
			fmt.Printf("❌ Unknown command: %s\n", command)
			fmt.Println("Available commands:")
			fmt.Println("  help     - Show help information")
			fmt.Println("  query    - Ask questions about your calendar")
			fmt.Println("  schedule - Schedule new events")
			fmt.Println("  search   - Search calendar events")
			fmt.Println("  stats    - Show calendar statistics")
			fmt.Println("  batch    - Process multiple queries")
			fmt.Println("  demo     - Run demo with sample queries")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  go run main.go")
			fmt.Println("  go run main.go query \"when is my next meeting?\"")
			fmt.Println("  go run main.go schedule")
			fmt.Println("  go run main.go batch \"gym time?\" \"next meeting?\"")
		}
	}
}

// runDemo runs a demonstration with sample queries
func runDemo(ctx context.Context, scheduler *planner.EnhancedScheduler) {
	fmt.Println("🎬 Running Demo Mode...")
	fmt.Println("This will demonstrate the calendar assistant capabilities.")
	fmt.Println()

	// Sample queries for demonstration
	demoQueries := []string{
		"What's my schedule today?",
		"When is my next meeting?",
		"What time is gym?",
		"When am I free today?",
		"Do I have any work events?",
	}

	fmt.Println("📋 Demo Queries:")
	for i, query := range demoQueries {
		fmt.Printf("  %d. %s\n", i+1, query)
	}
	fmt.Println()

	// Run the demo queries
	if err := scheduler.RunBatchQueries(ctx, demoQueries); err != nil {
		log.Fatalf("❌ Demo failed: %v", err)
	}

	fmt.Println("🎉 Demo completed!")
	fmt.Println("💡 Try running with your own queries:")
	fmt.Println("   go run main.go query \"your question here\"")
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

// printUsage prints usage information
func printUsage() {
	fmt.Println("🤖 Enhanced LLM Calendar Assistant")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run main.go [command] [arguments...]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  (no command)  - Interactive mode with menu")
	fmt.Println("  help          - Show detailed help")
	fmt.Println("  query [q]     - Ask questions about calendar")
	fmt.Println("  schedule [s]  - Schedule new events")
	fmt.Println("  search        - Search calendar events")
	fmt.Println("  stats         - Show calendar statistics")
	fmt.Println("  batch         - Process multiple queries")
	fmt.Println("  demo          - Run demonstration")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  go run main.go")
	fmt.Println("  go run main.go query \"when is my next meeting?\"")
	fmt.Println("  go run main.go schedule")
	fmt.Println("  go run main.go batch \"gym time?\" \"free time?\"")
	fmt.Println("  go run main.go demo")
	fmt.Println()
	fmt.Println("ENVIRONMENT SETUP:")
	fmt.Println("  Required: GITHUB_TOKEN, GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET")
	fmt.Println("  Optional: GOOGLE_ACCESS_TOKEN, GOOGLE_REFRESH_TOKEN")
	fmt.Println("  Create .env file in project root with these variables")
	fmt.Println()
}

// validateConfig validates the configuration
func validateConfig(config models.Config) error {
	if config.AI.GitHubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN is required")
	}
	
	if config.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	
	if config.Google.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	
	return nil
}

// showWelcome displays welcome message
func showWelcome() {
	fmt.Println("🎉 Welcome to Enhanced LLM Calendar Assistant!")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  📅 AI-powered event scheduling")
	fmt.Println("  ❓ Natural language calendar queries")
	fmt.Println("  🔍 Smart event search")
	fmt.Println("  📊 Calendar statistics")
	fmt.Println("  🤖 GitHub Models integration")
	fmt.Println()
}

// init function for initialization
func init() {
	// Set up any initialization if needed
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
