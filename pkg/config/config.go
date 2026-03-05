package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Tg struct {
		ApiId   int
		ApiHash string
		Phone   string
	}
}

func LoadConfig(envFile string) (*Config, error) {
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("loading %s: %w", envFile, err)
	}

	apiID, err := strconv.Atoi(os.Getenv("TG_API_ID"))
	if err != nil {
		return nil, fmt.Errorf("TG_API_ID must be an integer: %w", err)
	}

	apiHash := os.Getenv("TG_API_HASH")
	phone := os.Getenv("TG_PHONE")
	if apiHash == "" || phone == "" {
		return nil, fmt.Errorf("TG_API_HASH and TG_PHONE are required")
	}

	cfg := &Config{}
	cfg.Tg.ApiId = apiID
	cfg.Tg.ApiHash = apiHash
	cfg.Tg.Phone = phone
	return cfg, nil
}
