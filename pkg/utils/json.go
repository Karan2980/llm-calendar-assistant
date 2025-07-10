package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// ParsePlan parses AI-generated plan JSON into tasks
func ParsePlan(planJSON string) ([]models.Task, error) {
	// Clean up the JSON if it's wrapped in markdown code blocks
	planJSON = strings.TrimSpace(planJSON)
	if strings.HasPrefix(planJSON, "```json") {
		planJSON = strings.TrimPrefix(planJSON, "```json")
		planJSON = strings.TrimSuffix(planJSON, "```")
		planJSON = strings.TrimSpace(planJSON)
	} else if strings.HasPrefix(planJSON, "```") {
		planJSON = strings.TrimPrefix(planJSON, "```")
		planJSON = strings.TrimSuffix(planJSON, "```")
		planJSON = strings.TrimSpace(planJSON)
	}

	var tasks []models.Task
	if err := json.Unmarshal([]byte(planJSON), &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %v", err)
	}

	// Validate tasks
	for i, task := range tasks {
		if err := task.ValidateTask(); err != nil {
			return nil, fmt.Errorf("task %d validation failed: %v", i+1, err)
		}
	}

	return tasks, nil
}
