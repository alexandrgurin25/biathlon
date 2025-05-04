package main

import (
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"biathlon/pkg/logger"
)

func main() {
	_ = config.New()

	_ = logger.New()

	_, _ = events.New()

}
