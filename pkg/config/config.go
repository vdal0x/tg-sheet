package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Tg struct {
		ApiId   int    `yaml:"api_id"`
		ApiHash string `yaml:"api_hash"`
		Phone   string `yaml:"phone"`
	} `yaml:"telegram"`
}

func LoadConfig(fileName string) (*Config, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}
