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
	NotStarted   bool
	NotFinish    bool
	Disqualified bool

	LapsMain     []lap
	LapsPenalty  []lap
	NumberOfHits int

	TimeDifference string
	WantStart      string
	ActualStart    string
	LastLapTime    string
	FinishTime     string

	PenaltyLenStart string
	AvgPenaltyTime  string
	AvgPenaltySpeed string

	CompletedLaps int
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

		competitors[cID] = c
	}
	allEvents := append(incommingEvents, outgoingEvents...)

	sortEventsByTime(allEvents)

	WriteToOutputLog(log, allEvents)

	// Вывод таблицы результатов
	outputResultTable(cfg, competitors, incommingEvents, result)
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

func outputResultTable(cfg *config.Config, competitors map[int]competitor, events []events.Event, res *log.Logger) {

	sequentialRegistration := make([]int, 0)

	for i := 0; i < len(events); i++ {
		if events[i].EventID == 1 {
			sequentialRegistration = append(sequentialRegistration, events[i].CompetitorID)
		}
	}

	for i := 0; i < len(competitors); i++ {
		id := sequentialRegistration[i]
		c := competitors[id]

		if c.NotStarted {
			res.Printf("NotStarted %d %v %v {%s, %s} %d/%d\n", id, c.LapsMain, c.LapsPenalty, c.AvgPenaltyTime, c.AvgPenaltySpeed, c.NumberOfHits, 5*cfg.Laps)
		} else if c.NotFinish {
			res.Printf("NotFinished %d %v {%s, %s} %d/%d\n", id, c.LapsMain, c.AvgPenaltyTime, c.AvgPenaltySpeed, c.NumberOfHits, 5*cfg.Laps)
		} else {
			res.Printf("%s %d %v {%s, %s} %d/%d\n", c.TimeDifference, id, c.LapsMain, c.AvgPenaltyTime, c.AvgPenaltySpeed, c.NumberOfHits, 5*cfg.Laps)
		}
	}

}

func calculateLapTime(cfg *config.Config, c *competitor, event *events.Event) error {
	var lapTimeSeconds float64
	var actualStartTime time.Time
	var err error

	if c.LastLapTime == "" {
		actualStartTime, err = parseTime(c.ActualStart)
	} else {
		actualStartTime, err = parseTime(c.LastLapTime)
	}

	if err != nil {
		return fmt.Errorf("actualStartTime error: %w", err)
	}

	eventTime, err := parseTime(event.Time)
	if err != nil {
		return fmt.Errorf("eventTime error:%w", err)
	}

	// Находим разницу от текущего события и начала круга
	lapDuration := eventTime.Sub(actualStartTime)

	// Перевод из time.Duration в time.Time
	lapDate := time.Time{}.Add(lapDuration).String()
	parts := strings.Split(lapDate, " ")

	//Убираем лишние данные из времени
	partsTimeStr := strings.FieldsFunc(parts[1], func(r rune) bool {
		return r == ':' || r == '.'
	})
	partsTime := make([]int, len(partsTimeStr))
	for i := 0; i < len(partsTimeStr); i++ {
		partsTime[i], err = strconv.Atoi(partsTimeStr[i])
		if err != nil {
			return fmt.Errorf("invalid converted to type: %w", err)
		}
	}

	lapTimeSeconds = sumSecondsInTime(lapTimeSeconds, partsTime)

	speedInFloat := float64(cfg.LapLen) / lapTimeSeconds
	speedInString := strconv.FormatFloat(float64(speedInFloat), 'f', 3, 64)

	lap := lap{
		Time:  parts[1],
		Speed: speedInString,
	}

	//Сохранение состояния
	c.LapsMain = append(c.LapsMain, lap)
	c.LastLapTime = event.Time

	return nil
}

