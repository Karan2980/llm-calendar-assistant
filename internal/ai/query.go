package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/calendar"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/pkg/utils"
)

// QueryProcessor handles natural language queries about calendar
type QueryProcessor struct {
	aiManager      *Manager
	calendarClient *calendar.Client
}

// NewQueryProcessor creates a new query processor
func NewQueryProcessor(aiManager *Manager, calendarClient *calendar.Client) *QueryProcessor {
	return &QueryProcessor{
		aiManager:      aiManager,
		calendarClient: calendarClient,
	}
}

// ProcessQuery processes a natural language query about the calendar
func (q *QueryProcessor) ProcessQuery(question string, context models.QueryContext) (*models.QueryResponse, error) {
	// First try rule-based processing for common queries
	if response := q.processRuleBasedQuery(question, context); response != nil {
		return response, nil
	}

	// Fall back to AI processing for complex queries
	return q.processAIQuery(question, context)
}

// processRuleBasedQuery handles common queries with simple rules
// processRuleBasedQuery handles common queries with simple rules
func (q *QueryProcessor) processRuleBasedQuery(question string, context models.QueryContext) *models.QueryResponse {
	question = strings.ToLower(strings.TrimSpace(question))

	// Check for delete intent first (highest priority)
	if q.isDeleteRequest(question) {
		response := q.handleDeleteIntent(question, context)
		response.Action = "delete"
		return response
	}

	// Check for scheduling queries (must come before other checks)
	if q.isSchedulingQuery(question) {
		response := q.handleSchedulingQuery(question, context)
		response.Action = "create"
		return response
	}

	// Events for tomorrow or a specific day
	if strings.Contains(question, "tomorrow") || strings.Contains(question, "next day") || strings.Contains(question, "day after tomorrow") {
		response := q.getEventsForSpecificDay(question, context)
		response.Action = "view"
		return response
	}

	// Show/display all upcoming events
	if (strings.Contains(question, "upcoming") && strings.Contains(question, "event")) ||
		(strings.Contains(question, "show") && strings.Contains(question, "event")) ||
		(strings.Contains(question, "display") && strings.Contains(question, "event")) {
		response := q.getAllUpcomingEvents(context)
		response.Action = "view"
		return response
	}

	// Next meeting queries
	if strings.Contains(question, "next meeting") || strings.Contains(question, "next event") {
		response := q.findNextEvent(context)
		response.Action = "view"
		return response
	}

	// Gym timing queries
	if strings.Contains(question, "gym") && (strings.Contains(question, "time") || strings.Contains(question, "when")) {
		response := q.findEventByKeyword("gym", context)
		response.Action = "view"
		return response
	}

	// Work timing queries
	if strings.Contains(question, "work") && (strings.Contains(question, "time") || strings.Contains(question, "when")) {
		response := q.findEventByKeyword("work", context)
		response.Action = "view"
		return response
	}

	// Lunch timing queries
	if strings.Contains(question, "lunch") && (strings.Contains(question, "time") || strings.Contains(question, "when")) {
		response := q.findEventByKeyword("lunch", context)
		response.Action = "view"
		return response
	}

	// Today's schedule
	if strings.Contains(question, "today") && (strings.Contains(question, "schedule") || strings.Contains(question, "events")) {
		response := q.getTodaysSchedule(context)
		response.Action = "view"
		return response
	}

	// Free time queries
	if strings.Contains(question, "free") || strings.Contains(question, "available") {
		response := q.findFreeTime(context)
		response.Action = "view"
		return response
	}

	return nil // No rule-based match found
}


// isDeleteRequest checks if the question is asking to delete events
func (q *QueryProcessor) isDeleteRequest(question string) bool {
	question = strings.ToLower(strings.TrimSpace(question))
	deleteKeywords := []string{
		"delete", "remove", "cancel", "clear", "erase", "drop",
		"get rid of", "eliminate", "destroy", "wipe", "purge",
	}
	
	for _, keyword := range deleteKeywords {
		if strings.Contains(question, keyword) {
			return true
		}
	}
	return false
}

