package logger

import (
	"log"
	"os"
)

func New() *log.Logger {
	file, err := os.OpenFile("Output log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	logger := log.New(file, "", 0)

	return logger
}
