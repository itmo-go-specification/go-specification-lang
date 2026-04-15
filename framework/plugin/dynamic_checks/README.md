# Плагин `dynamic_checks`

Плагин читает ACSL-аннотации в комментариях и вставляет в код вызовы `checks.CheckAt(...)`, которые проверяют контракты во время выполнения программы. Оригинальные файлы не изменяются — результат применяется через `go build --overlay`.

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

Плагин вставит `checks.CheckAt(...)` для каждого `requires` в начало функции и `defer checks.CheckAt(...)` для каждого `ensures`.

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

Условие, которое должно быть истинным при возврате из функции.

**Одно возвращаемое значение** — используйте `\result`:

```go
/*@
ensures \result >= lo && \result <= hi;
*/
func Clamp(x, lo, hi int) int { ... }
```

**Несколько возвращаемых значений (неименованные)** — используйте `\result_0`, `\result_1`, ...:

```go
/*@
requires b != 0;
ensures \result_0 == a / b;
ensures \result_1 == a % b;
*/
func DivMod(a, b int) (int, int) { ... }
```

**Несколько возвращаемых значений (именованные)** — используйте `\result_0`, `\result_1`, ... или имена переменных напрямую через `\result_N` (плагин автоматически подставит имена из сигнатуры):

```go
/*@
requires len(a) > 0;
ensures \result_0 <= \result_1;
*/
func MinMax(a []int) (min, max int) { ... }
```

> При именованных возвращаемых значениях `\result_0` → `min`, `\result_1` → `max` и т.д. — плагин использует имена из сигнатуры функции.

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

Пишутся в однострочном комментарии `/*@ assert ...; */` внутри тела функции. Плагин вставит `checks.CheckAt(...)` перед следующим оператором:

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
| `\result` | возвращаемое значение (для функций с одним результатом) | только в `ensures` |
| `\result_0` | первое возвращаемое значение | только в `ensures` |
| `\result_N` | N-е возвращаемое значение (нумерация с 0) | только в `ensures` |
| `\old(expr)` | значение `expr` на входе в функцию | только в `ensures` |
| `\true`, `\false` | булевы константы | везде |

> При именованных возвращаемых значениях `\result_N` автоматически разрешается в имя N-го параметра из сигнатуры функции.

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

### Функция с одним возвращаемым значением

```go
/*@
ensures \result >= 0;
ensures (\result == x) || (\result == -x);
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
    var __result_0 int
    defer checks.CheckAt(__result_0 >= 0, "abs.go", 2, "postcondition")
    defer checks.CheckAt(__result_0 == x || __result_0 == -x, "abs.go", 3, "postcondition")
    if x < 0 {
        {
            __result_0 = -x
            return __result_0
        }
    }
    {
        __result_0 = x
        return __result_0
    }
}
```

> `\result` — синоним `\result_0`. Для функций с одним результатом оба варианта эквивалентны.

### Функция с несколькими неименованными возвращаемыми значениями

```go
/*@
requires b != 0;
ensures \result_0 == a / b;
ensures \result_1 == a % b;
ensures \result_0 * b + \result_1 == a;
*/
func DivMod(a, b int) (int, int) {
    return a / b, a % b
}
```

После обработки:

```go
func DivMod(a, b int) (int, int) {
    checks.CheckAt(b != 0, "divmod.go", 2, "precondition")
    var __result_0 int
    var __result_1 int
    defer checks.CheckAt(__result_0 == a/b, "divmod.go", 3, "postcondition")
    defer checks.CheckAt(__result_1 == a%b, "divmod.go", 4, "postcondition")
    defer checks.CheckAt(__result_0*b+__result_1 == a, "divmod.go", 5, "postcondition")
    {
        __result_0 = a / b
        __result_1 = a % b
        return __result_0, __result_1
    }
}
```

### Функция с именованными возвращаемыми значениями

```go
/*@
requires len(a) > 0;
ensures \result_0 <= \result_1;
*/
func MinMax(a []int) (min, max int) {
    min, max = a[0], a[0]
    for _, v := range a[1:] {
        if v < min { min = v }
        if v > max { max = v }
    }
    return
}
```

После обработки:

```go
func MinMax(a []int) (min, max int) {
    checks.CheckAt(len(a) > 0, "minmax.go", 2, "precondition")
    defer checks.CheckAt(min <= max, "minmax.go", 3, "postcondition")
    min, max = a[0], a[0]
    for _, v := range a[1:] {
        if v < min { min = v }
        if v > max { max = v }
    }
    return
}
```

> При именованных возвращаемых значениях:
> - `var` объявления не генерируются — Go уже объявил `min` и `max`
> - `\result_0` → `min`, `\result_1` → `max`
> - Голый `return` остаётся без изменений

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
    defer checks.CheckAt(*p == __old_1 && *q == __old_2, "swap.go", 2, "postcondition")
    save := *p
    *p = *q
    *q = save
}
```
