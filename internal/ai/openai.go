package ai

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"time"
// )

// // OpenAIClient implements the Client interface for OpenAI
// type OpenAIClient struct {
// 	apiKey string
// }

// // NewOpenAIClient creates a new OpenAI client
// func NewOpenAIClient(apiKey string) *OpenAIClient {
// 	return &OpenAIClient{apiKey: apiKey}
// }

// // GetName returns the client name
// func (c *OpenAIClient) GetName() string {
// 	return "OpenAI GPT-3.5"
// }

// // GeneratePlan generates a plan using OpenAI API
// func (c *OpenAIClient) GeneratePlan(prompt string) (string, error) {
// 	url := "https://api.openai.com/v1/chat/completions"

// 	reqBody := map[string]interface{}{
// 		"model": "gpt-3.5-turbo",
// 		"messages": []map[string]interface{}{
// 			{
// 				"role":    "system",
// 				"content": "You are a helpful personal assistant that creates daily schedules. Always respond with valid JSON format containing an array of tasks with summary, start, and end fields in ISO 8601 format.",
// 			},
// 			{
// 				"role":    "user",
// 				"content": prompt,
// 			},
// 		},
// 		"max_tokens":  500,
// 		"temperature": 0.7,
// 	}

// 	bodyBytes, _ := json.Marshal(reqBody)

// 	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+c.apiKey)

// 	client := &http.Client{Timeout: 60 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("network error: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, _ := ioutil.ReadAll(resp.Body)

// 	if resp.StatusCode != 200 {
// 		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result map[string]interface{}
// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to parse OpenAI response: %v", err)
// 	}

// 	choices, ok := result["choices"].([]interface{})
// 	if !ok || len(choices) == 0 {
// 		return "", fmt.Errorf("no choices in OpenAI response")
// 	}

// 	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
// 	text := message["content"].(string)

// 	return text, nil
// }
