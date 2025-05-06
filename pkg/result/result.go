package result

import (
	"log"
	"os"
)

func New() *log.Logger {
	file, err := os.Create("Resulting table")
	if err != nil {
		log.Fatal("Failed to create Resulting table file:", err)
	}

	result := log.New(file, "", 0)

	return result
}