// isSchedulingQuery checks if the question is asking to schedule/create events
func (q *QueryProcessor) isSchedulingQuery(question string) bool {
	question = strings.ToLower(strings.TrimSpace(question))
	scheduleKeywords := []string{
		"schedule", "book", "add", "create", "set up", "plan",
		"arrange", "organize", "make", "put", "insert", "place",
	}
	
	timeKeywords := []string{
		"at", "on", "tomorrow", "today", "next week", "monday", "tuesday",
		"wednesday", "thursday", "friday", "saturday", "sunday",
		"am", "pm", "o'clock", ":", "morning", "afternoon", "evening",
	}
	
	hasScheduleKeyword := false
	hasTimeKeyword := false
	
	for _, keyword := range scheduleKeywords {
		if strings.Contains(question, keyword) {
			hasScheduleKeyword = true
			break
		}
	}
	
	for _, keyword := range timeKeywords {
		if strings.Contains(question, keyword) {
			hasTimeKeyword = true
			break
		}
	}
	
	return hasScheduleKeyword && hasTimeKeyword
}

// handleDeleteIntent handles delete requests
func (q *QueryProcessor) handleDeleteIntent(question string, context models.QueryContext) *models.QueryResponse {
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	var eventsToDelete []models.Task
	
	// Check for "all" keyword
	if strings.Contains(question, "all") {
		if strings.Contains(question, "upcoming") {
			// Delete all upcoming events
			for _, event := range allEvents {
				eventTime, err := time.Parse(time.RFC3339, event.Start)
				if err != nil {
					continue
				}
				if eventTime.After(context.CurrentTime) {
					eventsToDelete = append(eventsToDelete, event)
				}
			}
		} else if strings.Contains(question, "today") {
			// Delete all today's events
			eventsToDelete = context.TodaysEvents
		} else {
			// Delete all events
			eventsToDelete = allEvents
		}
	} else {
		// Try to parse a specific date from the question
		dateStr := extractDateFromQuestion(question)
		var targetDate time.Time
		var dateParsed bool
		
		if dateStr != "" {
			// Try multiple formats
			for _, layout := range []string{"2 Jan 2006", "2 January 2006", "02-01-2006", "2006-01-02", "2nd January 2006", "2nd Jan 2006", "2/1/2006", "2.1.2006", "2 july 2006", "2nd july 2006"} {
				t, err := time.Parse(layout, dateStr)
				if err == nil {
					targetDate = t
					dateParsed = true
					break
				}
			}
		}

		if dateParsed {
			// Delete all events on that date
			for _, event := range allEvents {
				eventTime, err := time.Parse(time.RFC3339, event.Start)
				if err != nil {
					continue
				}
				if eventTime.Format("2006-01-02") == targetDate.Format("2006-01-02") {
					eventsToDelete = append(eventsToDelete, event)
				}
			}
		} else {
			// Fallback: match by summary keyword
			keyword := extractEventKeyword(question)
			if keyword != "" {
				for _, event := range allEvents {
					if strings.Contains(strings.ToLower(event.Summary), keyword) {
						eventsToDelete = append(eventsToDelete, event)
					}
				}
			}
		}
	}

	if len(eventsToDelete) == 0 {
		return &models.QueryResponse{
			Answer:  "No matching events found to delete.",
			Success: false,
		}
	}

	// Return events to be deleted (actual deletion handled by API layer)
	var eventNames []string
	for _, event := range eventsToDelete {
		eventNames = append(eventNames, event.Summary)
	}

	answer := fmt.Sprintf("Found %d event(s) to delete: %s", len(eventsToDelete), strings.Join(eventNames, ", "))
	
	return &models.QueryResponse{
		Answer:  answer,
		Success: true,
		Events:  eventsToDelete,
	}
}

