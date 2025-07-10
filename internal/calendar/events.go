package calendar

import (
	"fmt"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"google.golang.org/api/calendar/v3"
)

// GetTodaysEvents retrieves today's events
func (c *Client) GetTodaysEvents() ([]models.Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	fmt.Printf("🔍 Looking for events between %s and %s\n", 
		startOfDay.Format("2006-01-02 15:04"), 
		endOfDay.Format("2006-01-02 15:04"))

	events, err := c.service.Events.List("primary").
		TimeMin(startOfDay.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}

	fmt.Printf("🔍 Raw API returned %d events\n", len(events.Items))

	var tasks []models.Task
	for i, event := range events.Items {
		fmt.Printf("  Event %d: %s\n", i+1, event.Summary)
		
		start := event.Start.DateTime
		if start == "" {
			start = event.Start.Date + "T00:00:00+05:30"
		}
		end := event.End.DateTime
		if end == "" {
			end = event.End.Date + "T23:59:59+05:30"
		}
		
		fmt.Printf("    Start: %s\n", start)
		fmt.Printf("    End: %s\n", end)
		
		task := models.Task{
			Summary: event.Summary,
			Start:   start,
			End:     end,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// CreateEvent creates a new calendar event
func (c *Client) CreateEvent(task models.Task) error {
	fmt.Printf("🔍 Creating event: %s\n", task.Summary)
	fmt.Printf("  Start: %s\n", task.Start)
	fmt.Printf("  End: %s\n", task.End)
	
	event := &calendar.Event{
		Summary: task.Summary,
		Start: &calendar.EventDateTime{
			DateTime: task.Start,
			TimeZone: "Asia/Kolkata",
		},
		End: &calendar.EventDateTime{
			DateTime: task.End,
			TimeZone: "Asia/Kolkata",
		},
	}
	
	createdEvent, err := c.service.Events.Insert("primary", event).Do()
	if err != nil {
		return fmt.Errorf("failed to create event: %v", err)
	}
	
	fmt.Printf("✅ Event created successfully with ID: %s\n", createdEvent.Id)
	return nil
}

// CreateMultipleEvents creates multiple events at once
func (c *Client) CreateMultipleEvents(tasks []models.Task) (int, error) {
	successCount := 0
	for _, task := range tasks {
		if err := c.CreateEvent(task); err != nil {
			fmt.Printf("⚠️ Failed to create event '%s': %v\n", task.Summary, err)
		} else {
			successCount++
		}
	}
	return successCount, nil
}

// DeleteEvent deletes an event by ID
func (c *Client) DeleteEvent(eventID string) error {
	err := c.service.Events.Delete("primary", eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete event: %v", err)
	}
	fmt.Printf("🗑️ Event deleted successfully: %s\n", eventID)
	return nil
}
