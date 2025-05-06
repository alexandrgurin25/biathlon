# Системный прототип для биатлонных соревнований
Прототип должен уметь работать с файлом конфигурации и набором внешних событий определенного формата.
Решение должно содержать исходный файл/файлы golang (1.20 или новее) и модульные тесты (необязательно)

## Конфигурация  (json)

- **Laps**        - Количество кругов для основной дистанции
- **LapLen**      - Длина каждого основного круга
- **PenaltyLen**  - Длина каждого штрафного круга
- **FiringLines** - Количество огневых рубежей на круг
- **Start**       - Планируемое время старта для первого участника
- **StartDelta**  - Планируемый интервал между стартами

## События
Все события характеризуются временем и идентификатором события. Исходящие события - это события, создаваемые во время работы программы. События, относящиеся к категории "входящие", не могут быть сгенерированы и выводятся в той же форме, в которой они были представлены во входном файле.

- Все события происходят последовательно во времени.  (***Время события N+1***) >= (***Время события N***)
- Формат времен ***[HH:MM:SS.sss]***. Входные и выходные данные должны содержать завершающие нули

#### Общий формат для событий::
[***time***] **eventID** **competitorID** extraParams

```
Incoming events
EventID | extraParams | Comments
1       |             | The competitor registered
2       | startTime   | The start time was set by a draw
3       |             | The competitor is on the start line
4       |             | The competitor has started
5       | firingRange | The competitor is on the firing range
6       | target      | The target has been hit
7       |             | The competitor left the firing range
8       |             | The competitor entered the penalty laps
9       |             | The competitor left the penalty laps
10      |             | The competitor ended the main lap
11      | comment     | The competitor can`t continue
```
Участник дисквалифицируется, если он/она не стартует в течение своего стартового интервала. Это отмечается как **NotStarted** в итоговом отчете.
Если участник не может продолжать, это должно быть отмечено в итоговом отчете как **NotFinished**

```
Outgoing events
EventID | extraParams | Comments
32      |             | The competitor is disqualified
33      |             | The competitor has finished
```

## Итоговый отчет
Итоговый отчет должен содержать список всех зарегистрированных участников, отсортированных по возрастанию времени.
- Общее время включает разницу между запланированным и фактическим временем старта или отметки **NotStarted**/**NotFinished** marks
- Время, затраченное на прохождение каждого круга
- Средняя скорость для каждого круга [м/с]
- Время, затраченное на прохождение штрафных кругов
- Средняя скорость на штрафных кругах [м/с]
- Количество попаданий/количество выстрелов

Примеры:

`Config.conf`
```json
{
    "laps" : 2,
    "lapLen": 3651,
    "penaltyLen": 50,
    "firingLines": 1,
    "start": "09:30:00",
    "startDelta": "00:00:30"
}
```

`IncomingEvents`

```
[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:29:45.734] 3 1
[09:30:01.005] 4 1
[09:49:31.659] 5 1 1
[09:49:33.123] 6 1 1
[09:49:34.650] 6 1 2
[09:49:35.937] 6 1 4
[09:49:37.364] 6 1 5
[09:49:38.339] 7 1
[09:49:55.915] 8 1
[09:51:48.391] 9 1
[09:59:03.872] 10 1
[09:59:03.872] 11 1 Lost in the forest

```

`Output log`
```
[09:05:59.867] The competitor(1) registered
[09:15:00.841] The start time for the competitor(1) was set by a draw to 09:30:00.000
[09:29:45.734] The competitor(1) is on the start line
[09:30:01.005] The competitor(1) has started
[09:49:31.659] The competitor(1) is on the firing range(1)
[09:49:33.123] The target(1) has been hit by competitor(1)
[09:49:34.650] The target(2) has been hit by competitor(1)
[09:49:35.937] The target(4) has been hit by competitor(1)
[09:49:37.364] The target(5) has been hit by competitor(1)
[09:49:38.339] The competitor(1) left the firing range
[09:49:55.915] The competitor(1) entered the penalty laps
[09:51:48.391] The competitor(1) left the penalty laps
[09:59:03.872] The competitor(1) ended the main lap
[09:59:05.321] The competitor(1) can`t continue: Lost in the forest
```

`Resulting table`
```
[NotFinished] 1 [{00:29:03.872, 2.093}, {,}] {00:01:44.296, 0.481} 4/5
```

## Установка и запуск
```bash
   git clone https://github.com/alexandrgurin25/biathlon.git
   cd biathlon
   go run ./cmd/main.go
```
Для того, чтобы изменить конфигурацию, нужно изменить файл `config.json` в папке `config` <br>
Для того, чтобы изменить входящие события, нужно изменить файл `events` в папке `external_events`
Для тестирования, нужно запустить команду `go test ./internal/app`

## Архитектура решения
 Основные компоненты системы:
 - Конфигурация (pkg/config) - загрузка и парсинг параметров соревнований
 - Обработка событий (pkg/events) - чтение и парсинг входящих событий
 - Логика соревнований (internal/app) - ядро системы, обработка состояния участников
 - Логирование (pkg/logger) - запись выходного лога событий
    
## Особенности реализации

- **Проблема**: В логах нет событий типа "промах", только попадания (EventID=6).
- **Решение**:
  - Принято считать, что на каждом огневом рубеже 5 мишеней (стандарт биатлона).
  - Если EventID=6 меньше 5 — остальные считаются промахами.
  - Также считал в результирующией таблице сумму попаданий и выстрелов на всех кругах 
