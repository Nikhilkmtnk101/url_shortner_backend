package main

import (
	"github.com/nikhil/url-shortner-backend/config"
	"github.com/nikhil/url-shortner-backend/internal/app"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := app.NewApp(cfg)
	app.Run()
}
