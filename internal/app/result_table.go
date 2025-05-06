package app

import (
	"biathlon/pkg/config"
	"biathlon/pkg/events"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type competitor struct {
	Registered bool
	id         int

	NotStarted bool
	NotFinish  bool

	LapsMain     []lap
	LapsPenalty  []lap
	NumberOfHits int

	TimeDifference string
	WantStart      string
	ActualStart    string
	LastLapTime    string

	PenaltyLenStart string
	AvgPenaltyTime  string
	AvgPenaltySpeed string

	FinishTime    string
	CompletedLaps int

	TotalRouteTime string
}

type lap struct {
	Time  string
	Speed string
}

func GenerateResultTable(cfg *config.Config, result *log.Logger, log *log.Logger, incommingEvents []events.Event) {
	competitors := make(map[int]competitor)
	var outgoingEvents []events.Event

	for _, event := range incommingEvents {

		cID := event.CompetitorID
		c := competitors[cID]

		switch event.EventID {

		case 1:
			c.Registered = true
			c.id = event.CompetitorID

		// Время старта установлено жеребьевкой
		case 2:
			c.WantStart = event.ExtraParams

		// Участник стартовал
		case 4:
			c.ActualStart = event.Time
			// Проверяем, не опоздал ли участник на старт
			if err := calculateTimeDifference(cfg, &c); err != nil {
				fmt.Println("Error calculating time difference:", err)
			}
			if c.NotStarted {
				dqEvent := events.Event{
					Time:         c.ActualStart,
					EventID:      32,
					CompetitorID: cID,
				}
				outgoingEvents = append(outgoingEvents, dqEvent)
			}
		// Мишень поражена
		case 6:
			c.NumberOfHits++

		// Участник вышел на штрафные круги
		case 8:
			c.PenaltyLenStart = event.Time

		// Участник покинул штрафные круги
		case 9:
			if err := calculatePenaltyLapTime(cfg, &c, &event); err != nil {
				fmt.Println("Error calculating time penalty lap:", err)
			}

		// Участник закончил основной круг
		case 10:
			if err := calculateLapTime(cfg, &c, &event); err != nil {
				fmt.Println("Error calculating time main lap:", err)
			}

			c.CompletedLaps++
			if isLastLap(cfg, &c) {
				// Генерируем событие о финише
				finishEvent := events.Event{
					Time:         event.Time,
					EventID:      33,
					CompetitorID: cID,
				}

				c.FinishTime = event.Time

				if err := calculateTotalTime(cfg, &c); err != nil {
					fmt.Println("Error calculating Total time laps:", err)
				}
				outgoingEvents = append(outgoingEvents, finishEvent)
			}

		// Участник не может продолжать
		case 11:
			c.NotFinish = true
		}

		if c.TimeDifference == "" &&
			c.ActualStart != "" &&
			c.WantStart != "" {

			if err := calculateTimeDifference(cfg, &c); err != nil {
				fmt.Println("Error calculating time difference:", err)
			}
		}

		if len(c.LapsPenalty) == 0 {
			c.AvgPenaltyTime = "00:00:00.000"
			c.AvgPenaltySpeed = "0.000"
		}

		competitors[cID] = c
	}
	allEvents := append(incommingEvents, outgoingEvents...)

	sortEventsByTime(allEvents)

	WriteToOutputLog(log, allEvents)

	// Вывод таблицы результатов
	outputResultTable(cfg, competitors, result)
}

func sortResultByTotalTime(competitors map[int]competitor) []competitor {

	competitorsSlice := make([]competitor, 0)

	for _, c := range competitors {
		competitorsSlice = append(competitorsSlice, c)
	}

	sort.Slice(competitorsSlice, func(i, j int) bool {
		timeI, errI := parseTime(competitorsSlice[i].TotalRouteTime)
		timeJ, errJ := parseTime(competitorsSlice[j].TotalRouteTime)

		// В случае ошибок парсинга сохраняем исходный порядок
		if errI != nil || errJ != nil {
			return i < j
		}

		return timeI.Before(timeJ)
	})

	return competitorsSlice
}

func sortEventsByTime(events []events.Event) {
	sort.Slice(events, func(i, j int) bool {
		timeI, errI := parseTime(events[i].Time)
		timeJ, errJ := parseTime(events[j].Time)

		// В случае ошибок парсинга сохраняем исходный порядок
		if errI != nil || errJ != nil {
			return i < j
		}

		return timeI.Before(timeJ)
	})
}

func isLastLap(cfg *config.Config, c *competitor) bool {
	return c.CompletedLaps >= cfg.Laps
}

func outputResultTable(cfg *config.Config, competitors map[int]competitor, res *log.Logger) {
	competitorsSlice := sortResultByTotalTime(competitors)

	for _, c := range competitorsSlice {
		penaltyInfo := fmt.Sprintf("{%s, %s}", c.AvgPenaltyTime, c.AvgPenaltySpeed)

		if c.NotStarted {
			res.Printf("NotStarted %d %v %v %s %d/%d\n",
				c.id, c.LapsMain, c.LapsPenalty, penaltyInfo, c.NumberOfHits, 5*cfg.Laps)
		} else if c.NotFinish {
			res.Printf("NotFinished %d %v %s %d/%d\n",
				c.id, c.LapsMain, penaltyInfo, c.NumberOfHits, 5*cfg.Laps)
		} else {
			res.Printf("%s %d %v %s %d/%d\n",
				c.TotalRouteTime, c.id, c.LapsMain, penaltyInfo, c.NumberOfHits, 5*cfg.Laps)
		}
	}
}

// превращает строку "HH:MM:SS" или "HH:MM:SS.sss" в time.Duration
func parseHMS(s string) (time.Duration, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %q", s)
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	// Обрабатываем секции с или без миллисекунд
	secPart := parts[2]
	var seconds int
	var milliseconds int
	if idx := strings.Index(secPart, "."); idx >= 0 {
		seconds, err = strconv.Atoi(secPart[:idx])
		if err != nil {
			return 0, err
		}
		msStr := secPart[idx+1:]
		// добавим недостающие нули справа, если нужно
		if len(msStr) < 3 {
			msStr += strings.Repeat("0", 3-len(msStr))
		} else if len(msStr) > 3 {
			msStr = msStr[:3]
		}
		milliseconds, err = strconv.Atoi(msStr)
		if err != nil {
			return 0, err
		}
	} else {
		seconds, err = strconv.Atoi(secPart)
		if err != nil {
			return 0, err
		}
	}

	total := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(milliseconds)*time.Millisecond

	return total, nil
}

func calculateLapTime(cfg *config.Config, c *competitor, event *events.Event) error {
	var startTime time.Time
	var err error

	if c.LastLapTime == "" {
		startTime, err = parseTime(c.ActualStart)
	} else {
		startTime, err = parseTime(c.LastLapTime)
	}
	if err != nil {
		return err
	}

	eventTime, err := parseTime(event.Time)
	if err != nil {
		return err
	}

	// Получили время круга круга
	lapDuration := eventTime.Sub(startTime)

	hours := int(lapDuration / time.Hour)
	minutes := int((lapDuration % time.Hour) / time.Minute)
	seconds := int((lapDuration % time.Minute) / time.Second)
	millis := int((lapDuration % time.Second) / time.Millisecond)
	lapTimeStr := fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)

	// Скорость = пройденные метры / секунды
	speed := float64(cfg.LapLen) / lapDuration.Seconds()
	speedStr := strconv.FormatFloat(speed, 'f', 3, 64)

	c.LapsMain = append(c.LapsMain, lap{Time: lapTimeStr, Speed: speedStr})
	c.LastLapTime = event.Time
	return nil
}

