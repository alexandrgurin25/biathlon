package main

import (
	"biathlon/internal/app"
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"biathlon/pkg/logger"
	"biathlon/pkg/result"
)

func main() {
	cfg := config.New()
	log := logger.New()
	result := result.New()

	events, _ := events.New()

	app.WriteToOutputLog(log, events)
	app.GenerateResultTable(cfg, result, events)

}
