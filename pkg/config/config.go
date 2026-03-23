package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Tg struct {
		AppId   int
		AppHash string
		Phone   string
	}
}

func LoadConfig(envFile string) (*Config, error) {
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("loading %s: %w", envFile, err)
	}

	appID, err := strconv.Atoi(os.Getenv("TG_APP_ID"))
	if err != nil {
		return nil, fmt.Errorf("invalid tg app value: %w", err)
	}

	appHash := os.Getenv("TG_APP_HASH")
	phone := os.Getenv("TG_PHONE")
	if appHash == "" || phone == "" {
		return nil, fmt.Errorf("TG_APP_HASH and TG_PHONE are required")
	}

	cfg := &Config{}
	cfg.Tg.AppId = appID
	cfg.Tg.AppHash = appHash
	cfg.Tg.Phone = phone
	return cfg, nil
}
