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
	outputLog := logger.New()
	resultTable := result.New()

	events, _ := events.New()

	app.GenerateResultTable(cfg, resultTable, outputLog, events)

}
