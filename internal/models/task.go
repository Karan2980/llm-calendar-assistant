package models

import (
	"fmt"
	"time"
)

// Task represents a calendar task/event
type Task struct {
Summary     string `json:"summary"`
Start       string `json:"start"`
End         string `json:"end"`
EventID     string `json:"event_id,omitempty"`
Description string `json:"description,omitempty"`
Location    string `json:"location,omitempty"`
}

// ParsedTask represents a task with parsed time
type ParsedTask struct {
	Task
	StartTime time.Time
	EndTime   time.Time
}

// TaskRequest represents user input for tasks
type TaskRequest struct {
	UserInput string
	TimeZone  string
}

// ValidateTask checks if a task has required fields
func (t *Task) ValidateTask() error {
	if t.Summary == "" {
		return fmt.Errorf("task summary is required")
	}
	if t.Start == "" {
		return fmt.Errorf("task start time is required")
	}
	if t.End == "" {
		return fmt.Errorf("task end time is required")
	}
	return nil
}

// ParseTime converts string time to time.Time
func (t *Task) ParseTime() (*ParsedTask, error) {
	startTime, err := time.Parse(time.RFC3339, t.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %v", err)
	}

	endTime, err := time.Parse(time.RFC3339, t.End)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %v", err)
	}

	return &ParsedTask{
		Task:      *t,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// Duration returns the duration of the task
func (t *Task) Duration() (time.Duration, error) {
	startTime, err := time.Parse(time.RFC3339, t.Start)
	if err != nil {
		return 0, err
	}

	endTime, err := time.Parse(time.RFC3339, t.End)
	if err != nil {
		return 0, err
	}

	return endTime.Sub(startTime), nil
}

// IsToday checks if the task is scheduled for today
func (t *Task) IsToday() bool {
	startTime, err := time.Parse(time.RFC3339, t.Start)
	if err != nil {
		return false
	}

	now := time.Now()
	return startTime.Year() == now.Year() && 
		   startTime.Month() == now.Month() && 
		   startTime.Day() == now.Day()
}
