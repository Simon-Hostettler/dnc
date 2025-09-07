package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	CharacterDir string `json:"character_dir"`
}

func DefaultConfig() Config {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return Config{
		CharacterDir: filepath.Join(configDir, "dnc", "characters"),
	}
}

func LoadConfig() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return DefaultConfig(), nil
	}
	configPath := filepath.Join(configDir, "dnc", "config.json")
	f, err := os.Open(configPath)
	if err != nil {
		return DefaultConfig(), nil
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}
