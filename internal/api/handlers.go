package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/pkg/utils"
)

// ScheduleRequest represents a scheduling request
type ScheduleRequest struct {
	UserInput string `json:"user_input"`
}

// ScheduleResponse represents a scheduling response
type ScheduleResponse struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	EventsAdded int           `json:"events_added"`
	Events      []models.Task `json:"events,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Keyword string `json:"keyword"`
	Days    int    `json:"days,omitempty"`
}

// DeleteRequest represents a delete request
type DeleteRequest struct {
	EventID   string `json:"event_id,omitempty"`
	Summary   string `json:"summary,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	DeleteAll bool   `json:"delete_all,omitempty"`
}

// DeleteResponse represents a delete response
type DeleteResponse struct {
	Success        bool          `json:"success"`
	Message        string        `json:"message"`
	EventsDeleted  int           `json:"events_deleted"`
	DeletedEvents  []models.Task `json:"deleted_events,omitempty"`
	Error          string        `json:"error,omitempty"`
}


func (s *Server) handleUnifiedQuery(w http.ResponseWriter, r *http.Request) {
	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Question == "" {
		s.writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	fmt.Printf("🔍 Processing unified query: %s\n", req.Question)

	// More precise intent detection
	if isExplicitSchedulingRequest(req.Question) {
		fmt.Printf("🎯 Detected explicit scheduling intent - routing to schedule logic\n")
		s.handleSchedulingFromQuery(w, req.Question)
		return
	}

	if isExplicitDeleteRequest(req.Question) {
		fmt.Printf("🎯 Detected explicit delete intent - routing to delete logic\n")
		s.handleDeleteFromQuery(w, req.Question)
		return
	}

	// Default to query/view logic
	fmt.Printf("🎯 Detected view intent - routing to query logic\n")
	s.handleViewFromQuery(w, req.Question)
}


// More precise scheduling detection
func isExplicitSchedulingRequest(question string) bool {
	question = strings.ToLower(strings.TrimSpace(question))
	
	// First check for VIEW keywords - if present, it's NOT a scheduling request
	viewKeywords := []string{
		"what's", "what is", "show me", "tell me", "display", "view", "see",
		"check", "look at", "find", "get", "list", "when is", "when are",
		"my schedule", "schedule for", "events for",
	}
	
	for _, keyword := range viewKeywords {
		if strings.Contains(question, keyword) {
			return false
		}
	}
	
	// Explicit scheduling words - UPDATED to include "create"
	explicitScheduleKeywords := []string{
		"schedule a", "schedule an", "schedule", 
		"book a", "book an", "book",
		"add a", "add an", "add",
		"create a", "create an", "create", // Added these
		"set up a", "set up an", "set up",
		"plan a", "plan an", "plan",
		"arrange a", "arrange an", "arrange",
		"organize a", "organize an", "organize",
	}
	
	for _, keyword := range explicitScheduleKeywords {
		if strings.Contains(question, keyword) {
			return true
		}
	}
	
	return false
}


func (s *Server) handleSchedulingFromQuery(w http.ResponseWriter, question string) {
	fmt.Printf("📅 Processing scheduling request: %s\n", question)

	// Get existing events
	existingTasks, err := s.scheduler.GetCalendarClient().GetTodaysEvents()
	if err != nil {
		existingTasks = []models.Task{}
	}

	// Generate plan with AI (same logic as handleSchedule)
	prompt := s.scheduler.GetPromptGenerator().CreateRestrictivePrompt(existingTasks, question)
	planJSON, err := s.scheduler.GetAIManager().GeneratePlan(prompt)
	if err != nil {
		fmt.Printf("❌ AI planning failed: %v\n", err)
		s.writeError(w, http.StatusInternalServerError, "AI service unavailable. Please try again later.")
		return
	}

	fmt.Printf("✅ AI Generated plan: %s\n", planJSON)

	// Parse AI response
	tasks, err := utils.ParsePlan(planJSON)
	if err != nil {
		fmt.Printf("❌ Failed to parse AI response: %v\n", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to understand your request. Please be more specific.")
		return
	}

	// Validate that the AI only created what was requested - INCREASED LIMIT
	if len(tasks) > 5 { // Increased from 3 to 5
		fmt.Printf("⚠️ AI generated too many events (%d), rejecting\n", len(tasks))
		s.writeError(w, http.StatusBadRequest, "Request seems too broad. Please be more specific about what you want to schedule.")
		return
	}

	// Filter out conflicting tasks
	var validTasks []models.Task
	for _, task := range tasks {
		if !s.scheduler.GetConflictChecker().HasTimeConflict(task, existingTasks) {
			validTasks = append(validTasks, task)
		}
	}

	if len(validTasks) == 0 {
		s.writeError(w, http.StatusBadRequest, "No valid events could be created. Please check for time conflicts.")
		return
	}

	fmt.Printf("📋 Creating %d valid tasks\n", len(validTasks))

	// Create events
	eventsAdded, err := s.scheduler.GetCalendarClient().CreateMultipleEvents(validTasks)
	if err != nil {
		fmt.Printf("❌ Failed to create events: %v\n", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to create events")
		return
	}

	fmt.Printf("✅ Successfully created %d events\n", eventsAdded)

	// Create a summary of what was created
	var eventNames []string
	for _, task := range validTasks {
		eventNames = append(eventNames, task.Summary)
	}

	response := map[string]interface{}{
		"answer":       fmt.Sprintf("I've successfully scheduled %d event(s): %s", eventsAdded, strings.Join(eventNames, ", ")),
		"events":       validTasks,
		"success":      true,
		"action":       "create", // Make sure this is "create"
		"events_added": eventsAdded,
	}

	s.writeJSON(w, http.StatusOK, response)
}



// handleDeleteFromQuery handles delete requests from natural language
func (s *Server) handleDeleteFromQuery(w http.ResponseWriter, question string) {
	fmt.Printf("🗑️ Processing delete request: %s\n", question)

	// Get all events to search through
	todaysEvents, err := s.scheduler.GetCalendarClient().GetTodaysEvents()
	if err != nil {
		todaysEvents = []models.Task{}
	}

	upcomingEvents, err := s.scheduler.GetQueryService().GetUpcomingEvents(7)
	if err != nil {
		upcomingEvents = []models.Task{}
	}

	allEvents := append(todaysEvents, upcomingEvents...)
	
	if len(allEvents) == 0 {
		response := map[string]interface{}{
			"answer":  "You have no events to delete.",
			"success": false,
			"action":  "delete",
		}
		s.writeJSON(w, http.StatusOK, response)
		return
	}

	// Find events to delete based on user's specific request
	eventsToDelete := s.findEventsToDelete(question, allEvents)
	
	if len(eventsToDelete) == 0 {
		response := map[string]interface{}{
			"answer":  "No matching events found to delete. Please be more specific.",
			"success": false,
			"action":  "delete",
		}
		s.writeJSON(w, http.StatusOK, response)
		return
	}

	// Confirm what will be deleted (don't delete more than 5 events without explicit "all")
	if len(eventsToDelete) > 5 && !strings.Contains(strings.ToLower(question), "all") {
		var eventNames []string
		for _, event := range eventsToDelete {
			eventNames = append(eventNames, event.Summary)
		}
		
		response := map[string]interface{}{
			"answer":  fmt.Sprintf("Found %d events to delete: %s. This seems like a lot. Please be more specific or use 'delete all' if you're sure.", len(eventsToDelete), strings.Join(eventNames, ", ")),
			"events":  eventsToDelete,
			"success": false,
			"action":  "delete",
		}
		s.writeJSON(w, http.StatusOK, response)
		return
	}

	// Perform the deletion
	calendarClient := s.scheduler.GetCalendarClient()
	deletedEvents := []models.Task{}
	failedDeletes := []models.Task{}

	fmt.Printf("🗑️ Attempting to delete %d events\n", len(eventsToDelete))

	for _, event := range eventsToDelete {
		var err error
		if event.EventID != "" {
			err = calendarClient.DeleteEvent(event.EventID)
		} else {
			err = calendarClient.DeleteEventBySummaryAndTime(event.Summary, event.Start, event.End)
		}
		
		if err == nil {
			deletedEvents = append(deletedEvents, event)
			fmt.Printf("✅ Deleted: %s\n", event.Summary)
		} else {
			failedDeletes = append(failedDeletes, event)
			fmt.Printf("❌ Failed to delete: %s - %v\n", event.Summary, err)
		}
	}
	
	// Prepare response
	var answer string
	success := len(deletedEvents) > 0
	
	if len(deletedEvents) > 0 {
		var eventNames []string
		for _, event := range deletedEvents {
			eventNames = append(eventNames, event.Summary)
		}
		answer = fmt.Sprintf("Successfully deleted %d event(s): %s", len(deletedEvents), strings.Join(eventNames, ", "))
		
		if len(failedDeletes) > 0 {
			answer += fmt.Sprintf(". Failed to delete %d event(s).", len(failedDeletes))
		}
	} else {
		answer = "Failed to delete any events. Please try again."
	}

	response := map[string]interface{}{
		"answer":         answer,
		"events":         deletedEvents,
		"success":        success,
		"action":         "delete",
		"events_deleted": len(deletedEvents),
	}

	s.writeJSON(w, http.StatusOK, response)
}


func (s *Server) findEventsToDelete(question string, allEvents []models.Task) []models.Task {
	question = strings.ToLower(strings.TrimSpace(question))
	var eventsToDelete []models.Task
	
	// Check for "all" keyword with specific scope
	if strings.Contains(question, "all") {
		if strings.Contains(question, "today") {
			// Delete all today's events
			today := time.Now().Format("2006-01-02")
			for _, event := range allEvents {
				eventTime, err := time.Parse(time.RFC3339, event.Start)
				if err == nil && eventTime.Format("2006-01-02") == today {
					eventsToDelete = append(eventsToDelete, event)
				}
			}
		} else if strings.Contains(question, "tomorrow") {
			// Delete all tomorrow's events
			tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			for _, event := range allEvents {
				eventTime, err := time.Parse(time.RFC3339, event.Start)
				if err == nil && eventTime.Format("2006-01-02") == tomorrow {
					eventsToDelete = append(eventsToDelete, event)
				}
			}
		} else {
			// Delete all events (be careful with this)
			eventsToDelete = allEvents
		}
		return eventsToDelete
	}

	// Check for specific date mentions
	if dateEvents := s.findEventsByDate(question, allEvents); len(dateEvents) > 0 {
		return dateEvents
	}

	// Check for specific event keywords
	keywords := s.extractKeywords(question)
	for _, keyword := range keywords {
		for _, event := range allEvents {
			if strings.Contains(strings.ToLower(event.Summary), keyword) {
				// Avoid duplicates
				found := false
				for _, existing := range eventsToDelete {
					if existing.EventID == event.EventID || 
					   (existing.Summary == event.Summary && existing.Start == event.Start) {
						found = true
						break
					}
				}
				if !found {
					eventsToDelete = append(eventsToDelete, event)
				}
			}
		}
	}

	return eventsToDelete
}


func (s *Server) findEventsByDate(question string, allEvents []models.Task) []models.Task {
	var eventsForDate []models.Task
	var targetDate time.Time
	now := time.Now()
	
	if strings.Contains(question, "today") {
		targetDate = now
	} else if strings.Contains(question, "tomorrow") {
		targetDate = now.AddDate(0, 0, 1)
	} else if strings.Contains(question, "yesterday") {
		targetDate = now.AddDate(0, 0, -1)
	} else {
		// Try to parse specific dates (you can extend this)
		return eventsForDate
	}
	
	dateStr := targetDate.Format("2006-01-02")
	for _, event := range allEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err == nil && eventTime.Format("2006-01-02") == dateStr {
			eventsForDate = append(eventsForDate, event)
		}
	}
	
	return eventsForDate
}

// extractKeywords extracts relevant keywords from the delete request
func (s *Server) extractKeywords(question string) []string {
	question = strings.ToLower(question)
	var keywords []string
	
	// Remove delete-related words to get the actual event keywords
	deleteWords := []string{"delete", "remove", "cancel", "clear", "erase", "drop", "get rid of", "eliminate", "destroy", "wipe", "purge", "my", "the", "a", "an", "all", "today", "tomorrow", "yesterday"}
	
	words := strings.Fields(question)
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 2 { // Ignore very short words
			isDeleteWord := false
			for _, deleteWord := range deleteWords {
				if word == deleteWord {
					isDeleteWord = true
					break
				}
			}
			if !isDeleteWord {
				keywords = append(keywords, word)
			}
		}
	}
	
	return keywords
}

// handleViewFromQuery handles view/query requests
func (s *Server) handleViewFromQuery(w http.ResponseWriter, question string) {
	fmt.Printf("👀 Processing view request: %s\n", question)

	// Process query - this should only return information, not create events
	response, err := s.scheduler.GetQueryHandler().HandleQuery(context.Background(), question)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Action = "view"
	s.writeJSON(w, http.StatusOK, response)
}

// isSchedulingRequest checks if the question is asking to schedule/create events
// func isSchedulingRequest(question string) bool {
// 	question = strings.ToLower(strings.TrimSpace(question))
// 	scheduleKeywords := []string{
// 		"schedule", "book", "add", "create", "set up", "plan",
// 		"arrange", "organize", "make", "put", "insert", "place",
// 		"meeting", "appointment", "event", "reminder", "session",
// 	}
	
// 	timeKeywords := []string{
// 		"at", "on", "tomorrow", "today", "next week", "monday", "tuesday",
// 		"wednesday", "thursday", "friday", "saturday", "sunday",
// 		"am", "pm", "o'clock", ":", "morning", "afternoon", "evening",
// 	}
	
// 	hasScheduleKeyword := false
// 	hasTimeKeyword := false
	
// 	for _, keyword := range scheduleKeywords {
// 		if strings.Contains(question, keyword) {
// 			hasScheduleKeyword = true
// 			break
// 		}
// 	}
	
// 	for _, keyword := range timeKeywords {
// 		if strings.Contains(question, keyword) {
// 			hasTimeKeyword = true
// 			break
// 		}
// 	}
	
// 	return hasScheduleKeyword && hasTimeKeyword
// }




// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "LLM Calendar Assistant API",
		"version": "1.0.0",
	})
}

// handleRoot handles root endpoint with API documentation
// handleRoot handles root endpoint with API documentation
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"service": "LLM Calendar Assistant API",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"POST /api/unified": map[string]interface{}{
				"description": "🚀 UNIFIED ENDPOINT - AI understands and performs create, view, and delete operations",
				"body": map[string]string{
					"question": "Natural language query (create/view/delete)",
				},
				"examples": map[string]interface{}{
					"create": "create a team meeting tomorrow at 2 PM",
					"view":   "What's my schedule today?",
					"delete": "Delete my gym session",
				},
			},
			"GET /health": map[string]interface{}{
				"description": "Health check endpoint",
			},
		},
	}

	s.writeJSON(w, http.StatusOK, docs)
}
// isExplicitDeleteRequest checks if the question is explicitly asking to delete events
func isExplicitDeleteRequest(question string) bool {
	question = strings.ToLower(strings.TrimSpace(question))
	
	explicitDeleteKeywords := []string{
		"delete", "remove", "cancel", "clear", "erase", "drop",
		"get rid of", "eliminate", "destroy", "wipe", "purge",
	}
	
	for _, keyword := range explicitDeleteKeywords {
		if strings.Contains(question, keyword) {
			return true
		}
	}
	return false
}
