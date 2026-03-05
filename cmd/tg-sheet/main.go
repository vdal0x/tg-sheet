package main

import (
	"log"

	"github.com/vdal0x/tg-sheet/pkg/config"
	"github.com/vdal0x/tg-sheet/pkg/ui"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ui.NewTrayApp(cfg).Start()
}
