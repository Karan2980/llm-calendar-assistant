package planner

import (
	"strings"
	"time"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// ConflictChecker handles time conflict detection
type ConflictChecker struct{}

// NewConflictChecker creates a new conflict checker
func NewConflictChecker() *ConflictChecker {
	return &ConflictChecker{}
}

// HasTimeConflict checks if a new task conflicts with existing tasks
func (c *ConflictChecker) HasTimeConflict(newTask models.Task, existingTasks []models.Task) bool {
	newStart, err1 := time.Parse(time.RFC3339, newTask.Start)
	newEnd, err2 := time.Parse(time.RFC3339, newTask.End)
	
	if err1 != nil || err2 != nil {
		return false // If we can't parse, assume no conflict
	}
	
	for _, existing := range existingTasks {
		existingStart, err1 := time.Parse(time.RFC3339, existing.Start)
		existingEnd, err2 := time.Parse(time.RFC3339, existing.End)
		
		if err1 != nil || err2 != nil {
			continue
		}
		
		// Special handling for lunch breaks during work hours
		if c.isLunchBreakDuringWork(newTask, existing) {
			continue // Allow lunch breaks during work hours
		}
		
		// Check for overlap
		if c.timesOverlap(newStart, newEnd, existingStart, existingEnd) {
			return true
		}
	}
	
	return false
}

// isLunchBreakDuringWork checks if this is a lunch break during work hours
func (c *ConflictChecker) isLunchBreakDuringWork(newTask models.Task, existingTask models.Task) bool {
	// Check if new task is a lunch/break and existing is work
	newSummary := strings.ToLower(newTask.Summary)
	existingSummary := strings.ToLower(existingTask.Summary)
	
	isLunchBreak := strings.Contains(newSummary, "lunch") || 
					strings.Contains(newSummary, "break") ||
					strings.Contains(newSummary, "meal")
	
	isWorkEvent := strings.Contains(existingSummary, "work") ||
				   strings.Contains(existingSummary, "office") ||
				   strings.Contains(existingSummary, "job")
	
	if !isLunchBreak || !isWorkEvent {
		return false
	}
	
	// Check if lunch break is within work hours
	newStart, err1 := time.Parse(time.RFC3339, newTask.Start)
	newEnd, err2 := time.Parse(time.RFC3339, newTask.End)
	existingStart, err3 := time.Parse(time.RFC3339, existingTask.Start)
	existingEnd, err4 := time.Parse(time.RFC3339, existingTask.End)
	
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false
	}
	
	// Allow lunch break if it's within work hours and reasonable duration (< 2 hours)
	duration := newEnd.Sub(newStart)
	withinWorkHours := newStart.After(existingStart) && newEnd.Before(existingEnd)
	reasonableDuration := duration <= 2*time.Hour
	
	return withinWorkHours && reasonableDuration
}

// timesOverlap checks if two time ranges overlap
func (c *ConflictChecker) timesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

// FindConflicts returns all conflicting tasks
func (c *ConflictChecker) FindConflicts(newTasks []models.Task, existingTasks []models.Task) []models.Task {
	var conflicts []models.Task
	
	for _, newTask := range newTasks {
		if c.HasTimeConflict(newTask, existingTasks) {
			conflicts = append(conflicts, newTask)
		}
	}
	
	return conflicts
}

// ResolveConflicts attempts to resolve conflicts by adjusting times
func (c *ConflictChecker) ResolveConflicts(tasks []models.Task, existingTasks []models.Task) []models.Task {
	var resolvedTasks []models.Task
	
	for _, task := range tasks {
		if !c.HasTimeConflict(task, existingTasks) {
			resolvedTasks = append(resolvedTasks, task)
		} else {
			// Try to find a new time slot
			if adjustedTask, found := c.findAlternativeTime(task, existingTasks); found {
				resolvedTasks = append(resolvedTasks, adjustedTask)
			}
		}
	}
	
	return resolvedTasks
}

// findAlternativeTime tries to find an alternative time for a conflicting task
func (c *ConflictChecker) findAlternativeTime(task models.Task, existingTasks []models.Task) (models.Task, bool) {
	startTime, err := time.Parse(time.RFC3339, task.Start)
	if err != nil {
		return task, false
	}
	
	endTime, err := time.Parse(time.RFC3339, task.End)
	if err != nil {
		return task, false
	}
	
	duration := endTime.Sub(startTime)
	
	// Try different time slots throughout the day
	for hour := 6; hour <= 22; hour++ {
		newStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 
			hour, 0, 0, 0, startTime.Location())
		newEnd := newStart.Add(duration)
		
		adjustedTask := models.Task{
			Summary: task.Summary,
			Start:   newStart.Format(time.RFC3339),
			End:     newEnd.Format(time.RFC3339),
		}
		
		if !c.HasTimeConflict(adjustedTask, existingTasks) {
			return adjustedTask, true
		}
	}
	
	return task, false
}
