package logger

import (
	"log"
	"os"
)

func New() *log.Logger {
	file, err := os.Create("Output log")
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	logger := log.New(file, "", 0)

	return logger
}
