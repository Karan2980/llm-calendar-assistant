package ai

import (
	"fmt"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
)

// Client interface for AI services
type Client interface {
	GeneratePlan(prompt string) (string, error)
	GetName() string
}

// Manager handles multiple AI clients
type Manager struct {
	clients []Client
}

// NewManager creates a new AI manager - GitHub Models only
func NewManager(config models.AIConfig) *Manager {
	var clients []Client
	
	// Only add GitHub Models if token is available
	if config.GitHubToken != "" && config.GitHubToken != "your_github_token_here" {
		clients = append(clients, NewGitHubClient(config.GitHubToken))
	}
	
	return &Manager{clients: clients}
}

// GeneratePlan tries AI clients in order
func (m *Manager) GeneratePlan(prompt string) (string, error) {
	if len(m.clients) == 0 {
		return "", fmt.Errorf("no GitHub token configured. Please add GITHUB_TOKEN to your .env file")
	}

	var lastError error
	for i, client := range m.clients {
		fmt.Printf("🤖 Trying %s (%d/%d)...\n", client.GetName(), i+1, len(m.clients))
		
		result, err := client.GeneratePlan(prompt)
		if err == nil {
			fmt.Printf("✅ Successfully used %s\n", client.GetName())
			return result, nil
		}
		
		fmt.Printf("⚠️ %s failed: %v\n", client.GetName(), err)
		lastError = err
	}
	
	return "", fmt.Errorf("GitHub Models failed: %v", lastError)
}

// GetAvailableClients returns list of available clients
func (m *Manager) GetAvailableClients() []string {
	var names []string
	for _, client := range m.clients {
		names = append(names, client.GetName())
	}
	return names
}

// HasClients returns true if any clients are configured
func (m *Manager) HasClients() bool {
	return len(m.clients) > 0
}

// GetClientCount returns the number of configured clients
func (m *Manager) GetClientCount() int {
	return len(m.clients)
}

// PrintAvailableClients prints the list of available AI clients
func (m *Manager) PrintAvailableClients() {
	if len(m.clients) == 0 {
		fmt.Println("❌ No GitHub token configured!")
		fmt.Println("💡 Please add your GitHub token to .env file:")
		fmt.Println("   GITHUB_TOKEN=your_actual_github_token")
		fmt.Println("")
		fmt.Println("🔗 Get your GitHub token from:")
		fmt.Println("   https://github.com/settings/tokens")
		fmt.Println("   Note: You may need to request access to GitHub Models")
		return
	}
	
	fmt.Printf("🤖 Available AI clients (%d):\n", len(m.clients))
	for i, client := range m.clients {
		fmt.Printf("   %d. %s\n", i+1, client.GetName())
	}
}
