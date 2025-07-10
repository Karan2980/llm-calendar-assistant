package planner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Karan2980/llm-planner-golang-project/internal/ai"
	"github.com/Karan2980/llm-planner-golang-project/internal/calendar"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/pkg/utils"
	calendarv3 "google.golang.org/api/calendar/v3"
)

// EnhancedScheduler includes both scheduling and query capabilities
type EnhancedScheduler struct {
	*Scheduler
	queryHandler *QueryHandler
}




// GetCalendarClient returns the calendar client
func (es *EnhancedScheduler) GetCalendarClient() *calendar.Client {
	return es.calendarClient
}

// GetAIManager returns the AI manager
func (es *EnhancedScheduler) GetAIManager() *ai.Manager {
	return es.aiManager
}

// GetFallbackPlanner returns the fallback planner
func (es *EnhancedScheduler) GetFallbackPlanner() *ai.FallbackPlanner {
	return es.fallbackPlanner
}

// GetConflictChecker returns the conflict checker
func (es *EnhancedScheduler) GetConflictChecker() *ConflictChecker {
	return es.conflictChecker
}

// GetPromptGenerator returns the prompt generator
func (es *EnhancedScheduler) GetPromptGenerator() *PromptGenerator {
	return es.promptGenerator
}

// GetQueryHandler returns the query handler
func (es *EnhancedScheduler) GetQueryHandler() *QueryHandler {
	return es.queryHandler
}

// GetQueryService returns the query service
func (es *EnhancedScheduler) GetQueryService() *calendar.QueryService {
	return es.queryHandler.queryService
}

// NewEnhancedScheduler creates a new enhanced scheduler with query capabilities
func NewEnhancedScheduler(calendarService *calendarv3.Service, aiConfig models.AIConfig, timeZone string) *EnhancedScheduler {
	scheduler := NewScheduler(calendarService, aiConfig)
	queryHandler := NewQueryHandler(calendarService, aiConfig, timeZone)

	return &EnhancedScheduler{
		Scheduler:    scheduler,
		queryHandler: queryHandler,
	}
}

// Run executes the main application with both scheduling and query options
func (es *EnhancedScheduler) Run(ctx context.Context) error {
	fmt.Println("🚀 Starting Enhanced LLM Calendar Assistant...")

	// Debug calendar access
	if err := es.calendarClient.Debug(); err != nil {
		return fmt.Errorf("calendar debug failed: %v", err)
	}

	// Show main menu
	return es.showMainMenu(ctx)
}

// showMainMenu displays the main menu and handles user choices
// Update the showMainMenu method to handle the return from query mode:

func (es *EnhancedScheduler) showMainMenu(ctx context.Context) error {
	for {
		fmt.Println("\n🎯 What would you like to do?")
		fmt.Println("1. 📅 Schedule new events")
		fmt.Println("2. ❓ Ask questions about your calendar")
		fmt.Println("3. 🔍 Search calendar events")
		fmt.Println("4. 📊 Show calendar statistics")
		fmt.Println("5. 🚪 Exit")
		fmt.Print("\nEnter your choice (1-5): ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := es.HandleScheduling(ctx); err != nil {
				fmt.Printf("❌ Scheduling error: %v\n", err)
			}
		case "2":
			if err := es.RunInteractiveQuery(ctx); err != nil {
				// Check if user wants to exit completely
				if err.Error() == "user_exit" {
					fmt.Println("👋 Goodbye!")
					return nil
				}
				fmt.Printf("❌ Query error: %v\n", err)
			}
			// If no error or non-exit error, continue to main menu
		case "3":
			if err := es.HandleSearch(ctx); err != nil {
				fmt.Printf("❌ Search error: %v\n", err)
			}
		case "4":
			if err := es.HandleStats(ctx); err != nil {
				fmt.Printf("❌ Stats error: %v\n", err)
			}
		case "5":
			fmt.Println("👋 Goodbye!")
			return nil
		default:
			fmt.Println("❌ Invalid choice. Please enter 1-5.")
		}
	}
}


