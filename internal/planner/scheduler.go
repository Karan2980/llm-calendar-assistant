package planner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Karan2980/llm-planner-golang-project/internal/ai"
	"github.com/Karan2980/llm-planner-golang-project/internal/calendar"
	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/pkg/utils"
	calendarv3 "google.golang.org/api/calendar/v3"
)

// Scheduler handles the main scheduling logic
type Scheduler struct {
	calendarClient  *calendar.Client
	aiManager       *ai.Manager
	fallbackPlanner *ai.FallbackPlanner
	conflictChecker *ConflictChecker
	promptGenerator *PromptGenerator
}

// NewScheduler creates a new scheduler
func NewScheduler(calendarService *calendarv3.Service, aiConfig models.AIConfig) *Scheduler {
	return &Scheduler{
		calendarClient:  calendar.NewClient(calendarService),
		aiManager:       ai.NewManager(aiConfig),
		fallbackPlanner: ai.NewFallbackPlanner(),
		conflictChecker: NewConflictChecker(),
		promptGenerator: NewPromptGenerator(),
	}
}

// Run executes the main scheduling workflow
func (s *Scheduler) Run(ctx context.Context) error {
	fmt.Println("🚀 Starting LLM Planner...")

	// 1. Debug calendar access
	if err := s.calendarClient.Debug(); err != nil {
		return fmt.Errorf("calendar debug failed: %v", err)
	}

	// 2. Get existing events
	existingTasks, err := s.calendarClient.GetTodaysEvents()
	if err != nil {
		fmt.Printf("⚠️ Warning: Could not read existing events: %v\n", err)
		existingTasks = []models.Task{}
	}

	fmt.Printf("📋 Found %d existing events today\n", len(existingTasks))
	for _, task := range existingTasks {
		fmt.Printf("  - %s (%s to %s)\n", task.Summary,
			utils.FormatTime(task.Start), utils.FormatTime(task.End))
	}

	// 3. Get user input
	userInput := s.getUserInput()

	// 4. Generate plan with AI
	prompt := s.promptGenerator.CreatePlanningPrompt(existingTasks, userInput)
	fmt.Println("🤖 Planning your day with AI...")
	
	planJSON, err := s.aiManager.GeneratePlan(prompt)
	if err != nil {
		fmt.Printf("⚠️ AI planning failed: %v\n", err)
		fmt.Println("🔄 Creating fallback plan...")
		return s.executeFallbackPlan(userInput, existingTasks)
	}

	fmt.Println("✅ AI Generated plan:\n", planJSON)

	// 5. Parse and execute plan
	tasks, err := utils.ParsePlan(planJSON)
	if err != nil {
		fmt.Printf("⚠️ Error parsing AI response: %v\n", err)
		fmt.Println("🔄 Creating fallback plan...")
		return s.executeFallbackPlan(userInput, existingTasks)
	}

	return s.executePlan(tasks, existingTasks)
}

// getUserInput gets user input for new tasks
func (s *Scheduler) getUserInput() string {
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

// executePlan executes the generated plan
func (s *Scheduler) executePlan(tasks []models.Task, existingTasks []models.Task) error {
	addedCount := 0
	
	for _, task := range tasks {
		// Check for conflicts using the conflict checker
		if s.conflictChecker.HasTimeConflict(task, existingTasks) {
			fmt.Printf("⚠️ Skipping '%s' - conflicts with existing event\n", task.Summary)
			continue
		}

		// Create the event
		if err := s.calendarClient.CreateEvent(task); err != nil {
			fmt.Printf("⚠️ Error creating event '%s': %v\n", task.Summary, err)
		} else {
			fmt.Printf("✅ Event created: %s (%s to %s)\n",
				task.Summary, utils.FormatTime(task.Start), utils.FormatTime(task.End))
			addedCount++
			
			// Add the created task to existing tasks to avoid future conflicts
			existingTasks = append(existingTasks, task)
		}
	}

	if addedCount > 0 {
		fmt.Printf("🎉 Successfully added %d new events to your calendar! 🎉\n", addedCount)
	} else {
		fmt.Println("ℹ️ No new events were added (conflicts or errors)")
	}

	return nil
}

// executeFallbackPlan executes the fallback plan when AI fails
func (s *Scheduler) executeFallbackPlan(userInput string, existingTasks []models.Task) error {
	tasks := s.fallbackPlanner.CreateFallbackPlan(userInput, existingTasks)
	return s.executePlan(tasks, existingTasks)
}
