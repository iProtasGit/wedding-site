package config

import (
	"encoding/json"
	"os"
)

// AppConfig holds all configuration for the application.
type AppConfig struct {
	ServerPort      string `json:"serverPort"`
	SpreadsheetID   string `json:"spreadsheetId"`
	CredentialsFile string `json:"credentialsFile"`
}

// Load loads the configuration from a JSON file.
func Load(path string) (*AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg AppConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
