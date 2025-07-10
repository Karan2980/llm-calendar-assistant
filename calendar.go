package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"

// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/google"
// 	"google.golang.org/api/calendar/v3"
// 	"google.golang.org/api/option"
// )

// // GetCalendarService creates an authenticated Calendar client using OAuth
// func GetCalendarService(ctx context.Context) (*calendar.Service, error) {
// 	fmt.Println("🔍 Setting up OAuth authentication...")

// 	b, err := os.ReadFile("credentials.json")
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to read credentials.json: %v", err)
// 	}

// 	// Use full calendar scope for read/write access
// 	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
// 	}

// 	client := getClient(config)
// 	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
// 	}

// 	fmt.Println("✅ OAuth authentication successful!")
// 	return srv, nil
// }

// // Retrieve a token, saves the token, then returns the generated client
// func getClient(config *oauth2.Config) *http.Client {
// 	// The file token.json stores the user's access and refresh tokens, and is
// 	// created automatically when the authorization flow completes for the first time
// 	tokFile := "token.json"
// 	tok, err := tokenFromFile(tokFile)
// 	if err != nil {
// 		tok = getTokenFromWeb(config)
// 		saveToken(tokFile, tok)
// 	}
// 	return config.Client(context.Background(), tok)
// }

// // Request a token from the web, then returns the retrieved token
// func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
// 	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
// 	fmt.Printf("🔐 Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

// 	var authCode string
// 	if _, err := fmt.Scan(&authCode); err != nil {
// 		log.Fatalf("Unable to read authorization code: %v", err)
// 	}

// 	tok, err := config.Exchange(context.TODO(), authCode)
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve token from web: %v", err)
// 	}
// 	return tok
// }

// // Retrieves a token from a local file
// func tokenFromFile(file string) (*oauth2.Token, error) {
// 	f, err := os.Open(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()
// 	tok := &oauth2.Token{}
// 	err = json.NewDecoder(f).Decode(tok)
// 	return tok, err
// }

// // Saves a token to a file path
// func saveToken(path string, token *oauth2.Token) {
// 	fmt.Printf("💾 Saving credential file to: %s\n", path)
// 	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
// 	if err != nil {
// 		log.Fatalf("Unable to cache oauth token: %v", err)
// 	}
// 	defer f.Close()
// 	json.NewEncoder(f).Encode(token)
// }

// // Debug calendar access
// func debugCalendar(srv *calendar.Service) {
// 	fmt.Println("🔍 Debugging calendar access...")

// 	// Test 1: List calendars
// 	fmt.Println("\n📋 Test 1: Listing calendars...")
// 	calendarList, err := srv.CalendarList.List().Do()
// 	if err != nil {
// 		fmt.Printf("❌ Error listing calendars: %v\n", err)
// 		return
// 	}

// 	fmt.Printf("✅ Found %d calendars:\n", len(calendarList.Items))
// 	for i, cal := range calendarList.Items {
// 		fmt.Printf("   %d. %s\n", i+1, cal.Summary)
// 		fmt.Printf("      ID: %s\n", cal.Id)
// 		fmt.Printf("      Access: %s\n", cal.AccessRole)
// 		fmt.Printf("      Primary: %v\n", cal.Primary)
// 		fmt.Println()
// 	}

// 	// Test 2: Try to read events
// 	fmt.Println("📋 Test 2: Reading today's events...")
// 	now := time.Now()
// 	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	endOfDay := startOfDay.Add(24 * time.Hour)

// 	events, err := srv.Events.List("primary").
// 		TimeMin(startOfDay.Format(time.RFC3339)).
// 		TimeMax(endOfDay.Format(time.RFC3339)).
// 		SingleEvents(true).
// 		OrderBy("startTime").
// 		Do()

// 	if err != nil {
// 		fmt.Printf("❌ Cannot read events: %v\n", err)
// 	} else {
// 		fmt.Printf("✅ Successfully read %d events from today\n", len(events.Items))
// 		for i, event := range events.Items {
// 			fmt.Printf("   %d. %s\n", i+1, event.Summary)
// 		}
// 	}

