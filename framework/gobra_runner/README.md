# gobra_runner

Пакет запускает Gobra на сгенерированном коде и переводит позиции ошибок обратно в исходные Go-файлы через координатный маппинг.

## Runner.Run()

```
func (r *Runner) Run(gobraCode string, mapping *CoordMapping) (*VerificationResult, error)
```

**Входные данные:**

- `gobraCode string` — текст сгенерированного Gobra-файла (содержимое, не путь). Записывается во временный файл и передаётся Gobra.
- `mapping *CoordMapping` — маппинг позиций из сгенерированного файла в оригинальный Go-исходник. Если `nil`, ошибки возвращаются с позициями в сгенерированном файле.

**Выходные данные:**

- `*VerificationResult` — результат верификации:
  - `Success bool` — `true`, если Gobra завершился без ошибок и без таймаута
  - `TimedOut bool` — `true`, если процесс был убит по таймауту из `Settings.Timeout`
  - `Errors []VerificationError` — список ошибок верификации (пустой при успехе)
  - `RawOutput string` — необработанный stdout+stderr процесса Gobra
- `error` — возвращается только при внутренней ошибке: Gobra не найден, упал с паникой, не смог создать временный файл. Ошибки верификации (precondition/postcondition и т.д.) не являются `error` — они попадают в `VerificationResult.Errors`.

**Каждый `VerificationError` содержит:**

- `File`, `Line`, `Column`, `Offset` — позиция в оригинальном Go-файле (заполняется из `mapping`; пустые, если маппинг не передан или не найден)
- `GenLine`, `GenColumn` — позиция в сгенерированном Gobra-файле (всегда заполнены)
- `Kind` — вид нарушения: `"precondition"`, `"postcondition"`, `"invariant"`, `"assertion"`, `"unknown"`
- `Message` — текст ошибки из вывода Gobra
- `MappingInfo` — дополнительная информация о маппинге (тип маппинга, уверенность, и т.д.)

## CoordMapping

Структура для перевода позиций. Каждый `MappingEntry` связывает одну позицию в сгенерированном файле (`GenLine`, `GenColumn`) с позицией в оригинальном файле (`Orig.File`, `Orig.Line`, `Orig.Column`) и видом спецификации (`Kind`).

`Lookup(line, col int)` возвращает запись с наибольшей позицией, не превышающей `(line, col)` — то есть ближайшую предшествующую аннотацию. Перед вызовом `Lookup` или `Run` маппинг должен быть отсортирован через `Sort()`.

Маппинг можно сохранить в JSON и загрузить обратно:

```go
gobra_runner.SaveMappingToFile(m, "mapping.json")
m, err := gobra_runner.LoadMappingFromFile("mapping.json")
```

## Настройки

`DefaultSettings()` возвращает рабочие значения по умолчанию. Обязательно переопределить `GobraJAR` (или `GobraBin` для нативного бинарника):

```go
s := gobra_runner.DefaultSettings()
s.GobraJAR = "/path/to/gobra.jar"
r := gobra_runner.New(s)
```

### Поддержка переменных окружения

Настройки могут быть переопределены через переменные окружения:

- `GPMA_EXECUTION_BACKEND` — бэкенд выполнения: `"native"` или `"docker"` (по умолчанию: `"native"`)
- `GPMA_DOCKER_IMAGE` — Docker образ для Gobra (по умолчанию: `"ghcr.io/viperproject/gobra:latest"`)
- `GPMA_JAVA_PATH` — путь к Java (по умолчанию: `"java"`)
- `GPMA_GOBRA_JAR` — путь к gobra.jar (по умолчанию: `"gobra.jar"`)
- `GPMA_GOBRA_BACKEND` — бэкенд Gobra: `"silicon"` или `"carbon"` (по умолчанию: `"silicon"`)
- `GPMA_TIMEOUT` — таймаут в формате duration (по умолчанию: `"120s"`)
- `GPMA_LOG_LEVEL` — уровень логов Gobra (по умолчанию: `"ERROR"`)

### Поля `Settings`

| Поле | Тип | По умолчанию | Описание |
|------|-----|--------------|----------|
| **Execution Backend** | | | |
| `ExecutionBackend` | `Backend` | `BackendNative` | Бэкенд выполнения: `BackendNative` или `BackendDocker` |
| `DockerImage` | string | `"ghcr.io/viperproject/gobra:latest"` | Docker образ для Gobra |
| **Native Execution** | | | |
| `JavaPath` | string | `"java"` | Путь к java |
| `GobraJAR` | string | `"gobra.jar"` | Путь к gobra.jar |
| `GobraBin` | string | `"gobra"` | Нативный бинарник (вместо JAR) |
| `Z3Exe` | string | `"z3"` | Путь к Z3 |
| **Gobra Verifier** | | | |
| `GobraBackend` | string | `"silicon"` | Бэкенд Gobra: `"silicon"` или `"carbon"` |
| `Timeout` | time.Duration | `120s` | Таймаут; 0 — без ограничений |
| `LogLevel` | string | `"ERROR"` | Уровень логов Gobra |
| **Verification Flags** | | | |
| `NoInformation` | bool | `true` | Подавить информационные сообщения |
| `ParallelizeBranches` | bool | `false` | Параллелизация веток |
| `AssumeInjectivityOnInhale` | bool | `false` | Предполагать инъективность при inhale |
| `CheckOverflow` | bool | `false` | Проверка переполнения целых |
| `ExtraFlags` | []string | — | Дополнительные флаги Gobra |

### Тип `Backend`

```go
type Backend string

const (
    BackendNative Backend = "native"
    BackendDocker Backend = "docker"
)
```

Метод `IsValid()` проверяет корректность значения бэкенда.

## Тестирование

Тесты в `output_parser_test.go` покрывают: парсинг вывода Gobra (оба формата), фильтрацию по файлу, `CoordMapping.Lookup`, `mapErrors` с маппингом и без, инференс вида ошибки. Gobra и Java не нужны — тесты работают с фиктивным выводом.

```
go test gpma/framework/gobra_runner
go test -v gpma/framework/gobra_runner
```
