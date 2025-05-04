package output

import (
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"log"
)

func ProcessingLogger(config *config.Config, log *log.Logger, events []events.Event) {

	for _, event := range events {
		switch event.EventID {
		case 1:
			
		case 2:
		case 3:
		case 4:
		case 5:
		case 6:
		case 7:
		case 8:
		case 9:
		case 10:
		case 11:

		}

	}
}
