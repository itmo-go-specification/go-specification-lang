# Плагин `dynamic_checks`

Плагин читает ACSL-аннотации в комментариях и вставляет в код вызовы `checks.Check(...)`, которые проверяют контракты во время выполнения программы. Оригинальные файлы не изменяются — результат применяется через `go build --overlay`.

## Быстрый старт

```bash
# Сгенерировать overlay и запустить с проверками
go run framework/cli/main.go project --plugin dynamic_checks \
  ./samples/projects/abs/ ./overlays/

go run --overlay overlays/gpma_overlay.json samples/projects/abs/abs.go
```

---

## Что можно писать в аннотациях

### Контракт функции

Пишется в комментарии `/*@ ... */` прямо перед функцией:

```go
/*@
requires x >= 0;
ensures \result >= 0;
ensures (\result == x) || (\result == -x);
assigns \nothing;
*/
func Abs(x int) int { ... }
```

Плагин вставит `checks.Check(...)` для каждого `requires` в начало функции и `defer checks.Check(...)` для каждого `ensures`.

#### `requires` — предусловие

Условие, которое должно быть истинным при вызове функции:

```go
/*@
requires n > 0;
requires x >= 0 && x < n;
*/
func Get(a []int, x, n int) int { ... }
```

#### `ensures` — постусловие

Условие, которое должно быть истинным при возврате из функции. Можно использовать `\result` для обращения к возвращаемому значению:

```go
/*@
ensures \result >= lo && \result <= hi;
*/
func Clamp(x, lo, hi int) int { ... }
```

> ⚠️ `\result` работает только для функций с **одним** возвращаемым значением.

Можно использовать `\old(expr)` для обращения к значению выражения **на момент входа** в функцию:

```go
/*@
ensures (*p == \old(*q)) && (*q == \old(*p));
*/
func swap(p *int, q *int) { ... }
```

#### `assigns` — список изменяемых переменных

Синтаксис поддерживается парсером, но плагин `dynamic_checks` его **игнорирует** — клауза предназначена для других плагинов:

```go
assigns \nothing;   // игнорируется
assigns x, y;       // игнорируется
```

---

### Утверждения внутри функции

Пишутся в однострочном комментарии `/*@ assert ...; */` внутри тела функции. Плагин вставит `checks.Check(...)` перед следующим оператором:

```go
func clamp(x, min, max int) int {
    /*@ assert min <= max; */
    if x < min { return min }

    /*@ assert x >= min; */
    if x > max { return max }

    return x
}
```

Утверждения работают в любом вложенном блоке: `if`, `for`, `switch`, `select`.

> `\old` и `\result` в `assert` **не поддерживаются**.

---

### Аннотации циклов

Синтаксис поддерживается парсером, но вставка проверок в тело цикла **ещё не реализована** (в разработке):

```go
/*@
loop invariant 0 <= i && i <= n;
loop variant n - i;
*/
for i := 0; i < n; i++ { ... }
```

---

## Что можно писать в выражениях

### Операторы

| Оператор | Пример | Работает |
|---|---|---|
| Сравнение | `x == y`, `x != y`, `x < y`, `x <= y`, `x > y`, `x >= y` | ✅ |
| Арифметика | `x + y`, `x - y`, `x * y`, `x / y`, `x % y` | ✅ |
| Логика | `p && q`, `p \|\| q`, `!p` | ✅ |
| Исключающее ИЛИ | `p ^^ q` | ✅ |
| Унарный минус/плюс | `-x`, `+x` | ✅ |
| Тернарный оператор | `cond ? a : b` | ✅ |
| Импликация | `p ==> q` | ❌ используйте `!p \|\| q` |
| Эквивалентность | `p <==> q` | ❌ |

### Специальные термы

| Терм | Значение | Где работает |
|---|---|---|
| `\result` | возвращаемое значение функции | только в `ensures` |
| `\old(expr)` | значение `expr` на входе в функцию | только в `ensures` |
| `\true`, `\false` | булевы константы | везде |

### Вызовы функций

В выражениях можно вызывать обычные Go-функции:

```go
/*@
ensures \result == Max(a, b);
*/
func max(a, b int) int { ... }
```

### Квантификаторы

Плагин поддерживает `\forall` и `\exists` с целочисленными переменными (`integer`).

Тело квантификатора должно иметь стандартную форму с явными границами:

- **`\forall`**: `(lo <= i && i < hi) ==> P(i)`
- **`\exists`**: `(lo <= i && i < hi) && P(i)`

Пример:

```c
/*@
requires \forall integer i; 0 <= i && i < n ==> a[i] >= 0;
ensures  \exists integer i; 0 <= i && i < n && \result == a[i];
*/
```

Генерируемый код использует IIFE (immediately-invoked function expression) с циклом `for`:

```go
// \forall integer i; lo <= i && i < hi ==> P(i)
func() bool {
    for i := lo; i < hi; i++ {
        if !(P(i)) { return false }
    }
    return true
}()

// \exists integer i; lo <= i && i < hi && P(i)
func() bool {
    for i := lo; i < hi; i++ {
        if P(i) { return true }
    }
    return false
}()
```

---

## Режимы работы

### Строгость (`strict_mode`)

По умолчанию плагин работает в **строгом режиме**: нарушение `requires` — это ошибка, `ensures` проверяется всегда.

В **нестрогом режиме** нарушение `requires` не является ошибкой само по себе, а `ensures` проверяется только если все `requires` выполнены:

```bash
go run framework/cli/main.go project --plugin dynamic_checks \
  --plugin-config 'dynamic_checks:strict_mode:false' \
  ./samples/projects/interactive/ ./overlays/
```

### Реакция на нарушение (`CHECK_MODE`)

Управляется переменной окружения при запуске программы:

```bash
# Паника (по умолчанию)
go run --overlay overlays/gpma_overlay.json samples/projects/abs/abs.go

# Вывод в stderr
CHECK_MODE=console go run --overlay overlays/gpma_overlay.json samples/projects/abs/abs.go

# Запись в файл
CHECK_MODE=file CHECK_LOG_PATH=violations.txt \
  go run --overlay overlays/gpma_overlay.json samples/projects/abs/abs.go

# Игнорировать нарушения
CHECK_MODE=silent go run --overlay overlays/gpma_overlay.json samples/projects/abs/abs.go
```

---

## Примеры

### Функция с предусловием и постусловием

```go
/*@
requires x >= 0;
ensures \result >= 0;
*/
func Abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

После обработки плагином:

```go
func Abs(x int) int {
    checks.Check(x >= 0, "precondition")
    var __result int
    defer checks.Check(__result >= 0, "postcondition")
    if x < 0 {
        { __result = -x; return __result }
    }
    { __result = x; return __result }
}
```

### Функция с `\old`

```go
/*@
ensures (*p == \old(*q)) && (*q == \old(*p));
*/
func swap(p *int, q *int) {
    save := *p
    *p = *q
    *q = save
}
```

После обработки:

```go
func swap(p *int, q *int) {
    __old_1 := *q
    __old_2 := *p
    defer checks.Check(*p == __old_1 && *q == __old_2, "postcondition")
    save := *p
    *p = *q
    *q = save
}
```

