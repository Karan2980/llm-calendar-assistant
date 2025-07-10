package calendar

import (
	"fmt"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"google.golang.org/api/calendar/v3"
)

// GetTodaysEvents retrieves today's events
// GetTodaysEvents retrieves today's events
// GetTodaysEvents retrieves today's events
func (c *Client) GetTodaysEvents() ([]models.Task, error) {
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
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}

	var tasks []models.Task
	for _, event := range events.Items { // Remove unused 'i' variable
		start := event.Start.DateTime
		if start == "" {
			start = event.Start.Date + "T00:00:00+05:30"
		}
		end := event.End.DateTime
		if end == "" {
			end = event.End.Date + "T23:59:59+05:30"
		}
		
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
	
	   createdEvent, err := c.service.Events.Insert("pikobyte006@gmail.com", event).Do()
	   if err != nil {
			   return fmt.Errorf("failed to create event: %v", err)
	   }
	   // Set the EventID on the task (if pointer, else return it)
	   task.EventID = createdEvent.Id
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
	   err := c.service.Events.Delete("pikobyte006@gmail.com", eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete event: %v", err)
	}
	fmt.Printf("🗑️ Event deleted successfully: %s\n", eventID)
	return nil
}

// DeleteEventBySummaryAndTime deletes an event by summary and time (first match)
func (c *Client) DeleteEventBySummaryAndTime(summary, start, end string) error {
	// Search for events in a reasonable window (e.g., +/- 1 day)
	startTime, err1 := time.Parse(time.RFC3339, start)
	endTime, err2 := time.Parse(time.RFC3339, end)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("invalid time format for event deletion")
	}
	windowStart := startTime.Add(-24 * time.Hour)
	windowEnd := endTime.Add(24 * time.Hour)

	events, err := c.service.Events.List("primary").
		TimeMin(windowStart.Format(time.RFC3339)).
		TimeMax(windowEnd.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return fmt.Errorf("unable to search for events: %v", err)
	}

	for _, event := range events.Items {
		if event.Summary == summary &&
			event.Start != nil && event.End != nil &&
			(event.Start.DateTime == start || event.Start.Date == start[:10]) &&
			(event.End.DateTime == end || event.End.Date == end[:10]) {
			// Found a match, delete it
			return c.DeleteEvent(event.Id)
		}
	}
	return fmt.Errorf("no matching event found to delete")
}
