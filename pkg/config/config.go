package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

func New() *Config {
	var cfg Config

	file, err := os.ReadFile("./config/config.json")
	if err != nil {
		log.Fatal("Error opening the file: ", err)
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatal("Error occurred during unmarshal: ", err)
	}

	return &cfg
}
