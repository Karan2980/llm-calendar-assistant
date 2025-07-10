package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// GitHubClient handles GitHub Models API
type GitHubClient struct {
	apiKey string
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient(apiKey string) *GitHubClient {
	return &GitHubClient{
		apiKey: apiKey,
	}
}

// GetName returns the client name
func (g *GitHubClient) GetName() string {
	return "GitHub GPT-4o"
}

// GeneratePlan generates a plan using GitHub's GPT-4o model
func (g *GitHubClient) GeneratePlan(prompt string) (string, error) {
	return GetPlanFromGitHub(g.apiKey, prompt)
}

// GetPlanFromGitHub - Uses the correct GitHub Models endpoint from the C# sample
// GetPlanFromGitHub - Uses the correct GitHub Models endpoint from the C# sample
func GetPlanFromGitHub(apiKey, prompt string) (string, error) {
	// Correct GitHub Models endpoint from the C# sample
	url := "https://models.github.ai/inference/chat/completions"
	
	// Try different models that might be available
	models := []string{"gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"}
	
	for _, model := range models {
		// Remove these debug lines:
		// fmt.Printf("🔍 Trying GitHub Models with %s...\n", model)
		result, err := makeGitHubRequest(url, apiKey, prompt, model)
		if err == nil {
			// Remove this debug line:
			// fmt.Printf("✅ Success with model: %s\n", model)
			return result, nil
		}
		// Remove this debug line:
		// fmt.Printf("❌ Model %s failed: %v\n", model, err)
	}
	
	return "", fmt.Errorf("all GitHub Models failed")
}


func makeGitHubRequest(url, apiKey, prompt, model string) (string, error) {
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a helpful personal assistant that creates daily schedules. Always respond with valid JSON format containing an array of tasks with summary, start, and end fields in ISO 8601 format.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   4096,
		"temperature":  1.0,
		"stream":       false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Headers based on the C# sample
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "LLM-Planner-Go/1.0")
	
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub Models API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no message in choice")
	}
	
	text, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("no content in message")
	}

	return text, nil
}
