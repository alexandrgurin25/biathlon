package app

import (
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"testing"
)

func TestCompetitorFlow(t *testing.T) {
	cfg := &config.Config{
		Laps:        2,
		StartDelta:  "00:00:30",
		FiringLines: 1,
	}

	tests := []struct {
		name     string
		events   []events.Event
		expected string
	}{
		{
			name: "Successful race",
			events: []events.Event{
				{Time: "10:00:00.000", EventID: 1, CompetitorID: 1},
				{Time: "10:00:01.000", EventID: 2, CompetitorID: 1, ExtraParams: "10:00:00.000"},
				{Time: "10:00:02.000", EventID: 3, CompetitorID: 1},
				{Time: "10:00:03.000", EventID: 4, CompetitorID: 1},
				{Time: "10:05:00.000", EventID: 10, CompetitorID: 1}, // Lap 1
				{Time: "10:10:00.000", EventID: 10, CompetitorID: 1}, // Lap 2 (finish)
			},
			expected: "00:10:00.000",
		},
		{
			name: "Not finished",
			events: []events.Event{
				{Time: "10:00:00.000", EventID: 1, CompetitorID: 1},
				{Time: "10:00:01.000", EventID: 2, CompetitorID: 1, ExtraParams: "10:00:00.000"},
				{Time: "10:00:02.000", EventID: 3, CompetitorID: 1},
				{Time: "10:00:03.000", EventID: 4, CompetitorID: 1},
				{Time: "10:05:00.000", EventID: 10, CompetitorID: 1}, // Lap 1
				{Time: "10:06:00.000", EventID: 11, CompetitorID: 1, ExtraParams: "Injury"},
			},
			expected: "NotFinished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			competitors := make(map[int]competitor)

			for _, event := range tt.events {
				cID := event.CompetitorID
				c := competitors[cID]

				switch event.EventID {
				case 1:
					c.Registered = true
					c.id = event.CompetitorID
				case 2:
					c.WantStart = event.ExtraParams
				case 4:
					c.ActualStart = event.Time
					calculateTimeDifference(cfg, &c)
				case 10:
					calculateLapTime(cfg, &c, &event)
					c.CompletedLaps++
					if isLastLap(cfg, &c) {
						c.FinishTime = event.Time
						calculateTotalTime(cfg, &c)
					}
				case 11:
					c.NotFinish = true
				}

				competitors[cID] = c
			}

			c := competitors[1]
			if tt.expected == "NotFinished" {
				if !c.NotFinish {
					t.Error("Expected NotFinish=true, got false")
				}
			} else if c.TotalRouteTime != tt.expected {
				t.Errorf("Expected total time %s, got %s", tt.expected, c.TotalRouteTime)
			}
		})
	}
}