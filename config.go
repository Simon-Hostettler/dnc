package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	CharacterDir string `json:"character_dir"`
	DatabasePath string `json:"database_path"`
}

func DefaultConfig() Config {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return Config{
		CharacterDir: filepath.Join(configDir, "dnc", "characters"),
		DatabasePath: filepath.Join(configDir, "dnc", "dnc.db"),
	}
}

func LoadConfig() (Config, error) {
	configDir, err := os.UserConfigDir()
	def := DefaultConfig()
	if err != nil {
		return def, nil
	}
	configPath := filepath.Join(configDir, "dnc", "config.json")
	f, err := os.Open(configPath)
	if err != nil {
		return def, nil
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return def, err
	}
	if cfg.CharacterDir == "" {
		cfg.CharacterDir = def.CharacterDir
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = def.DatabasePath
	}
	return cfg, nil
}
