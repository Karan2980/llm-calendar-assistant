package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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

// handleSchedule handles scheduling requests
func (s *Server) handleSchedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.UserInput == "" {
		s.writeError(w, http.StatusBadRequest, "user_input is required")
		return
	}

	// Get existing events
	existingTasks, err := s.scheduler.GetCalendarClient().GetTodaysEvents()
	if err != nil {
		existingTasks = []models.Task{}
	}

	// Generate plan with AI
	prompt := s.scheduler.GetPromptGenerator().CreatePlanningPrompt(existingTasks, req.UserInput)
	planJSON, err := s.scheduler.GetAIManager().GeneratePlan(prompt)
	if err != nil {
		// Use fallback plan
		tasks := s.scheduler.GetFallbackPlanner().CreateFallbackPlan(req.UserInput, existingTasks)
		eventsAdded, _ := s.scheduler.GetCalendarClient().CreateMultipleEvents(tasks)

		s.writeJSON(w, http.StatusOK, ScheduleResponse{
			Success:     true,
			Message:     "Events scheduled using fallback planner",
			EventsAdded: eventsAdded,
			Events:      tasks,
		})
		return
	}

	// Parse AI response
	tasks, err := utils.ParsePlan(planJSON)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	// Filter out conflicting tasks
	var validTasks []models.Task
	for _, task := range tasks {
		if !s.scheduler.GetConflictChecker().HasTimeConflict(task, existingTasks) {
			validTasks = append(validTasks, task)
		}
	}

	// Create events
	eventsAdded, err := s.scheduler.GetCalendarClient().CreateMultipleEvents(validTasks)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to create events")
		return
	}

	s.writeJSON(w, http.StatusOK, ScheduleResponse{
		Success:     true,
		Message:     "Events scheduled successfully",
		EventsAdded: eventsAdded,
		Events:      validTasks,
	})
}

// handleQuery handles query requests
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Question == "" {
		s.writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	response, err := s.scheduler.GetQueryHandler().HandleQuery(context.Background(), req.Question)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	calendarClient := s.scheduler.GetCalendarClient()
	createdEvents := []models.Task{}
	deletedEvents := []models.Task{}
	failedDeletes := []models.Task{}

	if len(response.Events) > 0 {
		// Check for scheduling or deletion intent in the answer (simple heuristic)
		if response.Answer != "" && (containsIgnoreCase(response.Answer, "delete") || containsIgnoreCase(response.Answer, "remove")) {
			// Try to delete events
			for _, event := range response.Events {
				err := calendarClient.DeleteEventBySummaryAndTime(event.Summary, event.Start, event.End)
				if err == nil {
					deletedEvents = append(deletedEvents, event)
				} else {
					failedDeletes = append(failedDeletes, event)
				}
			}
			if len(deletedEvents) > 0 {
				response.Answer += "\n🗑️ Event(s) deleted from Google Calendar."
				response.Success = true
				response.Events = deletedEvents
			} else {
				response.Answer += "\n❌ Failed to delete event(s) in Google Calendar."
				response.Success = false
			}
		} else {
			// Default: create events
			for _, event := range response.Events {
				err := calendarClient.CreateEvent(event)
				if err == nil {
					createdEvents = append(createdEvents, event)
				}
			}
			if len(createdEvents) > 0 {
				response.Answer += "\n✅ Event(s) created in Google Calendar."
				response.Success = true
				response.Events = createdEvents
			} else {
				response.Answer += "\n❌ Failed to create event(s) in Google Calendar."
				response.Success = false
			}
		}
	}

	s.writeJSON(w, http.StatusOK, response)

// containsIgnoreCase checks if substr is in s, case-insensitive
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (hasPrefixIgnoreCase(s, substr) || hasSuffixIgnoreCase(s, substr) || indexIgnoreCase(s, substr) >= 0))
}

func hasPrefixIgnoreCase(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return toLower(s[:len(prefix)]) == toLower(prefix)
}

func hasSuffixIgnoreCase(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return toLower(s[len(s)-len(suffix):]) == toLower(suffix)
}

func indexIgnoreCase(s, substr string) int {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// handleTodaysEvents handles today's events requests
func (s *Server) handleTodaysEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.scheduler.GetCalendarClient().GetTodaysEvents()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get today's events")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(events),
		"events":  events,
	})
}

// handleUpcomingEvents handles upcoming events requests
func (s *Server) handleUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7 // default

	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	events, err := s.scheduler.GetQueryService().GetUpcomingEvents(days)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get upcoming events")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"days":    days,
		"count":   len(events),
		"events":  events,
	})
}

// handleSearch handles search requests
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Keyword == "" {
		s.writeError(w, http.StatusBadRequest, "keyword is required")
		return
	}

	if req.Days <= 0 {
		req.Days = 7 // default
	}

	response, err := s.scheduler.GetQueryHandler().SearchCalendar(context.Background(), req.Keyword, req.Days)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleStats handles statistics requests
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	response, err := s.scheduler.GetQueryHandler().GetQuickStats(context.Background())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "LLM Calendar Assistant API",
		"version": "1.0.0",
	})
}

// handleRoot handles root endpoint with API documentation
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"service": "LLM Calendar Assistant API",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"POST /api/schedule": map[string]interface{}{
				"description": "Schedule new events",
				"body": map[string]string{
					"user_input": "Natural language description of events to schedule",
				},
				"example": map[string]string{
					"user_input": "1 hour gym, 30 min lunch break",
				},
			},
			"POST /api/query": map[string]interface{}{
				"description": "Ask questions about your calendar",
				"body": map[string]string{
					"question": "Natural language question about your calendar",
				},
				"example": map[string]string{
					"question": "When is my next meeting?",
				},
			},
			"GET /api/events/today": map[string]interface{}{
				"description": "Get today's events",
			},
			"GET /api/events/upcoming": map[string]interface{}{
				"description": "Get upcoming events",
				"query_params": map[string]string{
					"days": "Number of days to look ahead (default: 7)",
				},
			},
			"POST /api/search": map[string]interface{}{
				"description": "Search calendar events",
				"body": map[string]interface{}{
					"keyword": "Search keyword",
					"days":    "Number of days to search (optional, default: 7)",
				},
				"example": map[string]interface{}{
					"keyword": "gym",
					"days":    14,
				},
			},
			"GET /api/stats": map[string]interface{}{
				"description": "Get calendar statistics",
			},
			"GET /health": map[string]interface{}{
				"description": "Health check endpoint",
			},
		},
	}

	s.writeJSON(w, http.StatusOK, docs)
}