// 	// Test 3: Try to create a test event
// 	fmt.Println("\n📋 Test 3: Creating a test event...")
// 	testEvent := &calendar.Event{
// 		Summary: "🧪 Test Event - LLM Planner",
// 		Start: &calendar.EventDateTime{
// 			DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
// 		},
// 		End: &calendar.EventDateTime{
// 			DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
// 		},
// 	}

// 	createdEvent, err := srv.Events.Insert("primary", testEvent).Do()
// 	if err != nil {
// 		fmt.Printf("❌ Cannot create test event: %v\n", err)
// 	} else {
// 		fmt.Printf("✅ Test event created successfully!\n")
// 		fmt.Printf("   Event ID: %s\n", createdEvent.Id)
// 		fmt.Printf("   Check your Google Calendar now!\n")

// 		// Clean up - delete the test event after 5 seconds
// 		time.Sleep(5 * time.Second)
// 		err = srv.Events.Delete("primary", createdEvent.Id).Do()
// 		if err != nil {
// 			fmt.Printf("⚠️ Could not delete test event: %v\n", err)
// 		} else {
// 			fmt.Printf("🧹 Test event cleaned up\n")
// 		}
// 	}
// }

// // GetExistingEvents retrieves today's events
// func GetExistingEvents(srv *calendar.Service) ([]Task, error) {
// 	now := time.Now()
// 	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	endOfDay := startOfDay.Add(24 * time.Hour)

// 	fmt.Printf("🔍 Looking for events between %s and %s\n",
// 		startOfDay.Format("2006-01-02 15:04"),
// 		endOfDay.Format("2006-01-02 15:04"))

// 	events, err := srv.Events.List("primary").
// 		TimeMin(startOfDay.Format(time.RFC3339)).
// 		TimeMax(endOfDay.Format(time.RFC3339)).
// 		SingleEvents(true).
// 		OrderBy("startTime").
// 		Do()

// 	if err != nil {
// 		return nil, fmt.Errorf("unable to retrieve events: %v", err)
// 	}

// 	fmt.Printf("🔍 Raw API returned %d events\n", len(events.Items))

// 	var tasks []Task
// 	for i, event := range events.Items {
// 		fmt.Printf("  Event %d: %s\n", i+1, event.Summary)

// 		start := event.Start.DateTime
// 		if start == "" {
// 			start = event.Start.Date + "T00:00:00+05:30"
// 		}
// 		end := event.End.DateTime
// 		if end == "" {
// 			end = event.End.Date + "T23:59:59+05:30"
// 		}

// 		fmt.Printf("    Start: %s\n", start)
// 		fmt.Printf("    End: %s\n", end)

// 		tasks = append(tasks, Task{
// 			Summary: event.Summary,
// 			Start:   start,
// 			End:     end,
// 		})
// 	}

// 	return tasks, nil
// }

// // InsertEvent creates a new event
// func InsertEvent(srv *calendar.Service, task Task) error {
// 	fmt.Printf("🔍 Creating event: %s\n", task.Summary)
// 	fmt.Printf("  Start: %s\n", task.Start)
// 	fmt.Printf("  End: %s\n", task.End)

// 	event := &calendar.Event{
// 		Summary: task.Summary,
// 		Start: &calendar.EventDateTime{
// 			DateTime: task.Start,
// 			TimeZone: "Asia/Kolkata",
// 		},
// 		End: &calendar.EventDateTime{
// 			DateTime: task.End,
// 			TimeZone: "Asia/Kolkata",
// 		},
// 	}

// 	createdEvent, err := srv.Events.Insert("primary", event).Do()
// 	if err != nil {
// 		return fmt.Errorf("failed to create event: %v", err)
// 	}

// 	fmt.Printf("✅ Event created successfully with ID: %s\n", createdEvent.Id)
// 	return nil
// }