// handleAdvancedScheduling handles complex scheduling requests
func (q *QueryProcessor) handleAdvancedScheduling(question string, context models.QueryContext) *models.QueryResponse {
	events := q.parseSchedulingRequest(question, context)
	
	if len(events) == 0 {
		return &models.QueryResponse{
			Answer:  "I couldn't understand the event details. Please specify what you want to schedule and when.",
			Success: false,
			Action:  "create",
		}
	}

	var eventNames []string
	for _, event := range events {
		eventNames = append(eventNames, event.Summary)
	}

	answer := fmt.Sprintf("I'll create %d event(s): %s", len(events), strings.Join(eventNames, ", "))
	
	return &models.QueryResponse{
		Answer:  answer,
		Success: true,
		Events:  events,
		Action:  "create",
	}
}

// parseSchedulingRequest parses natural language scheduling requests
func (q *QueryProcessor) parseSchedulingRequest(question string, context models.QueryContext) []models.Task {
	var events []models.Task
	now := context.CurrentTime
	
	// Enhanced parsing logic
	question = strings.ToLower(question)
	
	// Extract event title with better pattern matching
	title := q.extractEventTitle(question)
	
	// Extract date
	targetDate := q.extractDate(question, now)
	
	// Extract time
	startTime := q.extractTime(question, targetDate)
	
	// Extract duration
	duration := q.extractDuration(question)
	
	endTime := startTime.Add(duration)
	
	event := models.Task{
		Summary: title,
		Start:   startTime.Format(time.RFC3339),
		End:     endTime.Format(time.RFC3339),
	}
	
	events = append(events, event)
	return events
}

// extractEventTitle extracts the event title from the question
func (q *QueryProcessor) extractEventTitle(question string) string {
	// Common patterns for event titles
	patterns := map[string]string{
		"team meet":     "Team Meeting",
		"team meeting":  "Team Meeting",
		"standup":       "Standup Meeting",
		"daily standup": "Daily Standup",
		"gym":           "Gym Session",
		"workout":       "Workout",
		"lunch":         "Lunch",
		"dinner":        "Dinner",
		"breakfast":     "Breakfast",
		"call":          "Call",
		"interview":     "Interview",
		"presentation":  "Presentation",
		"review":        "Review Meeting",
		"planning":      "Planning Session",
	}
	
	for pattern, title := range patterns {
		if strings.Contains(question, pattern) {
			return title
		}
	}
	
	// Try to extract from "create/schedule X" patterns
	words := strings.Fields(question)
	for i, word := range words {
		if (word == "create" || word == "schedule" || word == "add" || word == "book") && i+1 < len(words) {
			nextWord := words[i+1]
			if nextWord == "a" || nextWord == "an" {
				if i+2 < len(words) {
					return strings.Title(words[i+2])
				}
			} else {
				return strings.Title(nextWord)
			}
		}
	}
	
	return "Event"
}

// extractDate extracts the target date from the question
func (q *QueryProcessor) extractDate(question string, now time.Time) time.Time {
	if strings.Contains(question, "tomorrow") {
		return now.AddDate(0, 0, 1)
	} else if strings.Contains(question, "today") {
		return now
	} else if strings.Contains(question, "next week") {
		return now.AddDate(0, 0, 7)
	} else if strings.Contains(question, "monday") {
		return q.getNextWeekday(now, time.Monday)
	} else if strings.Contains(question, "tuesday") {
		return q.getNextWeekday(now, time.Tuesday)
	} else if strings.Contains(question, "wednesday") {
		return q.getNextWeekday(now, time.Wednesday)
	} else if strings.Contains(question, "thursday") {
		return q.getNextWeekday(now, time.Thursday)
	} else if strings.Contains(question, "friday") {
		return q.getNextWeekday(now, time.Friday)
	} else if strings.Contains(question, "saturday") {
		return q.getNextWeekday(now, time.Saturday)
	} else if strings.Contains(question, "sunday") {
		return q.getNextWeekday(now, time.Sunday)
	}
	
	return now.AddDate(0, 0, 1) // Default to tomorrow
}

