package utils

import (
	"time"
)

// FormatTime formats a time string for display
func FormatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04")
}

// FormatDateTime formats a time string with date for display
func FormatDateTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("2006-01-02 15:04")
}

// ParseTimeString parses a time string to time.Time
func ParseTimeString(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// GetTodayString returns today's date as string
func GetTodayString() string {
	return time.Now().Format("2006-01-02")
}

// GetCurrentTimeString returns current time as string
func GetCurrentTimeString() string {
	return time.Now().Format("15:04")
}

// AddDuration adds duration to a time string
func AddDuration(timeStr string, duration time.Duration) (string, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", err
	}
	return t.Add(duration).Format(time.RFC3339), nil
}

// SubtractDuration subtracts duration from a time string
func SubtractDuration(timeStr string, duration time.Duration) (string, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", err
	}
	return t.Add(-duration).Format(time.RFC3339), nil
}

// GetTimeDifference returns the duration between two time strings
func GetTimeDifference(startTime, endTime string) (time.Duration, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return 0, err
	}
	
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return 0, err
	}
	
	return end.Sub(start), nil
}

// IsValidTimeRange checks if start time is before end time
func IsValidTimeRange(startTime, endTime string) bool {
	start, err1 := time.Parse(time.RFC3339, startTime)
	end, err2 := time.Parse(time.RFC3339, endTime)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return start.Before(end)
}

// CreateTimeString creates a time string for today with given hour and minute
func CreateTimeString(hour, minute int) string {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	return t.Format(time.RFC3339)
}

// GetStartOfDay returns the start of today as time string
func GetStartOfDay() string {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return startOfDay.Format(time.RFC3339)
}

// GetEndOfDay returns the end of today as time string
func GetEndOfDay() string {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return endOfDay.Format(time.RFC3339)
}
