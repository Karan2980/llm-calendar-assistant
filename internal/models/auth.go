package models

// AuthConfig holds authentication configuration
type AuthConfig struct {
	ServiceAccountPath string `json:"service_account_path"`
	CredentialsPath    string `json:"credentials_path"`
	TokenPath          string `json:"token_path"`
}