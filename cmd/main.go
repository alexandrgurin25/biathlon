package main

import (
	"biathlon/internal/app"
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"biathlon/pkg/logger"
)

func main() {
	cfg := config.New()
	log := logger.New()

	events, _ := events.New()

	app.WriteToOutputLog(log, events)
	app.GenerateResultTable(cfg, log, events)

}
