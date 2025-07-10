package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// ParsePlan parses JSON response from AI into tasks
func ParsePlan(planJSON string) ([]models.Task, error) {
	// Clean up the response - sometimes AI adds extra text
	planJSON = strings.TrimSpace(planJSON)
	
	// Remove markdown code blocks if present
	planJSON = strings.ReplaceAll(planJSON, "```json", "")
	planJSON = strings.ReplaceAll(planJSON, "```", "")
	planJSON = strings.TrimSpace(planJSON)
	
	// Try to extract JSON from the response using regex
	jsonRegex := regexp.MustCompile(`\[[\s\S]*\]`)
	matches := jsonRegex.FindString(planJSON)
	if matches != "" {
		planJSON = matches
	}
	
	var tasks []models.Task
	err := json.Unmarshal([]byte(planJSON), &tasks)
	if err != nil {
		// Try parsing as single object
		var singleTask models.Task
		err = json.Unmarshal([]byte(planJSON), &singleTask)
		if err == nil {
			tasks = []models.Task{singleTask}
			return tasks, nil
		}
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	
	// Filter out tasks that might be duplicates of existing ones
	var newTasks []models.Task
	for _, task := range tasks {
		if err := task.ValidateTask(); err != nil {
			fmt.Printf("⚠️ Skipping invalid task: %v\n", err)
			continue
		}
		newTasks = append(newTasks, task)
	}
	
	return newTasks, nil
}

// TasksToJSON converts tasks to JSON string
func TasksToJSON(tasks []models.Task) (string, error) {
	jsonBytes, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal tasks to JSON: %v", err)
	}
	return string(jsonBytes), nil
}

// PrettyPrintTasks prints tasks in a readable format
func PrettyPrintTasks(tasks []models.Task) {
	fmt.Println("📋 Parsed Tasks:")
	for i, task := range tasks {
		fmt.Printf("  %d. %s\n", i+1, task.Summary)
		fmt.Printf("     Time: %s to %s\n", FormatTime(task.Start), FormatTime(task.End))
		if task.Description != "" {
			fmt.Printf("     Description: %s\n", task.Description)
		}
		if task.Location != "" {
			fmt.Printf("     Location: %s\n", task.Location)
		}
		fmt.Println()
	}
}

// ValidateJSON checks if a string is valid JSON
func ValidateJSON(jsonStr string) error {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js)
}

// ExtractJSONFromText extracts JSON array from mixed text
func ExtractJSONFromText(text string) (string, error) {
	// Remove markdown code blocks first
	text = strings.ReplaceAll(text, "```json", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.TrimSpace(text)
	
	// Look for JSON array pattern
	arrayRegex := regexp.MustCompile(`\[[\s\S]*?\]`)
	matches := arrayRegex.FindAllString(text, -1)
	
	for _, match := range matches {
		if ValidateJSON(match) == nil {
			return match, nil
		}
	}
	
	// Look for JSON object pattern
	objectRegex := regexp.MustCompile(`\{[\s\S]*?\}`)
	matches = objectRegex.FindAllString(text, -1)
	
	for _, match := range matches {
		if ValidateJSON(match) == nil {
			return fmt.Sprintf("[%s]", match), nil
		}
	}
	
	return "", fmt.Errorf("no valid JSON found in text")
}

// FilterNewTasks filters out tasks that already exist
func FilterNewTasks(newTasks []models.Task, existingTasks []models.Task) []models.Task {
	var filtered []models.Task
	
	for _, newTask := range newTasks {
		isExisting := false
		for _, existing := range existingTasks {
			// Check if task with same summary and similar time already exists
			if strings.EqualFold(newTask.Summary, existing.Summary) {
				isExisting = true
				break
			}
		}
		if !isExisting {
			filtered = append(filtered, newTask)
		}
	}
	
	return filtered
}
