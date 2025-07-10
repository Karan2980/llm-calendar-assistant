package main

// import (
// 	"encoding/json"
// 	"regexp"
// 	"strings"
// )

// type Task struct {
// 	Summary string `json:"summary"`
// 	Start   string `json:"start"`
// 	End     string `json:"end"`
// }

// func ParsePlan(planJSON string) ([]Task, error) {
// 	// Clean up the response - sometimes AI adds extra text
// 	planJSON = strings.TrimSpace(planJSON)

// 	// Try to extract JSON from the response
// 	jsonRegex := regexp.MustCompile(`\[[\s\S]*\]`)
// 	matches := jsonRegex.FindString(planJSON)
// 	if matches != "" {
// 		planJSON = matches
// 	}

// 	var tasks []Task
// 	err := json.Unmarshal([]byte(planJSON), &tasks)
// 	if err != nil {
// 		// Try parsing as single object
// 		var singleTask Task
// 		err = json.Unmarshal([]byte(planJSON), &singleTask)
// 		if err == nil {
// 			tasks = []Task{singleTask}
// 			return tasks, nil
// 		}
// 	}
// 	return tasks, err
// }
