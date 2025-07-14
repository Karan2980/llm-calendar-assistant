package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/api"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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

	// Create Google Calendar service with interactive token setup
	fmt.Println("🔐 Setting up Google Calendar access...")
	calendarService, err := createCalendarServiceInteractive(ctx, config.Google)
	if err != nil {
		log.Fatalf("❌ Unable to create Calendar service: %v", err)
	}

	// Create API server
	server := api.NewServer(calendarService, config.AI, config.Calendar.TimeZone)

	// Start server
	port := getEnvOrDefault("PORT", "8080")
	fmt.Printf("🚀 Starting API server on port %s...\n", port)
	fmt.Printf("📖 Unified API Endpoint: POST /api/unified\n")
	fmt.Printf("💡 Use Postman to test the unified endpoint!\n")

	log.Fatal(http.ListenAndServe(":"+port, server.Router()))
}

// createCalendarServiceInteractive creates calendar service with interactive token setup
func createCalendarServiceInteractive(ctx context.Context, config models.GoogleConfig) (*calendar.Service, error) {
	// Check if all required credentials are present
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("missing Google OAuth credentials. Please set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET in .env file")
	}

	// Create OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	var token *oauth2.Token

	// Check if we have existing tokens
	if config.AccessToken != "" {
		// Parse token expiry
		var tokenExpiry time.Time
		if config.TokenExpiry != "" {
			var err error
			tokenExpiry, err = time.Parse(time.RFC3339, config.TokenExpiry)
			if err != nil {
				fmt.Printf("⚠️ Warning: Could not parse token expiry: %v\n", err)
				tokenExpiry = time.Now().Add(-time.Hour) // Force refresh
			}
		} else {
			tokenExpiry = time.Now().Add(-time.Hour) // Force refresh if no expiry
		}

		// Create token from config
		token = &oauth2.Token{
			AccessToken:  config.AccessToken,
			RefreshToken: config.RefreshToken,
			Expiry:       tokenExpiry,
			TokenType:    "Bearer",
		}

		// Try to refresh if expired
		if token.Expiry.Before(time.Now()) {
			if config.RefreshToken != "" {
				fmt.Println("🔄 Token expired, attempting to refresh...")
				tokenSource := oauthConfig.TokenSource(ctx, token)
				newToken, err := tokenSource.Token()
				if err == nil {
					fmt.Println("✅ Token refreshed successfully!")
					token = newToken
					showTokenUpdateInfo(token)
				} else {
					fmt.Printf("❌ Failed to refresh token: %v\n", err)
					token = nil // Force re-authentication
				}
			} else {
				fmt.Println("❌ Token expired and no refresh token available!")
				token = nil // Force re-authentication
			}
		} else {
			fmt.Println("✅ Using existing valid token")
		}
	}

	// If no valid token, get new one interactively
	if token == nil {
		var err error
		token, err = getTokenInteractively(ctx, oauthConfig)
		if err != nil {
			return nil, err
		}
	}

	// Create HTTP client with token
	client := oauthConfig.Client(ctx, token)

	// Create calendar service
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	fmt.Println("✅ Google Calendar service created successfully!")
	return service, nil
}

// getTokenInteractively gets OAuth token through interactive process
func getTokenInteractively(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Generate auth URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	
	fmt.Println("\n🔐 Google Calendar Authentication Required")
	fmt.Println("==========================================")
	fmt.Println("1. Open this URL in your browser:")
	fmt.Println(authURL)
	fmt.Println("\n2. Grant permissions to access your Google Calendar")
	fmt.Println("3. Copy the authorization code from the redirect URL")
	fmt.Print("\n4. Enter the authorization code here: ")

	// Read authorization code from user
	reader := bufio.NewReader(os.Stdin)
	authCode, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read authorization code: %v", err)
	}
	
	authCode = strings.TrimSpace(authCode)
	if authCode == "" {
		return nil, fmt.Errorf("no authorization code provided")
	}

	// Exchange authorization code for token
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code for token: %v", err)
	}

	fmt.Println("\n✅ Authentication successful!")
	showTokenUpdateInfo(token)
	
	return token, nil
}

// showTokenUpdateInfo displays token information for updating .env file
func showTokenUpdateInfo(token *oauth2.Token) {
	fmt.Println("\n📝 Update your .env file with these new tokens:")
	fmt.Println("===============================================")
	fmt.Printf("GOOGLE_ACCESS_TOKEN=%s\n", token.AccessToken)
	if token.RefreshToken != "" {
		fmt.Printf("GOOGLE_REFRESH_TOKEN=%s\n", token.RefreshToken)
	}
	fmt.Printf("GOOGLE_TOKEN_EXPIRY=%s\n", token.Expiry.Format(time.RFC3339))
	fmt.Println("\n💡 Save these tokens to avoid re-authentication next time!")
	fmt.Println("Press Enter to continue...")
	
	// Wait for user to press Enter
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
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
			GitHubToken: os.Getenv("GITHUB_TOKEN"),
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
