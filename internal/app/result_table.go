package app

import (
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"fmt"
	"log"
	"time"
)

type competitor struct {
	NotStarted bool
	NotFinish  bool

	LapsMain     []lap
	LapsPenalty  []lap
	NumberOfHits int

	TimeDifference string
	WantStart      string
	ActualStart    string
}

type lap struct {
	Time  string
	Speed string
}

func GenerateResultTable(cfg *config.Config, log *log.Logger, events []events.Event) {
	competitors := make(map[int]competitor)

	for _, event := range events {

		cID := event.CompetitorID
		c := competitors[cID]

		switch event.EventID {
		case 2:
			c.WantStart = event.ExtraParams

		case 4:
			c.ActualStart = event.Time
		}

		if c.TimeDifference == "" &&
			c.ActualStart != "" &&
			c.WantStart != "" {

			if err := calculateTimeDifference(cfg, &c); err != nil {
				fmt.Println("Error calculating time difference:", err)
				return
			}
		}

		competitors[cID] = c
		fmt.Printf("%+v \n", c)
	}
}

func calculateTimeDifference(cfg *config.Config, c *competitor) error {
	actualStartTime, err := parseTime(c.ActualStart)
	if err != nil {
		return fmt.Errorf("actualStartTime error: %w", err)
	}

	wantedStartTime, err := parseTime(c.WantStart)
	if err != nil {
		return fmt.Errorf("desiredStartTime error: %w", err)
	}

	startDeltaTime, err := parseTime(cfg.StartDelta + ".000")
	if err != nil {
		return fmt.Errorf("startDeltaTime error: %w", err)
	}
	startDeltaTime = startDeltaTime.AddDate(1, 0, 0)

	timeDifference := actualStartTime.Sub(wantedStartTime)

	if startDeltaTime.Sub(time.Time{}.Add(timeDifference)) > 0 {
		c.TimeDifference = timeDifference.String()
	} else {
		c.NotStarted = true
	}

	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04:05.000", timeStr)
}
