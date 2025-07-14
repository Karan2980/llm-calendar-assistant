package planner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/ai"
	"github.com/Karan2980/llm-planner-golang-project/internal/calendar"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	calendarv3 "google.golang.org/api/calendar/v3"
)

// QueryHandler handles calendar queries
type QueryHandler struct {
	queryService    *calendar.QueryService
	queryProcessor  *ai.QueryProcessor
	timeZone        string
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(calendarService *calendarv3.Service, aiConfig models.AIConfig, timeZone string) *QueryHandler {
	calendarClient := calendar.NewClient(calendarService)
	queryService := calendar.NewQueryService(calendarClient)
	aiManager := ai.NewManager(aiConfig)
	queryProcessor := ai.NewQueryProcessor(aiManager, calendarClient)

	return &QueryHandler{
		queryService:   queryService,
		queryProcessor: queryProcessor,
		timeZone:       timeZone,
	}
}

// HandleQuery processes a user query about their calendar
func (qh *QueryHandler) HandleQuery(ctx context.Context, question string) (*models.QueryResponse, error) {
	// Get calendar context
	queryContext, err := qh.buildQueryContext()
	if err != nil {
		return &models.QueryResponse{
			Answer:  "Sorry, I couldn't access your calendar to answer that question.",
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Process the query
	response, err := qh.queryProcessor.ProcessQuery(question, *queryContext)
	if err != nil {
		return &models.QueryResponse{
			Answer:  "Sorry, I encountered an error while processing your question.",
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return response, nil
}

// buildQueryContext builds the context needed for query processing

// buildQueryContext builds the context needed for query processing
func (qh *QueryHandler) buildQueryContext() (*models.QueryContext, error) {
	now := time.Now()

	// Get today's events
	todaysEvents, err := qh.queryService.GetTodaysSchedule()
	if err != nil {
		todaysEvents = []models.Task{} // Continue with empty events
	}

	// Get upcoming events (next 7 days)
	upcomingEvents, err := qh.queryService.GetUpcomingEvents(7)
	if err != nil {
		upcomingEvents = []models.Task{} // Continue with empty events
	}

	// Filter out today's events from upcoming events to avoid duplicates
	var filteredUpcoming []models.Task
	today := now.Format("2006-01-02")
	
	for _, event := range upcomingEvents {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err != nil {
			continue
		}
		
		// Only include events that are not today
		if eventTime.Format("2006-01-02") != today {
			filteredUpcoming = append(filteredUpcoming, event)
		}
	}

	return &models.QueryContext{
		CurrentTime:    now,
		TodaysEvents:   todaysEvents,
		UpcomingEvents: filteredUpcoming,
		TimeZone:       qh.timeZone,
	}, nil
}

// RunInteractiveQuery runs an interactive query session with menu return option
func (qh *QueryHandler) RunInteractiveQuery(ctx context.Context) error {
	fmt.Println("\n🤖 Calendar Query Assistant")
	fmt.Println("Ask me questions about your calendar!")
	fmt.Println("Examples:")
	fmt.Println("  - When is my next meeting?")
	fmt.Println("  - What time is gym?")
	fmt.Println("  - What's my schedule today?")
	fmt.Println("  - When am I free?")
	fmt.Println("  - Schedule gym at 9am tomorrow")
	fmt.Println("  - Delete my gym session")
	fmt.Println("\nCommands:")
	fmt.Println("  - Type 'menu' to return to main menu")
	fmt.Println("  - Type 'quit' or 'exit' to stop the application")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("❓ Your question: ")
		
		// Use ReadString instead of Scanln to handle spaces
		question, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				// Handle Ctrl+C gracefully
				fmt.Println("\n🔙 Returning to main menu...")
				return nil
			}
			fmt.Printf("❌ Error reading input: %v\n", err)
			continue
		}
		
		question = strings.TrimSpace(question)
		
		// Handle special commands
		switch strings.ToLower(question) {
		case "quit", "exit":
			fmt.Println("👋 Goodbye!")
			return fmt.Errorf("user_exit") // Special error to indicate user wants to exit completely
		case "menu", "back", "main":
			fmt.Println("🔙 Returning to main menu...")
			return nil // Return to main menu
		case "":
			continue // Skip empty input
		}

		response, err := qh.HandleQuery(ctx, question)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		// Handle JSON response from AI
		if strings.HasPrefix(response.Answer, "```json") {
			// Try to extract clean answer from JSON response
			lines := strings.Split(response.Answer, "\n")
			for _, line := range lines {
				if strings.Contains(line, `"answer":`) {
					// Extract the answer value
					start := strings.Index(line, `"answer": "`) + 11
					end := strings.LastIndex(line, `"`)
					if start > 10 && end > start {
						cleanAnswer := line[start:end]
						fmt.Printf("💬 %s\n", cleanAnswer)
						break
					}
				}
			}
		} else {
			fmt.Printf("💬 %s\n", response.Answer)
		}
		
		// Show related events if any
		if len(response.Events) > 0 {
			fmt.Println("\n📅 Related events:")
			for i, event := range response.Events {
				fmt.Printf("   %d. %s", i+1, event.Summary)
				if event.Start != "" {
					eventTime, err := time.Parse(time.RFC3339, event.Start)
					if err == nil {
						fmt.Printf(" - %s", eventTime.Format("Mon Jan 2, 15:04"))
					}
				}
				if event.Location != "" {
					fmt.Printf(" at %s", event.Location)
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}
}

// HandleBatchQueries processes multiple queries at once
func (qh *QueryHandler) HandleBatchQueries(ctx context.Context, questions []string) ([]*models.QueryResponse, error) {
	var responses []*models.QueryResponse
	
	for _, question := range questions {
		response, err := qh.HandleQuery(ctx, question)
		if err != nil {
			response = &models.QueryResponse{
				Answer:  fmt.Sprintf("Error processing question: %v", err),
				Success: false,
				Error:   err.Error(),
			}
		}
		responses = append(responses, response)
	}
	
	return responses, nil
}

// GetQuickStats returns quick statistics about the calendar
func (qh *QueryHandler) GetQuickStats(ctx context.Context) (*models.QueryResponse, error) {
	queryContext, err := qh.buildQueryContext()
	if err != nil {
		return &models.QueryResponse{
			Answer:  "Could not get calendar statistics.",
			Success: false,
			Error:   err.Error(),
		}, err
	}

	todayCount := len(queryContext.TodaysEvents)
	upcomingCount := len(queryContext.UpcomingEvents)
	
	answer := fmt.Sprintf("📊 Calendar Stats:\n")
	answer += fmt.Sprintf("• Today: %d events\n", todayCount)
	answer += fmt.Sprintf("• Next 7 days: %d events\n", upcomingCount)
	
	// Find next event
	if upcomingCount > 0 {
		nextEvent := queryContext.UpcomingEvents[0]
		eventTime, err := time.Parse(time.RFC3339, nextEvent.Start)
		if err == nil {
			timeUntil := eventTime.Sub(queryContext.CurrentTime)
			if timeUntil > 0 {
				if timeUntil < 24*time.Hour {
					answer += fmt.Sprintf("• Next event: %s in %d hours\n", 
						nextEvent.Summary, int(timeUntil.Hours()))
				} else {
					answer += fmt.Sprintf("• Next event: %s on %s\n", 
						nextEvent.Summary, eventTime.Format("Mon Jan 2"))
				}
			}
		}
	}

	return &models.QueryResponse{
		Answer:  answer,
		Success: true,
	}, nil
}

// SearchCalendar searches for events matching a keyword
func (qh *QueryHandler) SearchCalendar(ctx context.Context, keyword string, days int) (*models.QueryResponse, error) {
	events, err := qh.queryService.SearchEvents(keyword, days)
	if err != nil {
		return &models.QueryResponse{
			Answer:  fmt.Sprintf("Error searching calendar: %v", err),
			Success: false,
			Error:   err.Error(),
		}, err
	}

	if len(events) == 0 {
		return &models.QueryResponse{
			Answer:  fmt.Sprintf("No events found matching '%s' in the next %d days.", keyword, days),
			Success: true,
		}, nil
	}

	answer := fmt.Sprintf("Found %d events matching '%s':\n", len(events), keyword)
	for i, event := range events {
		eventTime, err := time.Parse(time.RFC3339, event.Start)
		if err == nil {
			answer += fmt.Sprintf("%d. %s - %s", i+1, event.Summary, eventTime.Format("Mon Jan 2, 15:04"))
			if event.Location != "" {
				answer += fmt.Sprintf(" at %s", event.Location)
			}
			answer += "\n"
		}
	}

	return &models.QueryResponse{
		Answer:  answer,
		Events:  events,
		Success: true,
	}, nil
}
