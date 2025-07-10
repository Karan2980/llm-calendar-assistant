package models

// Config represents the application configuration
type Config struct {
	AI       AIConfig       `json:"ai"`
	Calendar CalendarConfig `json:"calendar"`
	Google   GoogleConfig   `json:"google"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	GitHubToken      string `json:"github_token"`
	OpenAIKey        string `json:"openai_key"`
	HuggingFaceToken string `json:"huggingface_token"`
}

// CalendarConfig holds calendar configuration
type CalendarConfig struct {
	TimeZone string `json:"timezone"`
}

// GoogleConfig holds Google OAuth configuration
type GoogleConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenExpiry  string `json:"token_expiry"`
}