// extractTime extracts the time from the question
// extractTime extracts the time from the question
func (q *QueryProcessor) extractTime(question string, targetDate time.Time) time.Time {
	// Time patterns
	timePatterns := map[string]int{
		"1:00 pm": 13, "1 pm": 13, "1pm": 13,
		"2:00 pm": 14, "2 pm": 14, "2pm": 14,
		"3:00 pm": 15, "3 pm": 15, "3pm": 15,
		"4:00 pm": 16, "4 pm": 16, "4pm": 16,
		"5:00 pm": 17, "5 pm": 17, "5pm": 17,
		"9:00 am": 9, "9 am": 9, "9am": 9,
		"10:00 am": 10, "10 am": 10, "10am": 10,
		"11:00 am": 11, "11 am": 11, "11am": 11,
		"12:00 pm": 12, "12 pm": 12, "12pm": 12,
	}
	
	for pattern, hour := range timePatterns {
		if strings.Contains(question, pattern) {
			return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), hour, 0, 0, 0, targetDate.Location())
		}
	}
	
	// Default to 9 AM
	return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 9, 0, 0, 0, targetDate.Location())
}

// extractDuration extracts the duration from the question
func (q *QueryProcessor) extractDuration(question string) time.Duration {
	if strings.Contains(question, "2 hours") || strings.Contains(question, "2 hour") {
		return 2 * time.Hour
	} else if strings.Contains(question, "3 hours") || strings.Contains(question, "3 hour") {
		return 3 * time.Hour
	} else if strings.Contains(question, "30 min") || strings.Contains(question, "30 minutes") {
		return 30 * time.Minute
	} else if strings.Contains(question, "45 min") || strings.Contains(question, "45 minutes") {
		return 45 * time.Minute
	} else if strings.Contains(question, "1 hour") {
		return time.Hour
	}
	
	return time.Hour // Default to 1 hour
}

