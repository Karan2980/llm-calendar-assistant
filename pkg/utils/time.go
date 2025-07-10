package utils

import (
	"fmt"
	"strings"
	"time"
)

// FormatTime formats time string for display
func FormatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04")
}

// FormatDateTime formats datetime string for display
func FormatDateTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("Mon Jan 2, 15:04")
}

// ParseTimeInput parses various time input formats
func ParseTimeInput(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	
	// Try different formats
	formats := []string{
		"15:04",
		"3:04 PM",
		"3:04PM",
		"15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time: %s", input)
}
