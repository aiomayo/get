package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"get/internal/repository"
)

type Config struct {
	DefaultEditor string                           `json:"default_editor"`
	Repositories  map[string]repository.Repository `json:"repositories"`
}

func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".get")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error creating .get directory: %v\n", err)
	}
	return filepath.Join(configDir, "config.json")
}

func Load() (*Config, error) {
	configPath := GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			DefaultEditor: "code",
			Repositories:  make(map[string]repository.Repository),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	if config.Repositories == nil {
		config.Repositories = make(map[string]repository.Repository)
	}

	return &config, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	return os.WriteFile(GetConfigPath(), data, 0644)
}

func (c *Config) UpdateRepositoryEditor(name, editor string) {
	if c.Repositories == nil {
		c.Repositories = make(map[string]repository.Repository)
	}

	repo := c.Repositories[name]
	repo.Editor = editor
	c.Repositories[name] = repo
}