func calculatePenaltyLapTime(cfg *config.Config, c *competitor, event *events.Event) error {
	var lapTimeSeconds float64

	actualPenaltyStartTime, err := parseTime(c.PenaltyLenStart)
	if err != nil {
		return fmt.Errorf("actualStartTime error: %w", err)
	}

	eventTime, err := parseTime(event.Time)
	if err != nil {
		return fmt.Errorf("eventTime error:%w", err)
	}

	// Находим разницу от текущего события и начала круга
	lapDuration := eventTime.Sub(actualPenaltyStartTime)

	// Перевод из time.Duration в time.Time
	lapDate := time.Time{}.Add(lapDuration).String()
	parts := strings.Split(lapDate, " ")

	//Убираем лишние данные из времени
	partsTimeStr := strings.FieldsFunc(parts[1], func(r rune) bool {
		return r == ':' || r == '.'
	})
	partsTime := make([]int, len(partsTimeStr))
	for i := 0; i < len(partsTimeStr); i++ {
		partsTime[i], err = strconv.Atoi(partsTimeStr[i])
		if err != nil {
			return fmt.Errorf("invalid converted to type: %w", err)
		}
	}

	lapTimeSeconds = sumSecondsInTime(lapTimeSeconds, partsTime)

	speedInFloat := float64(cfg.PenaltyLen) / lapTimeSeconds
	speedInString := strconv.FormatFloat(float64(speedInFloat), 'f', 3, 64)

	lap := lap{
		Time:  parts[1],
		Speed: speedInString,
	}

	//Сохранение состояния
	c.LapsPenalty = append(c.LapsPenalty, lap)

	//Считаем среднеее значение времени и скорости
	var totalTime time.Duration

	for _, penalty := range c.LapsPenalty {

		duration, err := parseTimeToDuration(penalty.Time)
		if err != nil {
			return fmt.Errorf("eventTime error: %w", err)
		}
		totalTime += duration
	}

	avgDuration := totalTime / time.Duration(len(c.LapsPenalty))

	// Форматируем в HH:MM:SS.000
	hours := int(avgDuration.Hours())
	minutes := int(avgDuration.Minutes()) % 60
	seconds := int(avgDuration.Seconds()) % 60
	milliseconds := (avgDuration.Nanoseconds() / 1e6) % 1000 //1e6 - равно 1 миллиону

	c.AvgPenaltyTime = fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)

	var totalSpeed float64

	for _, penalty := range c.LapsPenalty {
		speedFloat, err := strconv.ParseFloat(penalty.Speed, 64)
		if err != nil {
			return fmt.Errorf("ParseFloat error: %w", err)
		}
		totalSpeed += speedFloat
	}

	avgSpeed := totalSpeed / float64(len(c.LapsPenalty))
	c.AvgPenaltySpeed = strconv.FormatFloat(avgSpeed, 'f', 3, 64)
	return nil
}

func sumSecondsInTime(lapTimeSeconds float64, partsTime []int) float64 {
	// Переводим время одного круга в секунды
	if len(partsTime) >= 1 {
		lapTimeSeconds += (float64(partsTime[0]) * 60 * 60) //Часы
	}

	if len(partsTime) >= 2 {
		lapTimeSeconds += float64(partsTime[1]) * 60 // Минуты
	}

	if len(partsTime) >= 3 {
		lapTimeSeconds += float64(partsTime[2]) // Секунды
	}

	if len(partsTime) >= 4 {
		lapTimeSeconds += float64(partsTime[3]) * float64(0.001) //Милисекунды
	}

	return lapTimeSeconds
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
		hours := int(timeDifference.Hours())
		minutes := int(timeDifference.Minutes()) % 60
		seconds := int(timeDifference.Seconds()) % 60
		milliseconds := (timeDifference.Nanoseconds() / 1e6) % 1000 //1e6 - равно 1 миллиону

		c.TimeDifference = fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)

	} else {
		c.NotStarted = true
	}

	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04:05.000", timeStr)
}

func parseTimeToDuration(timeStr string) (time.Duration, error) {
	// Убедимся, что есть миллисекунды
	if len(timeStr) < 12 {
		timeStr = timeStr + ".000"
	}

	t, err := time.Parse("15:04:05.000", timeStr)
	if err != nil {
		return 0, err
	}

	// Берём только время
	return time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second +
		time.Duration(t.Nanosecond())*time.Nanosecond, nil
}