// HandleScheduling handles the scheduling workflow (EXPORTED)
func (es *EnhancedScheduler) HandleScheduling(ctx context.Context) error {
	fmt.Println("\n📅 SCHEDULING MODE")
	
	// Get existing events
	existingTasks, err := es.calendarClient.GetTodaysEvents()
	if err != nil {
		fmt.Printf("⚠️ Warning: Could not read existing events: %v\n", err)
		existingTasks = []models.Task{}
	}

	fmt.Printf("📋 Found %d existing events today\n", len(existingTasks))
	for _, task := range existingTasks {
		fmt.Printf("  - %s (%s to %s)\n", task.Summary,
			utils.FormatTime(task.Start), utils.FormatTime(task.End))
	}

	// Get user input
	userInput := es.getUserInput()

	// Generate plan with AI
	prompt := es.promptGenerator.CreatePlanningPrompt(existingTasks, userInput)
	fmt.Println("🤖 Planning your day with AI...")
	
	planJSON, err := es.aiManager.GeneratePlan(prompt)
	if err != nil {
		fmt.Printf("⚠️ AI planning failed: %v\n", err)
		fmt.Println("🔄 Creating fallback plan...")
		return es.executeFallbackPlan(userInput, existingTasks)
	}

	fmt.Println("✅ AI Generated plan:\n", planJSON)

	// Parse and execute plan
	tasks, err := utils.ParsePlan(planJSON)
	if err != nil {
		fmt.Printf("⚠️ Error parsing AI response: %v\n", err)
		fmt.Println("🔄 Creating fallback plan...")
		return es.executeFallbackPlan(userInput, existingTasks)
	}

	return es.executePlan(tasks, existingTasks)
}

// HandleSearch handles calendar search functionality (EXPORTED)
func (es *EnhancedScheduler) HandleSearch(ctx context.Context) error {
	fmt.Println("\n🔍 SEARCH MODE")
	fmt.Print("Enter search keyword: ")
	
	reader := bufio.NewReader(os.Stdin)
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)
	
	if keyword == "" {
		fmt.Println("❌ Please enter a search keyword.")
		return nil
	}

	fmt.Print("Search in how many days ahead? (default: 7): ")
	daysInput, _ := reader.ReadString('\n')
	daysInput = strings.TrimSpace(daysInput)
	
	days := 7 // default
	if daysInput != "" {
		if parsedDays, err := strconv.Atoi(daysInput); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	response, err := es.queryHandler.SearchCalendar(ctx, keyword, days)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", response.Answer)
	return nil
}

// HandleStats handles calendar statistics display (EXPORTED)
func (es *EnhancedScheduler) HandleStats(ctx context.Context) error {
	fmt.Println("\n📊 CALENDAR STATISTICS")
	
	response, err := es.queryHandler.GetQuickStats(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", response.Answer)
	return nil
}


// RunInteractiveQuery runs an interactive query session (EXPORTED)
func (es *EnhancedScheduler) RunInteractiveQuery(ctx context.Context) error {
	return es.queryHandler.RunInteractiveQuery(ctx)
}


// getUserInput gets user input for new tasks (inherited from base Scheduler)
func (es *EnhancedScheduler) getUserInput() string {
	fmt.Println("\n💬 What would you like to add to your schedule?")
	fmt.Println("Example: '1 hour gym, 9 to 5 work, 30 min lunch break'")
	fmt.Print("Enter your tasks: ")

	reader := bufio.NewReader(os.Stdin)
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)
	
	if userInput == "" {
		userInput = "1 hour gym, 9 to 5 work, 30 min lunch break" // Default for testing
		fmt.Printf("Using default input: %s\n", userInput)
	}

	return userInput
}

// RunQuickQuery runs a single query without the interactive menu
func (es *EnhancedScheduler) RunQuickQuery(ctx context.Context, question string) error {
	fmt.Printf("🚀 Quick Query Mode: %s\n", question)

	// Debug calendar access
	if err := es.calendarClient.Debug(); err != nil {
		return fmt.Errorf("calendar debug failed: %v", err)
	}

	response, err := es.queryHandler.HandleQuery(ctx, question)
	if err != nil {
		return err
	}

	fmt.Printf("\n💬 %s\n", response.Answer)
	
	// Show related events if any
	if len(response.Events) > 0 {
		fmt.Println("\n📅 Related events:")
		for i, event := range response.Events {
			fmt.Printf("   %d. %s", i+1, event.Summary)
			if event.Start != "" {
				fmt.Printf(" - %s", utils.FormatTime(event.Start))
			}
			if event.Location != "" {
				fmt.Printf(" at %s", event.Location)
			}
			fmt.Println()
		}
	}

	return nil
}

