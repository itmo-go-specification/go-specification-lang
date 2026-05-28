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

- `File`, `Line`, `Column` — позиция в оригинальном Go-файле (заполняется из `mapping`; пустые, если маппинг не передан или не найден)
- `GenLine`, `GenColumn` — позиция в сгенерированном Gobra-файле (всегда заполнены)
- `Kind` — вид нарушения: `"precondition"`, `"postcondition"`, `"invariant"`, `"assertion"`, `"unknown"`
- `Message` — текст ошибки из вывода Gobra

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

Поля `Settings`:

| Поле | Тип | По умолчанию | Описание |
|------|-----|--------------|----------|
| `JavaPath` | string | `"java"` | Путь к java |
| `GobraJAR` | string | `"gobra.jar"` | Путь к gobra.jar |
| `GobraBin` | string | — | Нативный бинарник (вместо JAR) |
| `Timeout` | time.Duration | `120s` | Таймаут; 0 — без ограничений |
| `Backend` | string | `"silicon"` | Бэкенд: `"silicon"` или `"carbon"` |
| `LogLevel` | string | `"ERROR"` | Уровень логов Gobra |
| `NoInformation` | bool | `true` | Подавить информационные сообщения |
| `Z3Exe` | string | — | Путь к Z3 |
| `CheckOverflow` | bool | `false` | Проверка переполнения целых |
| `ExtraFlags` | []string | — | Дополнительные флаги Gobra |

## Тестирование

Тесты в `output_parser_test.go` покрывают: парсинг вывода Gobra (оба формата), фильтрацию по файлу, `CoordMapping.Lookup`, `mapErrors` с маппингом и без, инференс вида ошибки. Gobra и Java не нужны — тесты работают с фиктивным выводом.

```
go test gpma/framework/gobra_runner
go test -v gpma/framework/gobra_runner
```
