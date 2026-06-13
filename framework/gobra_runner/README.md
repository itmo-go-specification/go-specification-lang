# gobra_runner

Пакет запускает Gobra на сгенерированном коде и переводит позиции ошибок обратно в исходные Go-файлы через гибридный маппинг.

## Runner.RunWithMapping()

```
func (r *Runner) RunWithMapping(gobraCode string, mapping *HybridMapping) (*VerificationResult, error)
```

**Входные данные:**

- `gobraCode string` — текст сгенерированного Gobra-файла (содержимое, не путь). Записывается во временный файл и передаётся Gobra.
- `mapping *HybridMapping` — гибридный маппинг позиций из сгенерированного файла в оригинальный Go-исходник. Если `nil`, ошибки возвращаются с позициями в сгенерированном файле.

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

**Статусы маппинга ошибок:**

- `IsMapped()` — ошибка успешно отображена на исходный код
- `IsInGeneratedCode()` — ошибка в сгенерированном коде без соответствия в исходном
- `IsInAddedCode()` — ошибка в коде, добавленном при трансляции
- `IsUnmapped()` — ошибка не может быть отображена (проблема с маппингом, см. `GenLine`/`GenColumn` для отладки)

## Runner.Run()

```
func (r *Runner) Run(gobraCode string) (*VerificationResult, error)
```

Упрощённый метод для запуска верификации без маппинга. Эквивалентен вызову `RunWithMapping(gobraCode, nil)`.

## HybridMapping

Гибридная система маппинга, которая поддерживает два типа отображений:

1. **Interval mappings** — для отображения целых интервалов добавленного кода
2. **Point mappings** — для точечного отображения позиций

### Создание маппинга

```go
mapping := gobra_runner.NewHybridMapping("source.go")

// Добавление интервала добавленного кода
mapping.AddInterval(gobra_runner.Interval{
    GenStart:  10,
    GenEnd:    15,
    SourcePos: gobra_runner.Position{File: "source.go", Line: 5, Column: 1},
    PredicateSpan: &gobra_runner.Position{File: "source.go", Line: 5, Column: 10},
})

// Добавление точечного маппинга
mapping.AddPointMapping(gobra_runner.PointMapping{
    GenLine:   20,
    GenColumn: 5,
    Orig:      gobra_runner.Position{File: "source.go", Line: 10, Column: 1},
})
```

### Отображение ошибок

```go
pos, err := mapping.MapError(genLine, genColumn)
if err == nil {
    // pos содержит позицию в исходном файле
}
```

## CoordMapping (Устаревший)

⚠️ **Deprecated:** Используйте `HybridMapping` вместо `CoordMapping`.

Старая структура для перевода позиций. Оставлена для обратной совместимости, но не рекомендуется к использованию в новом коде.

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
- `GPMA_DOCKER_IMAGE` — Docker образ для Gobra (по умолчанию: `"ghcr.io/viperproject/gobra:sha-7fa4f86"`)
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
| `DockerImage` | string | `"ghcr.io/viperproject/gobra:sha-7fa4f86"` | Docker образ для Gobra |
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

## Docker поддержка

### Проверка доступности Docker

Перед использованием Docker бэкенда можно проверить доступность Docker:

```go
import "gpma/framework/gobra_runner"

checker := gobra_runner.NewDockerChecker(10 * time.Second)
if err := checker.CheckDockerAvailability(); err != nil {
    fmt.Printf("Docker недоступен: %v\n", err)
    return
}
```

### Валидация Docker конфигурации

Проверьте конфигурацию перед запуском:

```go
settings := gobra_runner.DefaultSettings()
settings.ExecutionBackend = gobra_runner.BackendDocker
settings.DockerImage = "ghcr.io/viperproject/gobra:sha-7fa4f86"

if err := gobra_runner.ValidateDockerConfig(settings); err != nil {
    fmt.Printf("Неверная Docker конфигурация: %v\n", err)
    return
}
```

### Обработка Docker ошибок

Docker ошибки имеют специальный тип для лучшей диагностики:

```go
result, err := runner.Run(gobraCode)
if err != nil {
    if gobra_runner.IsDockerError(err) {
        dockerErr := err.(*gobra_runner.DockerError)
        fmt.Printf("Docker ошибка (контекст: %s): %v\n",
            dockerErr.Context, dockerErr.Err)
        fmt.Printf("Образ: %s, команда: %s\n",
            dockerErr.Image, dockerErr.Command)
    } else {
        fmt.Printf("Общая ошибка: %v\n", err)
    }
    return
}
```

