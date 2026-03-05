package main

import (
	"flag"
	"log"

	"go.uber.org/zap"

	"github.com/vdal0x/tg-sheet/pkg/config"
	"github.com/vdal0x/tg-sheet/pkg/ui"
)

func main() {
	envPath := flag.String("env", ".env", "path to env file")
	flag.Parse()

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
