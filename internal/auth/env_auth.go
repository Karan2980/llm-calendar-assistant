package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GetCalendarServiceFromEnv creates a calendar service using environment variables
func GetCalendarServiceFromEnv(ctx context.Context, config models.GoogleConfig) (*calendar.Service, error) {
	fmt.Println("🔍 Setting up authentication from environment variables...")

	// Check if all required environment variables are set
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("missing required Google OAuth credentials. Please set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET")
	}

	// Create OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	// Check if we have stored tokens
	if config.AccessToken == "" {
		fmt.Println("❌ No access token found in environment variables.")
		fmt.Println("🔧 Please run the authentication setup first:")
		fmt.Println("   1. Go to: https://console.cloud.google.com/apis/credentials")
		fmt.Println("   2. Create OAuth 2.0 credentials")
		fmt.Println("   3. Add your tokens to .env file")
		return nil, fmt.Errorf("no access token configured")
	}

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

	// Create token
	token := &oauth2.Token{
		AccessToken:  config.AccessToken,
		RefreshToken: config.RefreshToken,
		Expiry:       tokenExpiry,
		TokenType:    "Bearer",
	}

	// Check if token is expired and we have a refresh token
	if token.Expiry.Before(time.Now()) && config.RefreshToken != "" {
		fmt.Println("🔄 Token expired, attempting to refresh...")
		
		// Create token source for automatic refresh
		tokenSource := oauthConfig.TokenSource(ctx, token)
		
		// Get a fresh token
		newToken, err := tokenSource.Token()
		if err != nil {
			fmt.Printf("❌ Failed to refresh token: %v\n", err)
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
		
		fmt.Println("✅ Token refreshed successfully!")
		
		// Update the token
		token = newToken
		
		// Optionally save the new token back to environment/file
		fmt.Println("💡 Consider updating your .env file with the new token:")
		fmt.Printf("GOOGLE_ACCESS_TOKEN=%s\n", token.AccessToken)
		if token.RefreshToken != "" {
			fmt.Printf("GOOGLE_REFRESH_TOKEN=%s\n", token.RefreshToken)
		}
		fmt.Printf("GOOGLE_TOKEN_EXPIRY=%s\n", token.Expiry.Format(time.RFC3339))
	} else if token.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("token expired and no refresh token available. Please re-authenticate")
	}

	// Create HTTP client with the token
	client := oauthConfig.Client(ctx, token)

	// Create calendar service
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	fmt.Println("✅ Environment-based authentication successful!")
	return service, nil
}

// GetTokenFromAuthCode exchanges authorization code for tokens
func GetTokenFromAuthCode(ctx context.Context, config models.GoogleConfig, authCode string) (*oauth2.Token, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	token, err := oauthConfig.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code for token: %v", err)
	}

	return token, nil
}

// GetAuthURL returns the OAuth2 authorization URL
func GetAuthURL(config models.GoogleConfig) string {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	return oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// SaveTokenToFile saves token to a JSON file
func SaveTokenToFile(token *oauth2.Token, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(token)
}

// LoadTokenFromFile loads token from a JSON file
func LoadTokenFromFile(filename string) (*oauth2.Token, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}