func calculatePenaltyLapTime(cfg *config.Config, c *competitor, event *events.Event) error {
	startTime, err := parseTime(c.PenaltyLenStart)
	if err != nil {
		return err
	}
	eventTime, err := parseTime(event.Time)
	if err != nil {
		return err
	}

	dur := eventTime.Sub(startTime)
	hours := int(dur / time.Hour)
	minutes := int((dur % time.Hour) / time.Minute)
	seconds := int((dur % time.Minute) / time.Second)
	millis := int((dur % time.Second) / time.Millisecond)
	durStr := fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)

	speed := float64(cfg.PenaltyLen) / dur.Seconds()
	speedStr := strconv.FormatFloat(speed, 'f', 3, 64)

	c.LapsPenalty = append(c.LapsPenalty, lap{Time: durStr, Speed: speedStr})

	// считаем среднее
	var total time.Duration
	var sumSpeed float64
	for _, p := range c.LapsPenalty {
		d, _ := parseHMS(p.Time)
		total += d
		sp, _ := strconv.ParseFloat(p.Speed, 64)
		sumSpeed += sp
	}
	avgDur := total / time.Duration(len(c.LapsPenalty))
	ah := int(avgDur / time.Hour)
	am := int((avgDur % time.Hour) / time.Minute)
	asec := int((avgDur % time.Minute) / time.Second)
	ams := int((avgDur % time.Second) / time.Millisecond)
	c.AvgPenaltyTime = fmt.Sprintf("%02d:%02d:%02d.%03d", ah, am, asec, ams)
	c.AvgPenaltySpeed = strconv.FormatFloat(sumSpeed/float64(len(c.LapsPenalty)), 'f', 3, 64)

	return nil
}

func calculateTimeDifference(cfg *config.Config, c *competitor) error {
	actual, err := parseTime(c.ActualStart)
	if err != nil {
		return err
	}

	planned, err := parseTime(c.WantStart)
	if err != nil {
		return err
	}
	delta, err := parseHMS(cfg.StartDelta)
	if err != nil {
		return err
	}

	diff := actual.Sub(planned)

	// фальстарт (возможно стоит как то учитывать)
	if diff < 0 {
		diff = 0
	}

	// не успел стартануть
	if diff > delta {
		c.NotStarted = true
		return nil
	}

	h := int(diff / time.Hour)
	m := int((diff % time.Hour) / time.Minute)
	s := int((diff % time.Minute) / time.Second)
	ms := int((diff % time.Second) / time.Millisecond)
	c.TimeDifference = fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
	return nil
}

// Общее время включает разницу между запланированным и фактическим временем старта
func calculateTotalTime(config *config.Config, c *competitor) error {
	finish, err := parseTime(c.FinishTime)
	if err != nil {
		return err
	}

	planned, err := parseTime(c.WantStart)
	if err != nil {
		return err
	}

	diff := finish.Sub(planned)

	h := int(diff / time.Hour)
	m := int((diff % time.Hour) / time.Minute)
	s := int((diff % time.Minute) / time.Second)
	ms := int((diff % time.Second) / time.Millisecond)
	c.TotalRouteTime = fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04:05.000", timeStr)
}
