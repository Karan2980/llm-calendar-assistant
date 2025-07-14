package planner

import (
	"fmt"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/pkg/utils"
)

// PromptGenerator handles AI prompt generation
type PromptGenerator struct{}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator() *PromptGenerator {
	return &PromptGenerator{}
}

// CreatePlanningPrompt creates a comprehensive prompt for AI planning
func (p *PromptGenerator) CreatePlanningPrompt(existingTasks []models.Task, userInput string) string {
	now := time.Now()
	today := now.Format("2006-01-02")

	prompt := fmt.Sprintf(`You are a personal assistant helping to plan a daily schedule. 

Today's date: %s
Current time: %s

EXISTING CALENDAR EVENTS (DO NOT DUPLICATE THESE):
`, today, now.Format("15:04"))

	if len(existingTasks) == 0 {
		prompt += "No existing events found.\n"
	} else {
		for _, task := range existingTasks {
			prompt += fmt.Sprintf("- %s from %s to %s\n", 
				task.Summary, 
				utils.FormatTime(task.Start), 
				utils.FormatTime(task.End))
		}
	}

	prompt += fmt.Sprintf(`
NEW TASKS TO SCHEDULE: %s

IMPORTANT INSTRUCTIONS:
1. ONLY create events for the NEW TASKS mentioned above
2. DO NOT include existing events in your response
3. Schedule new tasks to fit around existing events without conflicts
4. Use reasonable time slots (lunch 30-60 min, meetings 1 hour, etc.)
5. Schedule tasks at appropriate times of day
6. Leave buffer time between events

Respond with ONLY a valid JSON array containing ONLY the NEW tasks:
[
  {
    "summary": "New Task Name",
    "start": "%sT12:30:00+05:30",
    "end": "%sT13:00:00+05:30"
  }
]

Use ISO 8601 format with +05:30 timezone. Make sure the JSON is valid and parseable.
REMEMBER: Only return NEW tasks, not existing ones!`, 
		userInput, today, today)

	return prompt
}


// CreateRestrictivePrompt creates a prompt that only creates what's specifically requested
// CreateRestrictivePrompt creates a prompt that only creates what's specifically requested
func (pg *PromptGenerator) CreateRestrictivePrompt(existingTasks []models.Task, userInput string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	
	existingEventsStr := ""
	if len(existingTasks) > 0 {
		existingEventsStr = "Existing events today:\n"
		for _, task := range existingTasks {
			existingEventsStr += fmt.Sprintf("- %s (%s to %s)\n", task.Summary, task.Start, task.End)
		}
	}

	prompt := fmt.Sprintf(`You are a calendar assistant. Create ONLY the specific event(s) mentioned in the user's request.

IMPORTANT RULES:
1. Create ONLY what the user explicitly asks for
2. Do NOT suggest additional events
3. Do NOT create a full day schedule
4. Use realistic durations (meetings: 1 hour, gym: 1 hour, lunch: 30 min, presentation: 1 hour)
5. If user specifies multiple events, create all of them

Current time: %s
%s

User request: "%s"

Respond with a JSON array of events. Each event should have:
- summary: Event title
- start: Start time in RFC3339 format (YYYY-MM-DDTHH:MM:SS+05:30)
- end: End time in RFC3339 format

Example response:
[
  {
    "summary": "Team Meeting",
    "start": "2025-07-15T14:00:00+05:30",
    "end": "2025-07-15T15:00:00+05:30"
  }
]

Only respond with the JSON array, no additional text.`, currentTime, existingEventsStr, userInput)

	return prompt
}


// CreateReschedulingPrompt creates a prompt for rescheduling conflicting tasks
func (p *PromptGenerator) CreateReschedulingPrompt(conflictingTasks []models.Task, existingTasks []models.Task) string {
	now := time.Now()
	today := now.Format("2006-01-02")

	prompt := fmt.Sprintf(`You need to reschedule the following conflicting tasks:

CONFLICTING TASKS:
`)

	for _, task := range conflictingTasks {
		prompt += fmt.Sprintf("- %s (%s to %s)\n", 
			task.Summary, 
			utils.FormatTime(task.Start), 
			utils.FormatTime(task.End))
	}

	prompt += "\nEXISTING EVENTS TO AVOID:\n"
	for _, task := range existingTasks {
		prompt += fmt.Sprintf("- %s (%s to %s)\n", 
			task.Summary, 
			utils.FormatTime(task.Start), 
			utils.FormatTime(task.End))
	}

	prompt += fmt.Sprintf(`
Please reschedule the conflicting tasks to avoid overlaps. Respond with ONLY a valid JSON array:
[
  {
    "summary": "Task name",
    "start": "%sT09:00:00+05:30",
    "end": "%sT10:00:00+05:30"
  }
]`, today, today)

	return prompt
}
