package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// QueryService handles calendar queries
type QueryService struct {
	client *Client
}

// NewQueryService creates a new query service
func NewQueryService(client *Client) *QueryService {
	return &QueryService{
		client: client,
	}
}

// GetTodaysSchedule returns today's events
func (qs *QueryService) GetTodaysSchedule() ([]models.Task, error) {
	return qs.client.GetTodaysEvents()
}

// GetUpcomingEvents returns events for the next N days
func (qs *QueryService) GetUpcomingEvents(days int) ([]models.Task, error) {
	now := time.Now()
	startTime := now
	endTime := now.AddDate(0, 0, days)

	events, err := qs.client.service.Events.List("primary").
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(int64(100)).
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve upcoming events: %v", err)
	}

	var tasks []models.Task
	for _, event := range events.Items {
		start := event.Start.DateTime
		if start == "" {
			start = event.Start.Date + "T00:00:00+05:30"
		}
		end := event.End.DateTime
		if end == "" {
			end = event.End.Date + "T23:59:59+05:30"
		}

		task := models.Task{
			Summary:     event.Summary,
			Start:       start,
			End:         end,
			Description: event.Description,
			Location:    event.Location,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// SearchEvents searches for events containing a keyword
func (qs *QueryService) SearchEvents(keyword string, days int) ([]models.Task, error) {
	events, err := qs.GetUpcomingEvents(days)
	if err != nil {
		return nil, err
	}

	var matchingEvents []models.Task
	keyword = strings.ToLower(keyword)

	for _, event := range events {
		if strings.Contains(strings.ToLower(event.Summary), keyword) ||
			strings.Contains(strings.ToLower(event.Description), keyword) ||
			strings.Contains(strings.ToLower(event.Location), keyword) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	return matchingEvents, nil
}

// GetEventsByDateRange returns events within a specific date range
func (qs *QueryService) GetEventsByDateRange(startDate, endDate time.Time) ([]models.Task, error) {
	events, err := qs.client.service.Events.List("primary").
		TimeMin(startDate.Format(time.RFC3339)).
		TimeMax(endDate.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(int64(100)).
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events by date range: %v", err)
	}

	var tasks []models.Task
	for _, event := range events.Items {
		start := event.Start.DateTime
		if start == "" {
			start = event.Start.Date + "T00:00:00+05:30"
		}
		end := event.End.DateTime
		if end == "" {
			end = event.End.Date + "T23:59:59+05:30"
		}

		task := models.Task{
			Summary:     event.Summary,
			Start:       start,
			End:         end,
			Description: event.Description,
			Location:    event.Location,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTomorrowsEvents returns tomorrow's events
func (qs *QueryService) GetTomorrowsEvents() ([]models.Task, error) {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
	endOfTomorrow := startOfTomorrow.Add(24 * time.Hour)

	return qs.GetEventsByDateRange(startOfTomorrow, endOfTomorrow)
}

// GetThisWeeksEvents returns events for the current week
func (qs *QueryService) GetThisWeeksEvents() ([]models.Task, error) {
	now := time.Now()
	
	// Calculate start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -(weekday-1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	
	// Calculate end of week (Sunday)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	return qs.GetEventsByDateRange(startOfWeek, endOfWeek)
}

// GetNextWeeksEvents returns events for next week
func (qs *QueryService) GetNextWeeksEvents() ([]models.Task, error) {
	now := time.Now()
	
	// Calculate start of next week
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysUntilNextWeek := 7 - (weekday - 1)
	startOfNextWeek := now.AddDate(0, 0, daysUntilNextWeek)
	startOfNextWeek = time.Date(startOfNextWeek.Year(), startOfNextWeek.Month(), startOfNextWeek.Day(), 0, 0, 0, 0, startOfNextWeek.Location())
	
	// Calculate end of next week
	endOfNextWeek := startOfNextWeek.AddDate(0, 0, 7)

	return qs.GetEventsByDateRange(startOfNextWeek, endOfNextWeek)
}

// GetFreeTimeSlots finds free time slots in a given day
func (qs *QueryService) GetFreeTimeSlots(date time.Time, minDuration time.Duration) ([]models.TimeSlot, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events, err := qs.GetEventsByDateRange(startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// Convert events to time slots
	var busySlots []models.TimeSlot
	for _, event := range events {
		startTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		endTime, err := time.Parse(time.RFC3339, event.End)
		if err != nil {
			continue
		}

		busySlots = append(busySlots, models.TimeSlot{
			Start: startTime,
			End:   endTime,
		})
	}

	// Find free slots
	return qs.findFreeSlots(startOfDay, endOfDay, busySlots, minDuration), nil
}

// findFreeSlots finds free time slots between busy periods
func (qs *QueryService) findFreeSlots(dayStart, dayEnd time.Time, busySlots []models.TimeSlot, minDuration time.Duration) []models.TimeSlot {
	var freeSlots []models.TimeSlot

	// Sort busy slots by start time
	for i := 0; i < len(busySlots)-1; i++ {
		for j := i + 1; j < len(busySlots); j++ {
			if busySlots[i].Start.After(busySlots[j].Start) {
				busySlots[i], busySlots[j] = busySlots[j], busySlots[i]
			}
		}
	}

	// Working hours (9 AM to 6 PM)
	workStart := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 9, 0, 0, 0, dayStart.Location())
	workEnd := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 18, 0, 0, 0, dayStart.Location())

	currentTime := workStart

	for _, busySlot := range busySlots {
		// Skip if busy slot is outside working hours
		if busySlot.End.Before(workStart) || busySlot.Start.After(workEnd) {
			continue
		}

		// Adjust busy slot to working hours
		slotStart := busySlot.Start
		if slotStart.Before(workStart) {
			slotStart = workStart
		}
		slotEnd := busySlot.End
		if slotEnd.After(workEnd) {
			slotEnd = workEnd
		}

		// Check if there's a free slot before this busy slot
		if currentTime.Before(slotStart) && slotStart.Sub(currentTime) >= minDuration {
			freeSlots = append(freeSlots, models.TimeSlot{
				Start: currentTime,
				End:   slotStart,
			})
		}

		// Move current time to end of busy slot
		if slotEnd.After(currentTime) {
			currentTime = slotEnd
		}
	}

	// Check for free time after the last busy slot
	if currentTime.Before(workEnd) && workEnd.Sub(currentTime) >= minDuration {
		freeSlots = append(freeSlots, models.TimeSlot{
			Start: currentTime,
			End:   workEnd,
		})
	}

	return freeSlots
}

// GetNextEvent returns the next upcoming event
func (qs *QueryService) GetNextEvent() (*models.Task, error) {
	now := time.Now()
	events, err := qs.GetUpcomingEvents(30) // Look ahead 30 days

	if err != nil {
		return nil, err
	}

	for _, event := range events {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}

		if eventTime.After(now) {
			return &event, nil
		}
	}

	return nil, fmt.Errorf("no upcoming events found")
}

// GetEventsByType returns events matching a specific type/category
func (qs *QueryService) GetEventsByType(eventType string, days int) ([]models.Task, error) {
	events, err := qs.GetUpcomingEvents(days)
	if err != nil {
		return nil, err
	}

	var matchingEvents []models.Task
	eventType = strings.ToLower(eventType)

	// Define keywords for different event types
	typeKeywords := map[string][]string{
		"meeting":  {"meeting", "call", "conference", "discussion", "standup"},
		"work":     {"work", "office", "project", "task", "deadline"},
		"personal": {"personal", "family", "friend", "birthday", "anniversary"},
		"health":   {"gym", "workout", "doctor", "appointment", "exercise", "fitness"},
		"travel":   {"flight", "travel", "trip", "vacation", "hotel"},
		"food":     {"lunch", "dinner", "breakfast", "meal", "restaurant"},
	}

	keywords, exists := typeKeywords[eventType]
	if !exists {
		// If no predefined keywords, use the eventType itself
		keywords = []string{eventType}
	}

	for _, event := range events {
		eventSummary := strings.ToLower(event.Summary)
		eventDescription := strings.ToLower(event.Description)

		for _, keyword := range keywords {
			if strings.Contains(eventSummary, keyword) || strings.Contains(eventDescription, keyword) {
				matchingEvents = append(matchingEvents, event)
				break
			}
		}
	}

	return matchingEvents, nil
}

// GetBusyHours returns the hours when the user is typically busy
func (qs *QueryService) GetBusyHours(days int) (map[int]int, error) {
	events, err := qs.GetUpcomingEvents(days)
	if err != nil {
		return nil, err
	}

	busyHours := make(map[int]int)

	for _, event := range events {
		startTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		endTime, err := time.Parse(time.RFC3339, event.End)
		if err != nil {
			continue
		}

		// Count busy hours
		for hour := startTime.Hour(); hour <= endTime.Hour(); hour++ {
			busyHours[hour]++
		}
	}

	return busyHours, nil
}

// GetEventCount returns the total number of events in the specified period
func (qs *QueryService) GetEventCount(days int) (int, error) {
	events, err := qs.GetUpcomingEvents(days)
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

// HasConflicts checks if there are any overlapping events
func (qs *QueryService) HasConflicts(days int) ([]models.ConflictPair, error) {
	events, err := qs.GetUpcomingEvents(days)
	if err != nil {
		return nil, err
	}

	var conflicts []models.ConflictPair

	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			event1 := events[i]
			event2 := events[j]

			start1, err1 := time.Parse(time.RFC3339, event1.Start)
			end1, err2 := time.Parse(time.RFC3339, event1.End)
			start2, err3 := time.Parse(time.RFC3339, event2.Start)
			end2, err4 := time.Parse(time.RFC3339, event2.End)

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				continue
			}

			// Check for overlap
			if start1.Before(end2) && end1.After(start2) {
				conflicts = append(conflicts, models.ConflictPair{
					Event1: event1,
					Event2: event2,
				})
			}
		}
	}

	return conflicts, nil
}
