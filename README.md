# go-specification-lang

**go-specification-lang** — фреймворк для анализа и трансформации программ на Go, вдохновлённый [Frama-C](https://frama-c.com/): так же, как Frama-C позволяет анализировать Си-программы с помощью аннотаций ACSL, этот проект привносит аналогичный подход в экосистему Go.

Разработчик вносит спецификации прямо в исходный код в виде аннотаций на языке **GPMA** (Go Program Modelling Annotations) — диалекте ACSL, адаптированном для Go:

```go
/*@
requires x >= 0;
ensures \result >= 0;
*/
func Abs(x int) int { ... }
```

Фреймворк разбирает эти аннотации, связывает их с соответствующими узлами AST и передаёт результат в **плагины**. Каждый плагин получает AST программы вместе со спецификациями и может трансформировать код произвольным образом. Плагины — основная точка расширения фреймворка.

В качестве примера в репозитории реализован плагин **`dynamic_checks`**, который вставляет в код вызовы `checks.Check(...)` для проверки предусловий, постусловий и утверждений во время выполнения. Результат применяется через механизм `go build --overlay`, не изменяя оригинальные файлы.

## Требования

- Go 1.21+
- (опционально) [gvm](https://github.com/moovweb/gvm) для управления версиями Go
- (опционально) Java — для перегенерации ANTLR-парсера

## Установка

Готовый дистрибутив включает `gpma` CLI, верификатор [Gobra](https://github.com/viperproject/gobra), Z3 и Boogie.  
Поддерживается **linux/amd64**. Для запуска программ с overlay (`go run --overlay`) по-прежнему нужен установленный Go toolchain.

### Вариант 1: `.deb`-пакет

**Скачать готовый пакет** — из артефактов GitHub Actions (workflow *Package*, хранение 3 дня) или из релиза, если он опубликован.

**Собрать локально:**

```bash
just build-deb
```

**Установить:**

```bash
sudo apt install ./packaging/dist/gpma_*.deb
```

Пакет подтянет зависимости: `openjdk-21-jre-headless`, `ca-certificates`.  
Z3, Boogie и совместимые библиотеки OpenSSL 1.1 для Boogie уже включены в пакет.

После установки доступны:

| Команда | Назначение |
|---------|------------|
| `gpma` | CLI фреймворка |
| `gobra` | обёртка над Gobra verifier |

Проверка:

```bash
gpma plugins
gobra -i /usr/share/gpma/examples/smoke.gobra
```

Пример с плагином `dynamic_checks`:

```bash
gpma project -p dynamic_checks ./samples/projects/simple/ ./overlays/
go run --overlay overlays/gpma_overlay.json ./samples/projects/simple/main.go
```

Переменные для Gobra (пути к JAR, Z3, Boogie) заданы в `/etc/gpma/gobra.env`.  
Плагин `gobra` также можно настроить через `--plugin-config` (`backend`, `gobra_jar`, `z3_exe`, `boogie_exe`).

### Вариант 2: Docker-образ

Образ на базе Ubuntu 22.04 с установленным `gpma` deb-пакетом.

**Полная сборка** (deb собирается внутри Docker, на хосте ничего кроме Docker не нужно):

```bash
just docker-gpma
```

**Быстрая пересборка** (сначала собирается deb на хосте, затем устанавливается в образ):

```bash
just docker-gpma-runtime
```

Проверка:

```bash
docker run --rm gpma:latest plugins
docker run --rm --entrypoint gobra gpma:latest \
  -i /usr/share/gpma/examples/smoke.gobra
```

Обработка проекта с монтированием рабочей директории:

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest project \
  -p dynamic_checks /work/samples/projects/simple/ /work/overlays/
```

Запуск с overlay по-прежнему выполняется на хосте через `go run --overlay`, используя сгенерированный `gpma_overlay.json`.

## Сборка

### Через Makefile (рекомендуется)

`make` автоматически запускает `go generate` для сборки ANTLR-парсера, а затем компилирует бинарник:

```bash
cd framework
make build
# бинарник появится в framework/build/framework
```

### Напрямую через `go build`

Перед сборкой необходимо вручную сгенерировать ANTLR-парсер (требует Java):

```bash
cd framework
go generate ./...
go build -o build/framework ./cli/main.go
```

## CLI

CLI-приложение (`framework/cli/main.go`) принимает Go-файлы или целые проекты, прогоняет их через один или несколько **плагинов** и записывает результат. Плагины трансформируют AST исходного кода — например, вставляют динамические проверки контрактов. Можно подключать несколько плагинов одновременно через повторный флаг `--plugin`.

Полная справка по любой команде:

```bash
go run framework/cli/main.go --help
go run framework/cli/main.go file --help
go run framework/cli/main.go project --help
```

Список зарегистрированных плагинов:

```bash
go run framework/cli/main.go plugins
```

### Обработка одного файла

```bash
go run framework/cli/main.go file --plugin <plugin> <source.go> <output.go>
```

### Обработка целого проекта

Обходит все `.go`-файлы в директории проекта, применяет плагины и записывает результат в `<output-dir/>`. Дополнительно создаёт файл `gpma_overlay.json`, который можно передать в `go run --overlay` для запуска изменённого кода без перезаписи оригиналов:

```bash
go run framework/cli/main.go project --plugin <plugin> <project-dir/> <output-dir/>
```

## Плагин `dynamic_checks`

Плагин `dynamic_checks` читает ACSL-аннотации в комментариях (`/*@ ... */`) и вставляет в код вызовы `checks.Check(...)` для проверки предусловий (`requires`), постусловий (`ensures`) и утверждений (`assert`) во время выполнения.

### Генерация overlay и запуск

Сначала сгенерируйте overlay — директорию с изменёнными файлами и файлом `gpma_overlay.json`:

```bash
go run framework/cli/main.go project --plugin dynamic_checks \
  ./samples/projects/simple/ ./overlays/
```

После этого можно запустить программу без проверок (оригинальный код):

```bash
go run samples/projects/simple/main.go
```

Или с включёнными проверками через overlay (оригинальные файлы не изменяются):

```bash
go run --overlay overlays/gpma_overlay.json samples/projects/simple/main.go
```

### Режим строгости (`strict_mode`)

**Строгий** (по умолчанию) — предусловия и постусловия проверяются безусловно:

```bash
go run framework/cli/main.go project -p dynamic_checks \
  --plugin-config 'dynamic_checks:strict_mode:true' \
  samples/projects/interactive/ overlays/
```

**Нестрогий** — постусловия проверяются только если предусловия выполнены; нарушение предусловий само по себе не является ошибкой:

```bash
go run framework/cli/main.go project -p dynamic_checks \
  --plugin-config 'dynamic_checks:strict_mode:false' \
  samples/projects/interactive/ overlays/
```

### Режим реакции на нарушения (`CHECK_MODE`)

Поведение при нарушении контракта управляется переменной окружения `CHECK_MODE`:

| Значение  | Поведение                               |
|-----------|-----------------------------------------|
| `panic`   | паника (по умолчанию)                   |
| `console` | вывод в stderr                          |
| `file`    | запись в файл (требует `CHECK_LOG_PATH`) |
| `silent`  | игнорировать нарушения                  |

```bash
# Запись нарушений в файл
CHECK_MODE=file CHECK_LOG_PATH=violations.txt \
  go run --overlay overlays/gpma_overlay.json samples/projects/interactive/main.go

# Тихий режим — нарушения игнорируются
CHECK_MODE=silent go run --overlay overlays/gpma_overlay.json samples/projects/interactive/main.go

# Паника при нарушении
CHECK_MODE=panic go run --overlay overlays/gpma_overlay.json samples/projects/interactive/main.go
```

## Тесты

Запуск всех тестов фреймворка:

```bash
cd framework
make test
# или напрямую:
go test ./...
```

### Golden-тесты плагина `dynamic_checks`

Golden-тесты проверяют, что плагин генерирует ожидаемый код для каждого проекта из `samples/projects/`.

**Запуск проверки:**

```bash
cd framework
go test ./plugin/dynamic_checks/... -v
```

**Обновление golden-файлов** (после изменения логики генерации):

```bash
cd framework
go test ./plugin/dynamic_checks/... -update
```

Флаг `-update` перезаписывает файлы в `framework/plugin/dynamic_checks/testdata/golden/` актуальным выводом плагина. После обновления зафиксируйте изменения в git, чтобы зафиксировать новое ожидаемое поведение.
