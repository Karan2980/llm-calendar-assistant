package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// 	"time"
// )

// // GPT4oRequest represents the request structure for GPT-4o API
// type GPT4oRequest struct {
// 	Model       string    `json:"model"`
// 	Messages    []Message `json:"messages"`
// 	MaxTokens   int       `json:"max_tokens"`
// 	Temperature float64   `json:"temperature"`
// 	Stream      bool      `json:"stream"`
// }

// // Message represents a chat message
// type Message struct {
// 	Role    string `json:"role"`
// 	Content string `json:"content"`
// }

// // GPT4oResponse represents the response from GPT-4o API
// type GPT4oResponse struct {
// 	ID      string   `json:"id"`
// 	Object  string   `json:"object"`
// 	Created int64    `json:"created"`
// 	Model   string   `json:"model"`
// 	Choices []Choice `json:"choices"`
// 	Usage   Usage    `json:"usage"`
// }

// // Choice represents a response choice
// type Choice struct {
// 	Index        int     `json:"index"`
// 	Message      Message `json:"message"`
// 	FinishReason string  `json:"finish_reason"`
// }

// // Usage represents token usage information
// type Usage struct {
// 	PromptTokens     int `json:"prompt_tokens"`
// 	CompletionTokens int `json:"completion_tokens"`
// 	TotalTokens      int `json:"total_tokens"`
// }

// // GetPlanFromGPT4o calls the GPT-4o API from GitHub
// // GetPlanFromGPT4o calls the GPT-4o API from GitHub
// func GetPlanFromGPT4o(apiKey, prompt string) (string, error) {
// 	// Update this URL with your actual GitHub GPT-4o API endpoint
// 	// Common GitHub API patterns:
// 	url := "https://models.inference.ai.azure.com/chat/completions" // Azure OpenAI via GitHub
// 	// OR if it's a different endpoint format:
// 	// url := "https://api.github.com/models/gpt-4o/chat/completions"
// 	// OR if it's GitHub Copilot API:
// 	// url := "https://api.githubcopilot.com/chat/completions"

// 	// You might need to replace this with your specific endpoint
// 	// Please check your GitHub API documentation for the exact URL

// 	reqBody := GPT4oRequest{
// 		Model: "gpt-4o", // or "gpt-4o-mini" depending on your access
// 		Messages: []Message{
// 			{
// 				Role:    "system",
// 				Content: "You are a helpful personal assistant that creates daily schedules. Always respond with valid JSON format containing an array of tasks with summary, start, and end fields in ISO 8601 format.",
// 			},
// 			{
// 				Role:    "user",
// 				Content: prompt,
// 			},
// 		},
// 		MaxTokens:   500,
// 		Temperature: 0.7,
// 		Stream:      false,
// 	}

// 	bodyBytes, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal request: %v", err)
// 	}

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create request: %v", err)
// 	}

// 	// Headers for GitHub API - adjust based on your specific API
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+apiKey)
// 	req.Header.Set("User-Agent", "LLM-Planner-Go/1.0")

// 	// Additional headers that might be needed for GitHub APIs:
// 	// req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
// 	// req.Header.Set("Accept", "application/vnd.github+json")

// 	client := &http.Client{Timeout: 60 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("network error: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read response: %v", err)
// 	}

// 	// Enhanced error handling
// 	if resp.StatusCode != 200 {
// 		fmt.Printf("Response Status: %d\n", resp.StatusCode)
// 		fmt.Printf("Response Headers: %v\n", resp.Header)
// 		fmt.Printf("Response Body: %s\n", string(body))
// 		return "", fmt.Errorf("GPT-4o API returned status %d: %s", resp.StatusCode, string(body))
// 	}

// 	// Parse the response
// 	var result GPT4oResponse
// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		// Try to parse as error response
// 		var errorResp map[string]interface{}
// 		if json.Unmarshal(body, &errorResp) == nil {
// 			if errorMsg, ok := errorResp["error"].(string); ok {
// 				return "", fmt.Errorf("API error: %s", errorMsg)
// 			}
// 			if errorObj, ok := errorResp["error"].(map[string]interface{}); ok {
// 				if message, ok := errorObj["message"].(string); ok {
// 					return "", fmt.Errorf("API error: %s", message)
// 				}
// 			}
// 		}
// 		fmt.Printf("Raw response: %s\n", string(body))
// 		return "", fmt.Errorf("failed to parse response: %v", err)
// 	}

// 	if len(result.Choices) == 0 {
// 		return "", fmt.Errorf("no choices in response")
// 	}

// 	return result.Choices[0].Message.Content, nil
// }

// // GetPlanFromHuggingFace calls the Hugging Face Inference API (keeping as fallback)
// func GetPlanFromHuggingFace(apiKey, prompt string) (string, error) {
// 	url := "https://api-inference.huggingface.co/models/microsoft/DialoGPT-large"

// 	reqBody := map[string]interface{}{
// 		"inputs": prompt,
// 		"parameters": map[string]interface{}{
// 			"max_new_tokens":   300,
// 			"temperature":      0.7,
// 			"do_sample":        true,
// 			"return_full_text": false,
// 		},
// 		"options": map[string]interface{}{
// 			"wait_for_model": true,
// 		},
// 	}

// 	bodyBytes, _ := json.Marshal(reqBody)

// 	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+apiKey)

// 	client := &http.Client{Timeout: 60 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("network error: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, _ := ioutil.ReadAll(resp.Body)

// 	if resp.StatusCode != 200 {
// 		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result []map[string]interface{}
// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		var errorResp map[string]interface{}
// 		if json.Unmarshal(body, &errorResp) == nil {
// 			if errorMsg, ok := errorResp["error"].(string); ok {
// 				return "", fmt.Errorf("API error: %s", errorMsg)
// 			}
// 		}
// 		return "", fmt.Errorf("failed to parse response: %s", string(body))
// 	}

// 	if len(result) == 0 {
// 		return "", fmt.Errorf("empty response from API")
// 	}

// 	generatedText, ok := result[0]["generated_text"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("no generated_text in response: %+v", result[0])
// 	}

// 	return generatedText, nil
// }

// // GetPlanFromOpenAI - Alternative using OpenAI API (keeping as fallback)
// func GetPlanFromOpenAI(apiKey, prompt string) (string, error) {
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
// 	req.Header.Set("Authorization", "Bearer "+apiKey)

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

// // GetPlanFromAI - Main function that tries different APIs
// func GetPlanFromAI(apiKey, prompt string) (string, error) {
// 	// Try GPT-4o first (your new GitHub API)
// 	if result, err := GetPlanFromGPT4o(apiKey, prompt); err == nil {
// 		fmt.Println("✅ Used GPT-4o API from GitHub")
// 		return result, nil
// 	} else {
// 		fmt.Printf("⚠️ GPT-4o failed: %v\n", err)
// 	}

// 	// Fallback to OpenAI if available
// 	if strings.Contains(apiKey, "sk-") {
// 		if result, err := GetPlanFromOpenAI(apiKey, prompt); err == nil {
// 			fmt.Println("✅ Used OpenAI API as fallback")
// 			return result, nil
// 		} else {
// 			fmt.Printf("⚠️ OpenAI failed: %v\n", err)
// 		}
// 	}

// 	// Final fallback to Hugging Face
// 	if result, err := GetPlanFromHuggingFace(apiKey, prompt); err == nil {
// 		fmt.Println("✅ Used Hugging Face API as fallback")
// 		return result, nil
// 	} else {
// 		fmt.Printf("❌ All APIs failed. Last error: %v\n", err)
// 		return "", fmt.Errorf("all AI APIs failed")
// 	}
// }
