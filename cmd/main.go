package main

import (
	"log"
	"os"

	"github.com/nktauserum/md-2320/internal/app"
	"github.com/nktauserum/md-2320/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	bot := app.NewApplication(cfg)

	if err := bot.Run(); err != nil {
		log.Printf("Error starting bot: %v\n", err)
		os.Exit(1)
	}
}
