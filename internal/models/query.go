package models

import "time"

// QueryRequest represents a user's calendar query
type QueryRequest struct {
	Question string `json:"question"`
	TimeZone string `json:"timezone"`
}

// QueryResponse represents the response to a calendar query
type QueryResponse struct {
	Answer  string `json:"answer"`
	Events  []Task `json:"events,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// QueryContext provides context for answering queries
type QueryContext struct {
	CurrentTime    time.Time `json:"current_time"`
	TodaysEvents   []Task    `json:"todays_events"`
	UpcomingEvents []Task    `json:"upcoming_events"`
	TimeZone       string    `json:"timezone"`
}

// TimeSlot represents a time slot
type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ConflictPair represents two conflicting events
type ConflictPair struct {
	Event1 Task `json:"event1"`
	Event2 Task `json:"event2"`
}
