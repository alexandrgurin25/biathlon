package main

import (
	"biathlon/internal/app"
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"biathlon/pkg/logger"
	"biathlon/pkg/result"
	"log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Error reading config.json file %v", err)
	}
	outputLog := logger.New()
	resultTable := result.New()

	events, err := events.New()
	if err != nil {
		log.Fatalf("Error reading events file: %v", err)
	}
	app.GenerateResultTable(cfg, resultTable, outputLog, events)

}
