package events

import (
	"os"
	"strconv"
	"strings"
)

type Event struct {
	Time         string
	EventID      int
	CompetitorID int
	ExtraParams  string
}

func New() ([]Event, error) {
	file, err := os.ReadFile("external_events\\events")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(file), "\n")

	var events []Event

	for _, line := range lines {
		if line == "" {
			continue
		}

		timeString := parseTimeEvent(line)

		// Разделяем Id события и Id участника на 2 части
		parts := strings.Fields(line[len(timeString):])

		event := Event{
			Time:         timeString,
			EventID:      parseInt(parts[0]),
			CompetitorID: parseInt(parts[1]),
		}

		if len(parts) > 2 {
			event.ExtraParams = strings.Join(parts[2:], " ")
		}

		events = append(events, event)
	}

	return events, nil
}

func parseInt(s string) int {
	id, _ := strconv.Atoi(s)
	return id
}

func parseTimeEvent(line string) string {
	timeEnd := strings.Index(line, "]")
	timeStr := line[1:timeEnd]

	return timeStr
}
