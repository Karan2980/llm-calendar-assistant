package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// FallbackPlanner creates plans when AI services fail
type FallbackPlanner struct{}

// NewFallbackPlanner creates a new fallback planner
func NewFallbackPlanner() *FallbackPlanner {
	return &FallbackPlanner{}
}

// CreateFallbackPlan creates a rule-based plan when AI fails
func (f *FallbackPlanner) CreateFallbackPlan(userInput string, existingTasks []models.Task) []models.Task {
	now := time.Now()
	today := now.Format("2006-01-02")
	
	var tasks []models.Task
	input := strings.ToLower(userInput)
	
	// Smart scheduling based on user input keywords
	if strings.Contains(input, "gym") || strings.Contains(input, "workout") || strings.Contains(input, "exercise") {
		tasks = append(tasks, models.Task{
			Summary: "Gym Workout",
			Start:   fmt.Sprintf("%sT07:00:00+05:30", today),
			End:     fmt.Sprintf("%sT08:00:00+05:30", today),
		})
	}
	
	if strings.Contains(input, "work") || strings.Contains(input, "office") {
		tasks = append(tasks, models.Task{
			Summary: "Work",
			Start:   fmt.Sprintf("%sT09:00:00+05:30", today),
			End:     fmt.Sprintf("%sT17:00:00+05:30", today),
		})
	}
	
	if strings.Contains(input, "lunch") || strings.Contains(input, "eat") {
		tasks = append(tasks, models.Task{
			Summary: "Lunch Break",
			Start:   fmt.Sprintf("%sT12:30:00+05:30", today),
			End:     fmt.Sprintf("%sT13:00:00+05:30", today),
		})
	}
	
	if strings.Contains(input, "read") || strings.Contains(input, "book") {
		tasks = append(tasks, models.Task{
			Summary: "Reading Time",
			Start:   fmt.Sprintf("%sT19:00:00+05:30", today),
			End:     fmt.Sprintf("%sT20:00:00+05:30", today),
		})
	}
	
	if strings.Contains(input, "study") || strings.Contains(input, "learn") {
		tasks = append(tasks, models.Task{
			Summary: "Study Session",
			Start:   fmt.Sprintf("%sT14:00:00+05:30", today),
			End:     fmt.Sprintf("%sT16:00:00+05:30", today),
		})
	}

	if strings.Contains(input, "meeting") {
		tasks = append(tasks, models.Task{
			Summary: "Meeting",
			Start:   fmt.Sprintf("%sT10:00:00+05:30", today),
			End:     fmt.Sprintf("%sT11:00:00+05:30", today),
		})
	}

	if strings.Contains(input, "break") || strings.Contains(input, "rest") {
		tasks = append(tasks, models.Task{
			Summary: "Break Time",
			Start:   fmt.Sprintf("%sT15:00:00+05:30", today),
			End:     fmt.Sprintf("%sT15:30:00+05:30", today),
		})
	}
	
	// If no specific keywords found, create a generic task
	if len(tasks) == 0 {
		tasks = append(tasks, models.Task{
			Summary: fmt.Sprintf("Task: %s", userInput),
			Start:   fmt.Sprintf("%sT%02d:00:00+05:30", today, now.Hour()+1),
			End:     fmt.Sprintf("%sT%02d:00:00+05:30", today, now.Hour()+2),
		})
	}
	
	return tasks
}