// RunBatchQueries processes multiple queries at once
func (es *EnhancedScheduler) RunBatchQueries(ctx context.Context, questions []string) error {
	fmt.Println("🚀 Batch Query Mode")

	// Debug calendar access
	if err := es.calendarClient.Debug(); err != nil {
		return fmt.Errorf("calendar debug failed: %v", err)
	}

	responses, err := es.queryHandler.HandleBatchQueries(ctx, questions)
	if err != nil {
		return err
	}

	for i, response := range responses {
		fmt.Printf("\n❓ Question %d: %s\n", i+1, questions[i])
		fmt.Printf("💬 Answer: %s\n", response.Answer)
		
		if len(response.Events) > 0 {
			fmt.Println("📅 Related events:")
			for j, event := range response.Events {
				fmt.Printf("   %d. %s", j+1, event.Summary)
				if event.Start != "" {
					fmt.Printf(" - %s", utils.FormatTime(event.Start))
				}
				if event.Location != "" {
					fmt.Printf(" at %s", event.Location)
				}
				fmt.Println()
			}
		}
		fmt.Println(strings.Repeat("-", 50))
	}

	return nil
}

// ShowHelp displays help information
func (es *EnhancedScheduler) ShowHelp() {
	fmt.Println("🤖 Enhanced LLM Calendar Assistant Help")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("📅 SCHEDULING:")
	fmt.Println("  - Add new events to your calendar")
	fmt.Println("  - AI-powered intelligent scheduling")
	fmt.Println("  - Conflict detection and resolution")
	fmt.Println()
	fmt.Println("❓ QUERIES:")
	fmt.Println("  Examples of questions you can ask:")
	fmt.Println("  • When is my next meeting?")
	fmt.Println("  • What time is gym?")
	fmt.Println("  • What's my schedule today?")
	fmt.Println("  • When am I free?")
	fmt.Println("  • Do I have any work meetings this week?")
	fmt.Println("  • What time is lunch?")
	fmt.Println()
	fmt.Println("🔍 SEARCH:")
	fmt.Println("  - Search for events by keyword")
	fmt.Println("  - Specify time range for search")
	fmt.Println()
	fmt.Println("📊 STATISTICS:")
	fmt.Println("  - View calendar statistics")
	fmt.Println("  - See upcoming events summary")
	fmt.Println()
	fmt.Println("🚀 USAGE MODES:")
	fmt.Println("  1. Interactive mode (default)")
	fmt.Println("  2. Quick query mode")
	fmt.Println("  3. Batch query mode")
	fmt.Println()
}


// GetAvailableCommands returns a list of available commands
func (es *EnhancedScheduler) GetAvailableCommands() []string {
	return []string{
		"schedule",     // Schedule new events
		"query",        // Ask questions
		"search",       // Search events
		"stats",        // Show statistics
		"help",         // Show help
		"exit",         // Exit application
	}
}

// ProcessCommand processes a single command
func (es *EnhancedScheduler) ProcessCommand(ctx context.Context, command string, args []string) error {
	switch strings.ToLower(command) {
	case "schedule":
		return es.HandleScheduling(ctx)
	case "query":
		if len(args) > 0 {
			question := strings.Join(args, " ")
			return es.RunQuickQuery(ctx, question)
		}
		return es.RunInteractiveQuery(ctx)
	case "search":
		return es.HandleSearch(ctx)
	case "stats":
		return es.HandleStats(ctx)
	case "help":
		es.ShowHelp()
		return nil
	case "exit":
		fmt.Println("👋 Goodbye!")
		return nil
	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		fmt.Println("Type 'help' to see available commands.")
		return nil
	}
}
