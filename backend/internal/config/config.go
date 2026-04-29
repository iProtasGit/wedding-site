package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port            string `json:"port"`
	SpreadsheetID   string `json:"spreadsheetId"`
	CredentialsFile string `json:"credentialsFile"`
	TgBotToken      string `json:"tgBotToken"`
	TgChatID        string `json:"tgChatId"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decoding json in %s: %w", path, err)
	}
	return &cfg, nil
}