// getNextWeekday returns the next occurrence of the specified weekday
func (q *QueryProcessor) getNextWeekday(now time.Time, weekday time.Weekday) time.Time {
	daysUntil := int(weekday - now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return now.AddDate(0, 0, daysUntil)
}

// processAIQuery uses AI to understand and process the query
func (q *QueryProcessor) processAIQuery(question string, context models.QueryContext) (*models.QueryResponse, error) {
	if !q.aiManager.HasClients() {
		// Fallback to enhanced rule-based processing
		response := q.processEnhancedRuleBasedQuery(question, context)
		if response != nil {
			return response, nil
		}
		return &models.QueryResponse{
			Answer:  "I can help with basic calendar queries, but AI-powered responses are not available right now.",
			Success: false,
			Error:   "No AI clients configured",
		}, nil
	}

	prompt := q.createUnifiedPrompt(question, context)
	aiResponse, err := q.aiManager.GeneratePlan(prompt)
	if err != nil {
		// Fallback to enhanced rule-based processing
		response := q.processEnhancedRuleBasedQuery(question, context)
		if response != nil {
			return response, nil
		}
		return &models.QueryResponse{
			Answer:  "I'm sorry, I couldn't process your request right now. Please try again later.",
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Try to parse AI response as JSON
	var response models.QueryResponse
	if err := json.Unmarshal([]byte(aiResponse), &response); err != nil {
		// If JSON parsing fails, treat as plain text answer
		return &models.QueryResponse{
			Answer:  aiResponse,
			Success: true,
			Action:  "view",
		}, nil
	}

	return &response, nil
}

// processEnhancedRuleBasedQuery provides enhanced rule-based processing
func (q *QueryProcessor) processEnhancedRuleBasedQuery(question string, context models.QueryContext) *models.QueryResponse {
	question = strings.ToLower(strings.TrimSpace(question))

	// Enhanced scheduling detection
	if q.isSchedulingQuery(question) {
		return q.handleAdvancedScheduling(question, context)
	}

	// Enhanced deletion detection
	if q.isDeleteRequest(question) {
		response := q.handleDeleteIntent(question, context)
		response.Action = "delete"
		return response
	}

	// All other queries are treated as view operations
	if response := q.processViewQuery(question, context); response != nil {
		response.Action = "view"
		return response
	}

	return &models.QueryResponse{
		Answer:  "I'm not sure how to help with that. You can ask me to create events, view your schedule, or delete events.",
		Success: false,
		Action:  "view",
	}
}

// processViewQuery handles view-only queries
func (q *QueryProcessor) processViewQuery(question string, context models.QueryContext) *models.QueryResponse {
	question = strings.ToLower(strings.TrimSpace(question))

	// Events for tomorrow or a specific day
	if strings.Contains(question, "tomorrow") || strings.Contains(question, "next day") || strings.Contains(question, "day after tomorrow") {
		return q.getEventsForSpecificDay(question, context)
	}

	// Show/display all upcoming events
	if (strings.Contains(question, "upcoming") && strings.Contains(question, "event")) ||
		(strings.Contains(question, "show") && strings.Contains(question, "event")) ||
		(strings.Contains(question, "display") && strings.Contains(question, "event")) {
		return q.getAllUpcomingEvents(context)
	}

	// Next meeting queries
	if strings.Contains(question, "next meeting") || strings.Contains(question, "next event") {
		return q.findNextEvent(context)
	}

	// Gym timing queries
	if strings.Contains(question, "gym") && (strings.Contains(question, "time") || strings.Contains(question, "when")) {
		return q.findEventByKeyword("gym", context)
	}

	// Work timing queries
	if strings.Contains(question, "work") && (strings.Contains(question, "time") || strings.Contains(question, "when")) {
		return q.findEventByKeyword("work", context)
	}

	// Lunch timing queries
	if strings.Contains(question, "lunch") && (strings.Contains(question, "time") || strings.Contains(question, "when")) {
		return q.findEventByKeyword("lunch", context)
	}

	// Today's schedule
	if strings.Contains(question, "today") && (strings.Contains(question, "schedule") || strings.Contains(question, "events")) {
		return q.getTodaysSchedule(context)
	}

	// Free time queries
	if strings.Contains(question, "free") || strings.Contains(question, "available") {
		return q.findFreeTime(context)
	}

	return nil
}

// createUnifiedPrompt creates a prompt for AI processing
func (q *QueryProcessor) createUnifiedPrompt(question string, context models.QueryContext) string {
	currentTime := context.CurrentTime.Format("2006-01-02 15:04:05")
	
	// Create a summary of current events
	var eventSummary string
	if len(context.TodaysEvents) > 0 {
		eventSummary += fmt.Sprintf("Today's events (%d):\n", len(context.TodaysEvents))
		for _, event := range context.TodaysEvents {
			eventSummary += fmt.Sprintf("- %s at %s\n", event.Summary, utils.FormatTime(event.Start))
		}
	}
	
	if len(context.UpcomingEvents) > 0 {
		eventSummary += fmt.Sprintf("Upcoming events (%d):\n", len(context.UpcomingEvents))
		for _, event := range context.UpcomingEvents {
			eventSummary += fmt.Sprintf("- %s at %s\n", event.Summary, utils.FormatTime(event.Start))
		}
	}
	
	if eventSummary == "" {
		eventSummary = "No events scheduled."
	}

	prompt := fmt.Sprintf(`
You are a calendar assistant that can CREATE, VIEW, and DELETE events. Analyze the user's request and respond with a JSON object.

Current time: %s

Calendar events:
%s

User request: "%s"

Respond with a JSON object in this format:
{
  "answer": "Your response to the user",
  "success": true,
  "action": "create|view|delete",
  "events": [
    {
      "summary": "Event Title",
      "start": "2024-01-15T14:00:00Z",
      "end": "2024-01-15T15:00:00Z",
      "location": "Optional location",
      "description": "Optional description"
    }
  ]
}

Rules:
1. For CREATE requests: Set action to "create" and include event details in events array
2. For VIEW requests: Set action to "view" and include relevant events in events array
3. For DELETE requests: Set action to "delete" and include events to delete in events array
4. Always use RFC3339 format for dates (YYYY-MM-DDTHH:MM:SSZ)
5. If creating events, avoid conflicts with existing events
6. Be helpful and conversational in your answer

Examples:
- "Schedule gym tomorrow at 2pm" → action: "create"
- "What's my schedule today?" → action: "view"
- "Delete my gym session" → action: "delete"
`, currentTime, eventSummary, question)

	return prompt
}

// getEventsForSpecificDay gets events for a specific day
func (q *QueryProcessor) getEventsForSpecificDay(question string, context models.QueryContext) *models.QueryResponse {
	var targetDate time.Time
	now := context.CurrentTime
	
	explicitDate := extractDateFromQuestion(question)
	if explicitDate != "" {
		parsed := false
		for _, layout := range []string{"2 Jan 2006", "2 January 2006", "02-01-2006", "2006-01-02", "2nd January 2006", "2nd Jan 2006", "2/1/2006", "2.1.2006", "2 july 2006", "2nd july 2006", "02/01/2006", "13/07/2025", "13-07-2025"} {
			t, err := time.Parse(layout, explicitDate)
			if err == nil {
				targetDate = t
				parsed = true
				break
			}
		}
		if !parsed {
			targetDate = now
		}
	} else if strings.Contains(question, "tomorrow") {
		targetDate = now.AddDate(0, 0, 1)
	} else if strings.Contains(question, "day after tomorrow") {
		targetDate = now.AddDate(0, 0, 2)
	} else if strings.Contains(question, "next day") {
		targetDate = now.AddDate(0, 0, 1)
	} else {
		targetDate = now
	}
	
	dateStr := targetDate.Format("2006-01-02")
	var eventsForDay []models.Task
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	
	for _, event := range allEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		if eventTime.Format("2006-01-02") == dateStr {
			eventsForDay = append(eventsForDay, event)
		}
	}
	
	if len(eventsForDay) == 0 {
		return &models.QueryResponse{
			Answer:  fmt.Sprintf("You have no events scheduled for %s.", targetDate.Format("Monday, January 2")),
			Success: true,
		}
	}
	
	answer := fmt.Sprintf("You have %d event(s) on %s:\n", len(eventsForDay), targetDate.Format("Monday, January 2"))
	for i, event := range eventsForDay {
		answer += fmt.Sprintf("%d. %s at %s", i+1, event.Summary, utils.FormatTime(event.Start))
		if event.Location != "" {
			answer += fmt.Sprintf(" (%s)", event.Location)
		}
		answer += "\n"
	}
	
	return &models.QueryResponse{
		Answer:  strings.TrimSpace(answer),
		Events:  eventsForDay,
		Success: true,
	}
}

// getAllUpcomingEvents returns all upcoming events
func (q *QueryProcessor) getAllUpcomingEvents(context models.QueryContext) *models.QueryResponse {
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	
	if len(allEvents) == 0 {
		return &models.QueryResponse{
			Answer:  "You have no upcoming events scheduled.",
			Success: true,
		}
	}

	answer := fmt.Sprintf("You have %d upcoming event(s):\n", len(allEvents))
	for i, event := range allEvents {
		answer += fmt.Sprintf("%d. %s at %s", i+1, event.Summary, utils.FormatTime(event.Start))
		if event.Location != "" {
			answer += fmt.Sprintf(" (%s)", event.Location)
		}
		answer += "\n"
	}

	return &models.QueryResponse{
		Answer:  strings.TrimSpace(answer),
		Events:  allEvents,
		Success: true,
	}
}

// findNextEvent finds the next upcoming event
// findNextEvent finds the next upcoming event
func (q *QueryProcessor) findNextEvent(context models.QueryContext) *models.QueryResponse {
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	
	if len(allEvents) == 0 {
		return &models.QueryResponse{
			Answer:  "You have no upcoming events scheduled.",
			Success: true,
		}
	}

	// Find the next event (events should be sorted by start time)
	var nextEvent *models.Task
	for _, event := range allEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		if eventTime.After(context.CurrentTime) {
			nextEvent = &event
			break
		}
	}

	if nextEvent == nil {
		return &models.QueryResponse{
			Answer:  "You have no upcoming events scheduled.",
			Success: true,
		}
	}

	answer := fmt.Sprintf("Your next event is: %s at %s", nextEvent.Summary, utils.FormatTime(nextEvent.Start))
	if nextEvent.Location != "" {
		answer += fmt.Sprintf(" (%s)", nextEvent.Location)
	}

	return &models.QueryResponse{
		Answer:  answer,
		Events:  []models.Task{*nextEvent},
		Success: true,
	}
}

// findEventByKeyword finds events by keyword
func (q *QueryProcessor) findEventByKeyword(keyword string, context models.QueryContext) *models.QueryResponse {
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	var matchingEvents []models.Task
	
	for _, event := range allEvents {
		if strings.Contains(strings.ToLower(event.Summary), keyword) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	if len(matchingEvents) == 0 {
		return &models.QueryResponse{
			Answer:  fmt.Sprintf("No events found related to '%s'.", keyword),
			Success: true,
		}
	}

	answer := fmt.Sprintf("Found %d event(s) related to '%s':\n", len(matchingEvents), keyword)
	for i, event := range matchingEvents {
		answer += fmt.Sprintf("%d. %s at %s", i+1, event.Summary, utils.FormatTime(event.Start))
		if event.Location != "" {
			answer += fmt.Sprintf(" (%s)", event.Location)
		}
		answer += "\n"
	}

	return &models.QueryResponse{
		Answer:  strings.TrimSpace(answer),
		Events:  matchingEvents,
		Success: true,
	}
}

// getTodaysSchedule returns today's schedule
func (q *QueryProcessor) getTodaysSchedule(context models.QueryContext) *models.QueryResponse {
	if len(context.TodaysEvents) == 0 {
		return &models.QueryResponse{
			Answer:  "You have no events scheduled for today.",
			Success: true,
		}
	}

	answer := fmt.Sprintf("Today's schedule (%d event(s)):\n", len(context.TodaysEvents))
	for i, event := range context.TodaysEvents {
		answer += fmt.Sprintf("%d. %s at %s", i+1, event.Summary, utils.FormatTime(event.Start))
		if event.Location != "" {
			answer += fmt.Sprintf(" (%s)", event.Location)
		}
		answer += "\n"
	}

	return &models.QueryResponse{
		Answer:  strings.TrimSpace(answer),
		Events:  context.TodaysEvents,
		Success: true,
	}
}

// findFreeTime finds available free time slots
func (q *QueryProcessor) findFreeTime(context models.QueryContext) *models.QueryResponse {
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	
	if len(allEvents) == 0 {
		return &models.QueryResponse{
			Answer:  "You have no scheduled events, so you're free all day!",
			Success: true,
		}
	}

	// Simple free time calculation - find gaps between events
	now := context.CurrentTime
	var freeSlots []string
	
	// Check if free right now
	hasCurrentEvent := false
	for _, event := range allEvents {
		startTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		endTime, err := time.Parse(time.RFC3339, event.End)
		if err != nil {
			endTime = startTime.Add(time.Hour) // Default 1 hour duration
		}
		
		if now.After(startTime) && now.Before(endTime) {
			hasCurrentEvent = true
			break
		}
	}
	
	if !hasCurrentEvent {
		freeSlots = append(freeSlots, "Right now")
	}
	
	// Find next free slot after current time
	nextFreeTime := now.Add(time.Hour)
	for _, event := range allEvents {
		startTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		if startTime.After(nextFreeTime) {
			freeSlots = append(freeSlots, fmt.Sprintf("After %s", utils.FormatTime(nextFreeTime.Format(time.RFC3339))))
			break
		}
	}
	
	if len(freeSlots) == 0 {
		return &models.QueryResponse{
			Answer:  "Your schedule looks quite busy. Consider checking for longer gaps between events.",
			Success: true,
		}
	}
	
	answer := fmt.Sprintf("You appear to be free: %s", strings.Join(freeSlots, ", "))
	return &models.QueryResponse{
		Answer:  answer,
		Success: true,
	}
}

// Helper functions for date/time extraction
func extractDateFromQuestion(question string) string {
	lower := strings.ToLower(question)
	words := strings.Fields(lower)
	
	for i := 0; i < len(words)-2; i++ {
		if isDay(words[i]) && isMonth(words[i+1]) && isYear(words[i+2]) {
			return words[i] + " " + words[i+1] + " " + words[i+2]
		}
	}
	
	for _, word := range words {
		if len(word) == 10 && word[4] == '-' && word[7] == '-' {
			return word
		}
	}
	
	for _, word := range words {
		if len(word) == 10 && (word[2] == '/' || word[2] == '-') && (word[5] == '/' || word[5] == '-') {
			return word
		}
	}
	
	return ""
}

func isDay(s string) bool {
	s = strings.TrimSuffix(s, "st")
	s = strings.TrimSuffix(s, "nd")
	s = strings.TrimSuffix(s, "rd")
	s = strings.TrimSuffix(s, "th")
	if len(s) == 1 || len(s) == 2 {
		for _, c := range s {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	return false
}

func isMonth(s string) bool {
	months := []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
	for _, m := range months {
		if s == m {
			return true
		}
	}
	return false
}

func isYear(s string) bool {
	if len(s) == 4 {
		for _, c := range s {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// extractEventKeyword tries to extract a keyword from the delete question
func extractEventKeyword(question string) string {
	lower := strings.ToLower(question)
	for _, kw := range []string{"delete", "remove", "cancel", "clear"} {
		idx := strings.Index(lower, kw)
		if idx != -1 {
			rest := strings.TrimSpace(lower[idx+len(kw):])
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}
// handleSchedulingQuery processes scheduling queries and sets action to "create"
func (q *QueryProcessor) handleSchedulingQuery(question string, context models.QueryContext) *models.QueryResponse {
	if !q.aiManager.HasClients() {
		return &models.QueryResponse{
			Answer:  "I can't schedule events right now because AI service is not available. Please use the /api/schedule endpoint instead.",
			Success: false,
			Action:  "create",
			Error:   "No AI clients configured",
		}
	}

	prompt := q.createSchedulingPrompt(question, context)
	aiResponse, err := q.aiManager.GeneratePlan(prompt)
	if err != nil {
		return &models.QueryResponse{
			Answer:  "Sorry, I couldn't understand your scheduling request. Please try rephrasing it.",
			Success: false,
			Action:  "create",
			Error:   err.Error(),
		}
	}

	// Parse the AI response to extract event details
	tasks, err := utils.ParsePlan(aiResponse)
	if err != nil {
		return &models.QueryResponse{
			Answer:  "I understood your request but couldn't create a properly formatted event. Please try again.",
			Success: false,
			Action:  "create",
			Error:   err.Error(),
		}
	}

	if len(tasks) == 0 {
		return &models.QueryResponse{
			Answer:  "I couldn't identify any events to create from your request.",
			Success: false,
			Action:  "create",
		}
	}

	// Create a summary of what will be created
	var eventNames []string
	for _, task := range tasks {
		eventNames = append(eventNames, task.Summary)
	}

	answer := fmt.Sprintf("I'll create %d event(s): %s", len(tasks), strings.Join(eventNames, ", "))

	return &models.QueryResponse{
		Answer:  answer,
		Success: true,
		Action:  "create",
		Events:  tasks,
	}
}
func (q *QueryProcessor) createSchedulingPrompt(question string, context models.QueryContext) string {
	currentTime := context.CurrentTime.Format("2006-01-02 15:04:05")
	
	prompt := fmt.Sprintf(`
You are a calendar assistant. The user wants to schedule an event based on their request: "%s"

Current time: %s
Current timezone: Asia/Kolkata

Please extract the event details and respond with a JSON array containing the events to create.
Each event should have:
- summary: The event title
- start: The start time in RFC3339 format with timezone
- end: The end time in RFC3339 format with timezone (default 1 hour duration if not specified)
- location: The location (if mentioned, otherwise empty)
- description: Additional details (if any, otherwise empty)

Example response:
[
  {
    "summary": "Team Meeting",
    "start": "2025-01-15T14:00:00+05:30",
    "end": "2025-01-15T15:00:00+05:30",
    "location": "",
    "description": ""
  }
]

Important:
- Use Asia/Kolkata timezone (+05:30)
- If no specific time is mentioned, use reasonable defaults
- If duration is not specified, default to 1 hour
- Only respond with the JSON array, no additional text

User request: "%s"
`, question, currentTime, question)

	return prompt
}
