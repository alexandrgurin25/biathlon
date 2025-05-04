package main

import (
	"biathlon/internal/app"
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"biathlon/pkg/logger"
)

func main() {
	_ = config.New()
	log := logger.New()

	events, _ := events.New()

	app.WriteToOutputLog(log, events)

}
