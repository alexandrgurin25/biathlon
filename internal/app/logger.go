package app

import (
	"biathlon/pkg/events"
	"log"
)

func WriteToOutputLog(log *log.Logger, events []events.Event) {
	for _, event := range events {
		switch event.EventID {
		case 1:
			log.Printf("%s The competitor(%d) registered", event.Time, event.CompetitorID)
		case 2:
			log.Printf("%s The start time for the competitor(%d) was set by a draw to %s", event.Time, event.CompetitorID, event.ExtraParams)
		case 3:
			log.Printf("%s The competitor(%d) is on the start line", event.Time, event.CompetitorID)
		case 4:
			log.Printf("%s The competitor(%d) has started", event.Time, event.CompetitorID)
		case 5:
			log.Printf("%s The competitor(%d) is on the firing range(%s)", event.Time, event.CompetitorID, event.ExtraParams)
		case 6:
			log.Printf("%s The target(%s) has been hit by competitor(%d)", event.Time, event.ExtraParams, event.CompetitorID)
		case 7:
			log.Printf("%s The competitor(%d) left the firing range", event.Time, event.CompetitorID)
		case 8:
			log.Printf("%s The competitor(%d) entered the penalty laps", event.Time, event.CompetitorID)
		case 9:
			log.Printf("%s The competitor(%d) left the penalty laps", event.Time, event.CompetitorID)
		case 10:
			log.Printf("%s The competitor(%d) ended the main lap", event.Time, event.CompetitorID)
		case 11:
			log.Printf("%s The competitor(%d) can`t continue: %s", event.Time, event.CompetitorID, event.ExtraParams)
		case 32:
			log.Printf("%s The competitor(%d) is disqualified", event.Time, event.CompetitorID)
		case 33:
			log.Printf("%s The competitor(%d) has finished", event.Time, event.CompetitorID)
		}
	}
}
