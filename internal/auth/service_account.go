package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// getServiceAccountCalendarService creates a service using service account
func getServiceAccountCalendarService(ctx context.Context, serviceAccountPath string) (*calendar.Service, error) {
	b, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read service-account.json: %v", err)
	}

	// Parse the service account to get email
	var serviceAccount map[string]interface{}
	if err := json.Unmarshal(b, &serviceAccount); err != nil {
		return nil, fmt.Errorf("invalid JSON in service account: %v", err)
	}

	clientEmail, ok := serviceAccount["client_email"].(string)
	if !ok {
		return nil, fmt.Errorf("client_email not found in service account")
	}
	
	fmt.Printf("🔑 Service Account Email: %s\n", clientEmail)
	fmt.Printf("💡 Make sure this email has access to your calendar!\n")
	fmt.Printf("   Go to Google Calendar → Settings → Share with specific people\n")
	fmt.Printf("   Add: %s with 'Make changes to events' permission\n\n", clientEmail)

	config, err := google.JWTConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse service account key: %v", err)
	}

	client := config.Client(ctx)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %v", err)
	}

	return srv, nil
}
