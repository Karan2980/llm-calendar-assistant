package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// HuggingFaceClient implements the Client interface for HuggingFace
type HuggingFaceClient struct {
	apiKey string
}

// NewHuggingFaceClient creates a new HuggingFace client
func NewHuggingFaceClient(apiKey string) *HuggingFaceClient {
	return &HuggingFaceClient{apiKey: apiKey}
}

// GetName returns the client name
func (c *HuggingFaceClient) GetName() string {
	return "HuggingFace DialoGPT"
}

// GeneratePlan generates a plan using HuggingFace API
func (c *HuggingFaceClient) GeneratePlan(prompt string) (string, error) {
	url := "https://api-inference.huggingface.co/models/microsoft/DialoGPT-large"

	reqBody := map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"max_new_tokens":   300,
			"temperature":      0.7,
			"do_sample":        true,
			"return_full_text": false,
		},
		"options": map[string]interface{}{
			"wait_for_model": true,
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				return "", fmt.Errorf("API error: %s", errorMsg)
			}
		}
		return "", fmt.Errorf("failed to parse response: %s", string(body))
	}

	if len(result) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	generatedText, ok := result[0]["generated_text"].(string)
	if !ok {
		return "", fmt.Errorf("no generated_text in response: %+v", result[0])
	}

	return generatedText, nil
}
