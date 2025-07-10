package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"google.golang.org/api/calendar/v3"
)

// GetCalendarService creates an authenticated Calendar client
func GetCalendarService(ctx context.Context) (*calendar.Service, error) {
	fmt.Println("🔍 Setting up authentication...")

	// Try environment variables first
	if os.Getenv("GOOGLE_CLIENT_ID") != "" && os.Getenv("GOOGLE_CLIENT_SECRET") != "" {
		fmt.Println("✅ Using environment variable authentication")
		config := models.GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			AccessToken:  os.Getenv("GOOGLE_ACCESS_TOKEN"),
			RefreshToken: os.Getenv("GOOGLE_REFRESH_TOKEN"),
			TokenExpiry:  os.Getenv("GOOGLE_TOKEN_EXPIRY"),
		}
		return GetCalendarServiceFromEnv(ctx, config)
	}

	// Fallback to file-based auth
	projectRoot := getProjectRoot()
	config := models.AuthConfig{
		ServiceAccountPath: filepath.Join(projectRoot, "service-account.json"),
		CredentialsPath:    filepath.Join(projectRoot, "credentials.json"),
		TokenPath:          filepath.Join(projectRoot, "token.json"),
	}

	fmt.Printf("📁 Looking for credentials in: %s\n", projectRoot)
	
	// Try service account first
	if _, err := os.Stat(config.ServiceAccountPath); err == nil {
		fmt.Println("✅ Using service account authentication")
		return getServiceAccountCalendarService(ctx, config.ServiceAccountPath)
	}
	
	// Fallback to OAuth
	if _, err := os.Stat(config.CredentialsPath); err == nil {
		fmt.Println("✅ Using OAuth file authentication")
		return getOAuthCalendarService(ctx, config)
	}
	
	return nil, fmt.Errorf("no authentication method available. Need either:\n1. Environment variables: GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET\n2. Files: service-account.json or credentials.json in %s", projectRoot)
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	
	if filepath.Base(wd) == "planner" && filepath.Base(filepath.Dir(wd)) == "cmd" {
		return filepath.Dir(filepath.Dir(wd))
	}
	
	if filepath.Base(wd) == "cmd" {
		return filepath.Dir(wd)
	}
	
	return wd
}
