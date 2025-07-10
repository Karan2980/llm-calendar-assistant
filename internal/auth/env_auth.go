package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// GetCalendarServiceFromEnv creates calendar service from environment variables
func GetCalendarServiceFromEnv(ctx context.Context, config models.GoogleConfig) (*calendar.Service, error) {
	fmt.Println("🔍 Setting up authentication from environment variables...")

	// Check if we have the required environment variables
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required")
	}

	// Create OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	// Create token from environment variables
	var token *oauth2.Token
	if config.AccessToken != "" && config.RefreshToken != "" {
		expiry, _ := time.Parse(time.RFC3339, config.TokenExpiry)
		token = &oauth2.Token{
			AccessToken:  config.AccessToken,
			RefreshToken: config.RefreshToken,
			Expiry:       expiry,
		}
	} else {
		// If no token in env, get it interactively
		token = getTokenFromWeb(oauthConfig)
		
		// Print the token values so user can add them to .env
		fmt.Println("\n🔑 Add these to your .env file:")
		fmt.Printf("GOOGLE_ACCESS_TOKEN=%s\n", token.AccessToken)
		fmt.Printf("GOOGLE_REFRESH_TOKEN=%s\n", token.RefreshToken)
		fmt.Printf("GOOGLE_TOKEN_EXPIRY=%s\n", token.Expiry.Format(time.RFC3339))
	}

	// Create HTTP client with token
	client := oauthConfig.Client(ctx, token)

	// Create calendar service
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	fmt.Println("✅ Environment-based authentication successful!")
	return srv, nil
}
