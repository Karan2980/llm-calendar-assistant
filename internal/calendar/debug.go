package calendar

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Debug performs comprehensive calendar access debugging
func (c *Client) Debug() error {
	fmt.Println("🔍 Debugging calendar access...")
	
	// Test 1: List calendars
	fmt.Println("\n📋 Test 1: Listing calendars...")
	if err := c.debugListCalendars(); err != nil {
		return err
	}
	
	// Test 2: Try to read events
	fmt.Println("\n📋 Test 2: Reading today's events...")
	if err := c.debugReadEvents(); err != nil {
		return err
	}
	
	// Test 3: Try to create a test event
	fmt.Println("\n📋 Test 3: Creating a test event...")
	return c.debugCreateEvent()
}

func (c *Client) debugListCalendars() error {
	calendarList, err := c.service.CalendarList.List().Do()
	if err != nil {
		fmt.Printf("❌ Error listing calendars: %v\n", err)
		fmt.Println("\n💡 Possible solutions:")
		fmt.Println("   1. Enable Google Calendar API in Google Cloud Console")
		fmt.Println("   2. Check your service account has proper permissions")
		fmt.Println("   3. Verify your service account JSON is valid")
		return err
	}
	
	fmt.Printf("✅ Found %d calendars:\n", len(calendarList.Items))
	for i, cal := range calendarList.Items {
		fmt.Printf("   %d. %s\n", i+1, cal.Summary)
		fmt.Printf("      ID: %s\n", cal.Id)
		fmt.Printf("      Access: %s\n", cal.AccessRole)
		fmt.Printf("      Primary: %v\n", cal.Primary)
		fmt.Println()
	}
	
	return nil
}

func (c *Client) debugReadEvents() error {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	events, err := c.service.Events.List("primary").
		TimeMin(startOfDay.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	
	if err != nil {
		fmt.Printf("❌ Cannot read events: %v\n", err)
		fmt.Println("\n💡 This usually means:")
		fmt.Println("   - Service account doesn't have calendar access")
		fmt.Println("   - Calendar API is not enabled")
		fmt.Println("   - Authentication issue")
		return err
	}
	
	fmt.Printf("✅ Successfully read %d events from today\n", len(events.Items))
	for i, event := range events.Items {
		fmt.Printf("   %d. %s\n", i+1, event.Summary)
	}
	
	return nil
}

func (c *Client) debugCreateEvent() error {
	testEvent := &calendar.Event{
		Summary: "🧪 Test Event - LLM Planner",
		Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		},
	}
	
	createdEvent, err := c.service.Events.Insert("primary", testEvent).Do()
	if err != nil {
		fmt.Printf("❌ Cannot create test event: %v\n", err)
		fmt.Println("\n💡 This confirms permission issues!")
		return err
	}
	
	fmt.Printf("✅ Test event created successfully!\n")
	fmt.Printf("   Event ID: %s\n", createdEvent.Id)
	fmt.Printf("   Check your Google Calendar now!\n")
	
	// Clean up - delete the test event after 5 seconds
	time.Sleep(5 * time.Second)
	err = c.service.Events.Delete("primary", createdEvent.Id).Do()
	if err != nil {
		fmt.Printf("⚠️ Could not delete test event: %v\n", err)
	} else {
		fmt.Printf("🧹 Test event cleaned up\n")
	}
	
	return nil
}