## Примеры использования

### Базовое использование с нативным бэкендом

```go
package main

import (
    "fmt"
    "time"
    "gpma/framework/gobra_runner"
)

func main() {
    // Создаём настройки
    settings := gobra_runner.DefaultSettings()
    settings.GobraJAR = "/path/to/gobra.jar"
    settings.Timeout = 30 * time.Second
    
    // Создаём раннер
    runner := gobra_runner.New(settings)
    
    // Запускаем верификацию
    gobraCode := `
package main

func main() {
    assert(true)
}
`
    
    result, err := runner.Run(gobraCode)
    if err != nil {
        fmt.Printf("Ошибка запуска: %v\n", err)
        return
    }
    
    if result.Success {
        fmt.Println("Верификация прошла успешно")
    } else {
        fmt.Printf("Верификация не удалась:\n")
        for _, e := range result.Errors {
            fmt.Printf("  %s:%d:%d - %s\n", e.File, e.Line, e.Column, e.Message)
        }
    }
}
```

### Использование с гибридным маппингом

```go
package main

import (
    "fmt"
    "gpma/framework/gobra_runner"
)

func main() {
    settings := gobra_runner.DefaultSettings()
    settings.GobraJAR = "/path/to/gobra.jar"
    
    runner := gobra_runner.New(settings)
    
    // Создаём гибридный маппинг
    mapping := gobra_runner.NewHybridMapping("source.go")
    mapping.AddInterval(gobra_runner.Interval{
        GenStart:  5,
        GenEnd:    10,
        SourcePos: gobra_runner.Position{File: "source.go", Line: 3, Column: 1},
    })
    
    // Запускаем верификацию с маппингом
    gobraCode := `...`
    
    result, err := runner.RunWithMapping(gobraCode, mapping)
    if err != nil {
        fmt.Printf("Ошибка запуска: %v\n", err)
        return
    }
    
    // Ошибки будут отображены в исходный файл
    for _, e := range result.Errors {
        if e.IsMapped() {
            fmt.Printf("Ошибка в исходном коде: %s:%d:%d - %s\n",
                e.File, e.Line, e.Column, e.Message)
        } else if e.IsUnmapped() {
            // Ошибка не может быть отображена - проблема с маппингом
            fmt.Printf("ОШИБКА МАППИНГА: не удалось отобразить ошибку %d:%d - %s\n",
                e.GenLine, e.GenColumn, e.Message)
            fmt.Printf("  Проверьте сгенерированный Gobra код для отладки\n")
        } else {
            fmt.Printf("Ошибка в сгенерированном коде: %d:%d - %s\n",
                e.GenLine, e.GenColumn, e.Message)
        }
    }
}
```

### Использование с Docker бэкендом

```go
package main

import (
    "fmt"
    "gpma/framework/gobra_runner"
)

func main() {
    settings := gobra_runner.DefaultSettings()
    settings.ExecutionBackend = gobra_runner.BackendDocker
    settings.DockerImage = "ghcr.io/viperproject/gobra:sha-7fa4f86"
    
    runner := gobra_runner.New(settings)
    
    gobraCode := `...`
    
    result, err := runner.Run(gobraCode)
    if err != nil {
        fmt.Printf("Ошибка запуска: %v\n", err)
        return
    }
    
    if result.Success {
        fmt.Println("Верификация прошла успешно")
    }
}
```

### Использование переменных окружения

```bash
# Установка бэкенда выполнения
export GPMA_EXECUTION_BACKEND=docker

# Установка Docker образа
export GPMA_DOCKER_IMAGE=ghcr.io/viperproject/gobra:sha-7fa4f86

# Установка пути к Gobra JAR
export GPMA_GOBRA_JAR=/path/to/gobra.jar

# Установка таймаута
export GPMA_TIMEOUT=60s
```

## Тестирование

Тесты в `runner_test.go` и `output_parser_test.go` покрывают:
- Создание и конфигурацию раннера
- Построение команд для нативного и Docker бэкендов
- Парсинг вывода Gobra (оба формата)
- Гибридный маппинг ошибок
- Обработку таймаутов и ошибок запуска

Gobra и Java не нужны — тесты работают с фиктивным выводом и mock-командами.

```
go test gpma/framework/gobra_runner
go test -v gpma/framework/gobra_runner
```
