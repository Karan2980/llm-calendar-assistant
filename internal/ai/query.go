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

// QueryProcessor handles natural language queries about calendar
type QueryProcessor struct {
	aiManager      *Manager
	calendarClient *calendar.Client // Use actual calendar client type
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
func (q *QueryProcessor) processRuleBasedQuery(question string, context models.QueryContext) *models.QueryResponse {
	question = strings.ToLower(strings.TrimSpace(question))

	// Check for scheduling queries first
	if q.isSchedulingQuery(question) {
		return q.handleSchedulingQuery(question, context)
	}

	// Show all upcoming events
	if strings.Contains(question, "upcoming events") || strings.Contains(question, "show me all") {
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

	return nil // No rule-based match found
}

func (q *QueryProcessor) isSchedulingQuery(question string) bool {
	schedulingKeywords := []string{
		"schedule", "book", "add", "create", "set up", "plan",
		"meeting", "appointment", "event", "reminder",
	}
	
	timeKeywords := []string{
		"at", "on", "tomorrow", "today", "next week", "monday", "tuesday",
		"wednesday", "thursday", "friday", "saturday", "sunday",
		"am", "pm", "o'clock", ":",
	}
	
	hasSchedulingKeyword := false
	hasTimeKeyword := false
	
	for _, keyword := range schedulingKeywords {
		if strings.Contains(question, keyword) {
			hasSchedulingKeyword = true
			break
		}
	}
	
	for _, keyword := range timeKeywords {
		if strings.Contains(question, keyword) {
			hasTimeKeyword = true
			break
		}
	}
	
	return hasSchedulingKeyword && hasTimeKeyword
}

func (q *QueryProcessor) handleSchedulingQuery(question string, context models.QueryContext) *models.QueryResponse {
	// Use AI to parse the scheduling request
	if !q.aiManager.HasClients() {
		return &models.QueryResponse{
			Answer:  "I can't schedule events right now because AI service is not available. Please use the /api/schedule endpoint instead.",
			Success: false,
			Error:   "No AI clients configured",
		}
	}

	prompt := q.createSchedulingPrompt(question, context)
	aiResponse, err := q.aiManager.GeneratePlan(prompt)
	if err != nil {
		return &models.QueryResponse{
			Answer:  "Sorry, I couldn't understand your scheduling request. Please try using the /api/schedule endpoint with a clear description.",
			Success: false,
			Error:   err.Error(),
		}
	}

	// Parse the AI response to extract event details
	cleanResponse := q.cleanAIResponse(aiResponse)

	// Try to parse as JSON first
	var schedulingResponse struct {
		Action  string        `json:"action"`
		Events  []models.Task `json:"events"`
		Message string        `json:"message"`
	}

	if err := json.Unmarshal([]byte(cleanResponse), &schedulingResponse); err != nil {
		return &models.QueryResponse{
			Answer:  fmt.Sprintf("I understood you want to schedule something, but I need more specific details. Please try: 'Schedule a meeting with John tomorrow at 2 PM for 1 hour'"),
			Success: false,
			Error:   "Could not parse scheduling request",
		}
	}

	// Handle scheduling
	if schedulingResponse.Action == "schedule" && len(schedulingResponse.Events) > 0 {
		return &models.QueryResponse{
			Answer:  fmt.Sprintf("I've prepared your scheduling request: %s.", schedulingResponse.Message),
			Events:  schedulingResponse.Events,
			Success: true,
		}
	}

	   // Handle deletion
	   if schedulingResponse.Action == "delete" && len(schedulingResponse.Events) > 0 {
			   var deleted []string
			   var failed []string
			   for _, e := range schedulingResponse.Events {
					   if e.EventID != "" {
							   err := q.calendarClient.DeleteEvent(e.EventID)
							   if err != nil {
									   failed = append(failed, e.Summary)
							   } else {
									   deleted = append(deleted, e.Summary)
							   }
					   } else {
							   // fallback to summary/time if no event ID
							   err := q.calendarClient.DeleteEventBySummaryAndTime(e.Summary, e.Start, e.End)
							   if err != nil {
									   failed = append(failed, e.Summary)
							   } else {
									   deleted = append(deleted, e.Summary)
							   }
					   }
			   }
			   answer := ""
			   if len(deleted) > 0 {
					   answer += fmt.Sprintf("Deleted events: %s. ", strings.Join(deleted, ", "))
			   }
			   if len(failed) > 0 {
					   answer += fmt.Sprintf("Failed to delete: %s.", strings.Join(failed, ", "))
			   }
			   return &models.QueryResponse{
					   Answer:  strings.TrimSpace(answer),
					   Success: len(failed) == 0,
					   Events:  schedulingResponse.Events,
			   }
	   }

	return &models.QueryResponse{
		Answer:  "I couldn't extract clear scheduling or deletion details from your request. Please be more specific about the time, date, and event details.",
		Success: false,
	}
}

func (q *QueryProcessor) createSchedulingPrompt(question string, context models.QueryContext) string {
	now := context.CurrentTime
	today := now.Format("2006-01-02")
	
	prompt := fmt.Sprintf(`You are a scheduling assistant. Parse the user's scheduling or deletion request and extract event details.

Current time: %s
Today's date: %s
Timezone: %s

USER REQUEST: %s

EXISTING EVENTS TODAY:
`, now.Format("2006-01-02 15:04"), today, context.TimeZone, question)

	if len(context.TodaysEvents) == 0 {
		prompt += "No events scheduled for today.\n"
	} else {
		for _, event := range context.TodaysEvents {
			prompt += fmt.Sprintf("- %s from %s to %s\n", 
				event.Summary, 
				utils.FormatTime(event.Start), 
				utils.FormatTime(event.End))
		}
	}

	prompt += fmt.Sprintf(`
Please analyze the request and respond with ONLY a JSON object in one of these formats:

For scheduling:
{
  "action": "schedule",
  "message": "Brief description of what will be scheduled",
  "events": [
	{
	  "summary": "Event Title",
	  "start": "%sT14:00:00+05:30",
	  "end": "%sT15:00:00+05:30"
	}
  ]
}

For deletion (if the user provides an event ID, include it in the event object as 'event_id'):
{
  "action": "delete",
  "message": "Brief description of what will be deleted",
  "events": [
	{
	  "summary": "Event Title",
	  "start": "%sT14:00:00+05:30",
	  "end": "%sT15:00:00+05:30",
	  "event_id": "the-event-id-if-provided"
	}
  ]
}

Rules:
1. Use ISO 8601 format with +05:30 timezone
2. If no specific time is mentioned, suggest a reasonable time
3. If no duration is mentioned, default to 1 hour
4. If date is relative (tomorrow, next week), calculate the actual date
5. Make sure the event doesn't conflict with existing events
6. If the request is unclear, set action to "clarify" instead
7. If the user provides an event ID for deletion, always include it as 'event_id' in the event object.

Respond with ONLY the JSON object, no markdown code blocks.`, today, today, today, today)

	return prompt
}

// processAIQuery uses AI to process complex queries
func (q *QueryProcessor) processAIQuery(question string, context models.QueryContext) (*models.QueryResponse, error) {
	if !q.aiManager.HasClients() {
		return &models.QueryResponse{
			Answer:  "AI service not available. Please try simpler queries like 'when is my next meeting?' or 'what time is gym?'",
			Success: false,
			Error:   "No AI clients configured",
		}, nil
	}

	prompt := q.createQueryPrompt(question, context)
	
	aiResponse, err := q.aiManager.GeneratePlan(prompt)
	if err != nil {
		return &models.QueryResponse{
			Answer:  "Sorry, I couldn't process your question. Please try asking about specific events like 'gym time' or 'next meeting'.",
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Clean up AI response - remove markdown code blocks
	cleanResponse := q.cleanAIResponse(aiResponse)

	// Try to parse AI response as JSON
	var response models.QueryResponse
	if err := json.Unmarshal([]byte(cleanResponse), &response); err != nil {
		// If JSON parsing fails, treat the response as plain text
		return &models.QueryResponse{
			Answer:  strings.TrimSpace(cleanResponse),
			Success: true,
		}, nil
	}

	return &response, nil
}

// cleanAIResponse removes markdown code blocks and cleans up AI response
func (q *QueryProcessor) cleanAIResponse(response string) string {
	response = strings.TrimSpace(response)
	
	// Remove markdown code blocks
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}
	
	return response
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

	// Filter events that are in the future
	now := context.CurrentTime
	var futureEvents []models.Task
	
	for _, event := range allEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		
		if eventTime.After(now) {
			futureEvents = append(futureEvents, event)
		}
	}

	if len(futureEvents) == 0 {
		return &models.QueryResponse{
			Answer:  "You have no upcoming events scheduled.",
			Success: true,
		}
	}

	   // Build response
	   answer := fmt.Sprintf("You have %d upcoming events:\n\n", len(futureEvents))
	   for i, event := range futureEvents {
			   eventTime, err := time.Parse(time.RFC3339, event.Start)
			   if err != nil {
					   continue
			   }
			   // Format the date and time nicely
			   var timeDesc string
			   if eventTime.Format("2006-01-02") == now.Format("2006-01-02") {
					   timeDesc = fmt.Sprintf("Today at %s", utils.FormatTime(event.Start))
			   } else {
					   timeDesc = fmt.Sprintf("%s at %s", eventTime.Format("Mon Jan 2"), utils.FormatTime(event.Start))
			   }
			   // Show event ID if present
			   eventIDStr := ""
			   if event.EventID != "" {
					   eventIDStr = fmt.Sprintf(" [ID: %s]", event.EventID)
			   }
			   answer += fmt.Sprintf("%d. %s - %s%s", i+1, event.Summary, timeDesc, eventIDStr)
			   if event.Location != "" {
					   answer += fmt.Sprintf(" (%s)", event.Location)
			   }
			   answer += "\n"
	   }

	return &models.QueryResponse{
		Answer:  strings.TrimSpace(answer),
		Events:  futureEvents,
		Success: true,
	}
}

// createQueryPrompt creates a prompt for AI query processing
func (q *QueryProcessor) createQueryPrompt(question string, context models.QueryContext) string {
	prompt := fmt.Sprintf(`You are a personal calendar assistant. Answer the user's question about their calendar.

Current time: %s
Timezone: %s

TODAY'S EVENTS:
`, context.CurrentTime.Format("2006-01-02 15:04"), context.TimeZone)

	if len(context.TodaysEvents) == 0 {
		prompt += "No events scheduled for today.\n"
	} else {
		for _, event := range context.TodaysEvents {
			prompt += fmt.Sprintf("- %s from %s to %s", 
				event.Summary, 
				utils.FormatTime(event.Start), 
				utils.FormatTime(event.End))
			if event.Location != "" {
				prompt += fmt.Sprintf(" at %s", event.Location)
			}
			prompt += "\n"
		}
	}

	prompt += "\nUPCOMING EVENTS:\n"
	if len(context.UpcomingEvents) == 0 {
		prompt += "No upcoming events found.\n"
	} else {
		for i, event := range context.UpcomingEvents {
			if i >= 10 { // Limit to first 10 events
				break
			}
			prompt += fmt.Sprintf("- %s from %s to %s", 
				event.Summary, 
				utils.FormatTime(event.Start), 
				utils.FormatTime(event.End))
			if event.Location != "" {
				prompt += fmt.Sprintf(" at %s", event.Location)
			}
			prompt += "\n"
		}
	}

	prompt += fmt.Sprintf(`
USER QUESTION: %s

Please provide a helpful, conversational answer based on the calendar information above. 
Be specific about times and dates. If the information isn't available, say so politely.

Respond with ONLY a plain text answer (no JSON, no markdown code blocks).
Keep the response concise and user-friendly.`, question)

	return prompt
}

// Helper methods for rule-based queries

func (q *QueryProcessor) findNextEvent(context models.QueryContext) *models.QueryResponse {
	now := context.CurrentTime
	
	// Look through all events (today's and upcoming)
	allEvents := append(context.TodaysEvents, context.UpcomingEvents...)
	
	for _, event := range allEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		
		if eventTime.After(now) {
			timeUntil := eventTime.Sub(now)
			var timeDesc string
			
			if timeUntil < time.Hour {
				timeDesc = fmt.Sprintf("in %d minutes", int(timeUntil.Minutes()))
			} else if timeUntil < 24*time.Hour {
				timeDesc = fmt.Sprintf("in %d hours", int(timeUntil.Hours()))
			} else {
			timeDesc = fmt.Sprintf("on %s", eventTime.Format("Monday, January 2"))
		}
		
		answer := fmt.Sprintf("Your next event is '%s' %s at %s", 
			event.Summary, timeDesc, utils.FormatTime(event.Start))
		
		if event.Location != "" {
			answer += fmt.Sprintf(" at %s", event.Location)
		}
		
		return &models.QueryResponse{
			Answer:  answer,
			Events:  []models.Task{event},
			Success: true,
		}
	}
}

return &models.QueryResponse{
	Answer:  "You don't have any upcoming events scheduled.",
	Success: true,
}
}

func (q *QueryProcessor) findEventByKeyword(keyword string, context models.QueryContext) *models.QueryResponse {
var matchingEvents []models.Task

// Search in all events
allEvents := append(context.TodaysEvents, context.UpcomingEvents...)

for _, event := range allEvents {
	if strings.Contains(strings.ToLower(event.Summary), keyword) {
		matchingEvents = append(matchingEvents, event)
	}
}

if len(matchingEvents) == 0 {
	return &models.QueryResponse{
		Answer:  fmt.Sprintf("I couldn't find any %s events in your calendar.", keyword),
		Success: true,
	}
}

event := matchingEvents[0]
eventTime, _ := time.Parse(time.RFC3339, event.Start)

var answer string
if event.Start != "" {
	if eventTime.Format("2006-01-02") == time.Now().Format("2006-01-02") {
		answer = fmt.Sprintf("Your %s is today at %s", keyword, utils.FormatTime(event.Start))
	} else {
		answer = fmt.Sprintf("Your %s is on %s at %s", 
			keyword, eventTime.Format("Monday, January 2"), utils.FormatTime(event.Start))
	}
	
	if event.Location != "" {
		answer += fmt.Sprintf(" at %s", event.Location)
	}
} else {
	answer = fmt.Sprintf("Found %s event: %s", keyword, event.Summary)
}

return &models.QueryResponse{
	Answer:  answer,
	Events:  []models.Task{event},
	Success: true,
}
}

func (q *QueryProcessor) getTodaysSchedule(context models.QueryContext) *models.QueryResponse {
if len(context.TodaysEvents) == 0 {
	return &models.QueryResponse{
		Answer:  "You have no events scheduled for today.",
		Success: true,
	}
}

answer := fmt.Sprintf("You have %d events today:\n", len(context.TodaysEvents))
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

func (q *QueryProcessor) findFreeTime(context models.QueryContext) *models.QueryResponse {
now := context.CurrentTime

// Simple free time detection - find gaps between events
if len(context.TodaysEvents) == 0 {
	return &models.QueryResponse{
		Answer:  "You're free all day today!",
		Success: true,
	}
}

// Sort events by start time and find gaps
var freeSlots []string

// Check if free before first event
if len(context.TodaysEvents) > 0 {
	firstEventTime, err := time.Parse(time.RFC3339, context.TodaysEvents[0].Start)
	if err == nil && firstEventTime.After(now.Add(time.Hour)) {
		freeSlots = append(freeSlots, fmt.Sprintf("Free until %s", utils.FormatTime(context.TodaysEvents[0].Start)))
	}
}

// Check gaps between events
for i := 0; i < len(context.TodaysEvents)-1; i++ {
	currentEnd, err1 := time.Parse(time.RFC3339, context.TodaysEvents[i].End)
	nextStart, err2 := time.Parse(time.RFC3339, context.TodaysEvents[i+1].Start)
	
	if err1 == nil && err2 == nil {
		gap := nextStart.Sub(currentEnd)
		if gap > 30*time.Minute { // Only mention gaps longer than 30 minutes
			freeSlots = append(freeSlots, fmt.Sprintf("Free from %s to %s", 
				utils.FormatTime(context.TodaysEvents[i].End), 
				utils.FormatTime(context.TodaysEvents[i+1].Start)))
		}
	}
}

if len(freeSlots) == 0 {
	return &models.QueryResponse{
		Answer:  "Your schedule looks pretty packed today with no significant free time slots.",
		Success: true,
	}
}

answer := "Here are your free time slots:\n" + strings.Join(freeSlots, "\n")

return &models.QueryResponse{
	Answer:  answer,
	Success: true,
}
}
