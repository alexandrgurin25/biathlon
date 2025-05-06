package events

import (
	"fmt"
	"os"
	"path/filepath"
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
	eventsPath := filepath.Join("external_events", "events")
	file, err := os.ReadFile(eventsPath)
	if err != nil {
		return nil, fmt.Errorf(" \"events\" could not be opened, it may be necessary to add events to the \"external_events\" folder -> %v", err)
	}

	lines := strings.Split(string(file), "\n")

	var events []Event

	for _, line := range lines {
		if line == "" {
			continue
		}

		timeString := parseTimeEvent(line)

		// Разделяем Id события и Id участника на 2 части
		// + 2, потому что пропускаем последнюю единицу времени и "]"
		parts := strings.Fields(line[len(timeString)+2:])

		event := Event{
			Time:         timeString,
			EventID:      parseInt(parts[0]),
			CompetitorID: parseInt(parts[1]),
		}

		if event.EventID == 32 || event.EventID == 33 {
			return nil, fmt.Errorf("входной файл не может содержать исходящие события (ID 32 или 33)")
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
