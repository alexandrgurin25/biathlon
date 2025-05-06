package app

import (
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"testing"
)

func TestCalculateTimeDifference(t *testing.T) {
	cfg := &config.Config{
		StartDelta: "00:00:30",
	}

	tests := []struct {
		name        string
		wantStart   string
		actualStart string
		expected    string
		notStarted  bool
	}{
		{
			name:        "On time",
			wantStart:   "10:00:00.000",
			actualStart: "10:00:00.000",
			expected:    "00:00:00.000",
		},
		{
			name:        "Slight delay",
			wantStart:   "10:00:00.000",
			actualStart: "10:00:00.500",
			expected:    "00:00:00.500",
		},
		{
			name:        "Too late",
			wantStart:   "10:00:00.000",
			actualStart: "10:00:31.000",
			notStarted:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := competitor{
				WantStart:   tt.wantStart,
				ActualStart: tt.actualStart,
			}

			err := calculateTimeDifference(cfg, &c)
			if err != nil {
				t.Fatalf("calculateTimeDifference failed: %v", err)
			}

			if tt.notStarted && !c.NotStarted {
				t.Error("Expected NotStarted=true, got false")
			}

			if !tt.notStarted && c.TimeDifference != tt.expected {
				t.Errorf("Expected time difference %s, got %s", tt.expected, c.TimeDifference)
			}
		})
	}
}

func TestCalculateLapTime(t *testing.T) {
	cfg := &config.Config{
		LapLen: 1000, // 1km
	}

	tests := []struct {
		name     string
		start    string
		end      string
		expected string
		speed    string
	}{
		{
			name:     "1 minute lap",
			start:    "10:00:00.000",
			end:      "10:01:00.000",
			expected: "00:01:00.000",
			speed:    "16.667", // 1000m/60s
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := competitor{
				ActualStart: tt.start,
				LastLapTime: tt.start,
			}

			event := &events.Event{
				Time: tt.end,
			}

			err := calculateLapTime(cfg, &c, event)
			if err != nil {
				t.Fatalf("calculateLapTime failed: %v", err)
			}

			if len(c.LapsMain) != 1 {
				t.Fatalf("Expected 1 lap, got %d", len(c.LapsMain))
			}

			if c.LapsMain[0].Time != tt.expected {
				t.Errorf("Expected lap time %s, got %s", tt.expected, c.LapsMain[0].Time)
			}

			if c.LapsMain[0].Speed != tt.speed {
				t.Errorf("Expected speed %s, got %s", tt.speed, c.LapsMain[0].Speed)
			}
		})
	}
}
