package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

func New() (*Config, error) {
	var cfg Config

	cfgPath := filepath.Join("config", "config.json")
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("the file could not be opened, you may need to add \"config.json\" to the \"config\" folder -> %v", err)
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error occurred during unmarshal: %v", err)
	}

	return &cfg, nil
}
