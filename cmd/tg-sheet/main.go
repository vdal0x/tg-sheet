package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/vdal0x/tg-sheet/pkg/config"
	"github.com/vdal0x/tg-sheet/pkg/ui"
)

func main() {
	envPath := flag.String("env", "", "path to .env file (default: auto-detect)")
	flag.Parse()

	if *envPath == "" {
		*envPath = resolveEnvPath()
	}

	cfg, err := config.LoadConfig(*envPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Sync()

	ui.NewTrayApp(cfg, logger).Start()
}

// resolveEnvPath returns the first .env path that exists, in priority order:
//  1. ~/Library/Application Support/tg-sheet/.env  (user-editable, survives app updates)
//  2. <executable>/../Resources/.env               (bundled inside .app)
//  3. .env in current working directory            (development)
func resolveEnvPath() string {
	// 1. User config dir
	if base, err := os.UserConfigDir(); err == nil {
		p := filepath.Join(base, "tg-sheet", ".env")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 2. Bundled Resources (inside .app)
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "..", "Resources", ".env")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 3. CWD fallback
	return ".env"
}
