package main

import (
	"log"
	"time"

	"github.com/nktauserum/md-2320/internal/app"
	"github.com/nktauserum/md-2320/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	bot := app.NewApplication(cfg)

	for i := range cfg.NUM_RETRIES {
		if err := bot.Run(); err != nil {
			log.Printf("Error starting bot, retrying... %d\n", i)
			time.Sleep(time.Duration(cfg.WAITING_TIME_MS) * time.Millisecond)
			continue
		}

		break
	}
}
