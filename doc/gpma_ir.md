### Введение

GPMA (Go Program Model Analysis) использует собственный язык спецификаций GOSL (ACSL-like; пока используем чистый ACSL) для описания контрактов функций и аннотаций циклов.

#### Данная документация описывает:

1. внутреннее IR-представление спецификаций (specir);

2. структуру данных для привязки IR к Go AST (ASTData);

3. механизм сопоставления ACSL-комментариев с узлами Go-кода; TODO

4. структуру AdditionalInfo;

5. поток обработки:
Go AST -> Raw Blocks -> GOSL front -> SpecModule -> Bind -> ASTData.


### В данный момент структура framework следующая:
go-specification-lang/
    go.work
    framework/
        go.mod

        cmd/        # Утилиты, бинарники

        codegen/    # Генерация кода из AST

        data/       # Основные структуры данных, используемые фреймворком: ASTData, AdditionalInfo

        gosl/       # Реализация GOSL IR
            bind/       # Биндинг нод Go AST к элементам SpecModule
            front/      # Парсинг и создание структуры SpecModule из блоков комментариев, полученных из Go AST.
            ir/         # Реализация структур, поддерживающих GOSL (сейчас это чистый ACSL)

        parser/     # Go парсер

        plugin/     # Реализация системы плагинов

        test/       # Набор тестов, входных данных для тестов и т.д. Однако есть тесты, которые находятся и в других директориях.

    ir/         # empty
    parser/     # empty
    plugin/     # empty
 
### IR
На данный момент мы хотим поддерживать аналоги FuncContract и LoopSpec. Поэтому GOSL IR сейчас состоит из следующего

`framework/gosl/ir/contract.go` - описание контрактов функций и спецификаций циклов.

Содержит:

- `NamedPredicate` - именованный предикат (`requires name: Pred;`, `ensures bound: 0 <= \result <= n;`, `loop invariant count: R;`).
- `LocationSet` - упрощённое `assigns` / `loop assigns` (поддерживаются `\nothing` и список имён переменных).

- `FuncContract` - контракт функции:
  - `FuncName`, `RecvTypeName` - идентификация Go-функции/метода;
  - `Requires`, `Ensures` - списки предикатов;
  - `Assigns` - множество модифицируемых локаций (см. LocationSet).

- `LoopSpec` - спецификация цикла:
  - `Invariants` - `loop invariant ...;`;
  - `Assigns` - `loop assigns ...;`;
  - `Variant` - `loop variant ...;`.

- `AssertSpec` - спецификация ассерта:
  - `Assertion` - предикат после ключевого слова `assert`

---

`framework/gosl/ir/expr.go` - дерево логических выражений.

Содержит интерфейс `Expr` и набор конкретных узлов:

- `ExprLiteral` - литералы (`\true`, `\false`, целые, вещественные, строки).
- `ExprVar` - логические переменные/идентификаторы.
- `ExprUnary` - унарные операторы (`+`, `-`, `!`).
- `ExprBinary` - бинарные операторы (`+`, `-`, `*`, `/`, `%`, сравнения, `&&`, `||`, `^^`).
- `ExprCall` - вызов логической функции/предиката (`Count(a, n, v)`, `\valid_read(...)`).
- `ExprConditional` - тернарный оператор (`cond ? then : else`).
- `ExprResult` - специальное значение `\result`.
- `ExprOld` - конструкция `\old(expr)`.

Все они реализуют `Expr` и используются внутри контрактов, спецификаций циклов и логических функциях

---

`framework/gosl/ir/logic.go` - логические функции и предикаты.

Содержит:
- `LogicDeclKind` - тип логического объявления (`LogicFunction`, `LogicPredicate`, `LogicAxiom`). На данный момент будем пользоваться только LogicFunction.
- `LogicParam` - параметр логической функции/предиката (имя + логический тип).
- `LogicDecl` - логическое объявление:
  - `Name` - имя (`Count`, `Sorted`, `\is_finite` и т.п.);
  - `Kind` - функция/предикат/аксиома;
  - `Params` - параметры;
  - `ResultType` - логический тип результата (`TypeExpr`);
  - `Body` - тело (`Expr`) или `nil`, если это объявление;
  - `Span` - положение в исходнике.

Здесь находятся все логические определения, которыми затем используются в контрактах и спецификациях циклов.

---


`framework/gosl/ir/span.go` - диапазоны исходного кода.
Содержит:
- структуру `Span`, описывающую положение фрагмента спецификации в файле:
  - имя файла (опционально);
  - начальную и конечную позицию в исходном файле. 

`Span` встраивается в большинство IR-типов (`Expr`, `NamedPredicate`, `LogicDecl`, `FuncContract`, `LoopSpec`) и используется для:
- привязки ошибок/варнингов к строкам исходника,
- биндинга спецификаций к Go AST,

---

`framework/gosl/ir/types.go` - логические типы.
Содержит:
- `TypeKind` - перечисление видов логических типов:
  - `TypeBoolean` - булевы значения;
  - `TypeInteger` - математические целые;
  - `TypeReal` - вещественные;
  - `TypeCustom` - пользовательские логические типы (`rounding_mode`, `value_type` и т.п.).
- `TypeExpr` - описание логического типа:
  - `Kind` - базовый класс типа;
  - `Name` - имя для `TypeCustom`.

Эти типы используются в сигнатурах логических функций/предикатов (`LogicDecl`), параметрах (`LogicParam`) и ... (где еще могут быть использованы?)

---

`framework/gosl/ir/spec.go` - корневой модуль спецификаций. Идея этой структуры содержать в себе полный контекст спецификации в go файле:
```go
type SpecModule struct {
	LogicDecls      map[string]*LogicDecl
	FuncContracts []*FuncContract
	LoopSpecs     []*LoopSpec
}
```

### Pipeline

Примерный пайплайн на данный момент обработки Go-кода, спецификаций GOSL и построение из него IR для GPMA приведена ниже.
**.go file -> Go parser -> Comment block retrieval -> GOSL frontend -> SpecModule creation -> Binding Go AST to SpecModule -> plugin main**

#### .go file
Точка входа приложения. На данный момент приложение умеет обрабатывать по одному *.go файлу. Модули и package не поддерживаются.

---

#### Go parser
Используется стандартный Go-парсер:

```go
parser.ParseFile(fset, filename, nil, parser.ParseComments)
```

Результат:

* создаётся полное дерево `*ast.File`,
* загружаются все `ast.Node`,
* комментарии (`/* ... */`, `// ...`) включаются в AST в виде `CommentGroup`.
* На вход парсер выдает data.ASTData

---

#### Comment block retrieval

Обход AST с целью найти GOSL спецификации (на данный момент ищем ACSL аннотированные комментарии):

```
/*@
    requires ...;
    ensures ...;
*/
```

и

```
/*@
    loop invariant ...;
*/
```

Для каждого такого блока определяется (опционально). Дополнительные данные для RawSpecBlock могут помочь в последствии при прикреплении нод go AST к данным из спецификации SpecModule:

* **Text** - содержимое комментария без `/*@` и `*/`.
* **Span** - позиция блока в исходнике,
* **StartPos** - расширенная информация о позиции начала блока. (Должно заменить в будущем **Span**)
* **EndPos** - расширенная информация о позиции конца блока. (Должно заменить в будущем **Span**)
* **FuncName** и **RecvTypeName** - если это контракт функции,
* **ParentFuncName** - если это loop-spec,

Результат этапа - список `RawSpecBlock`.

---

#### GOSL frontend

Фронтенд спецификаций получает `[]RawSpecBlock` и выполняет два шага:

1. **Парсинг текста** (ANTLR-Go, грамматика ACSL):

   * разбирает предикаты, логические выражения, loop-аннотации.

2. **Преобразование в IR (`specir`)**:

   * строит выражения (`Expr`),
   * логические функции (`LogicDecl`),
   * контракты (`FuncContract`),
   * loop-spec (`LoopSpec`)

Этот этап как и сам GOSL IR по идее не должен знать ничего о Go AST.

---

#### SpecModule creation

Результат фронтенда - объект:

```go
type SpecModule struct {
    LogicDecls    map[string]*LogicDecl
    FuncContracts []*FuncContract
    LoopSpecs     []*LoopSpec
    Asserts       []*AssertSpec
}
```

`SpecModule` содержит в себе весь контекст спецификации для *.go файла. На данном этапе мы хотим держать следующие конструкции:
* контракты функций (`requires`, `ensures`, `assigns`),
* спецификации циклов (`loop invariant`, `loop assigns`, `loop variant`).
* ассерты (`assert`)
* логические определения, которые используют контракты и спецификации циклов. Это можно расширить или уменьшить в зависимости от сроков для реализации.

Это контекст спецификации. Она ничего не знает про Go AST

---

#### Binding Go AST to SpecModule

На этом шаге спецификации присоединяются к конкретным узлам кода Go:

Создаётся:

```go
type ASTData struct {
    File     *ast.File
    FileSet  *token.FileSet
    NodeInfo map[ast.Node]*AdditionalInfo
}
```

Где:

```go
type AdditionalInfo struct {
    Data         map[string]interface{} // оставил для совместимости со старым кодом
    FuncContract *specir.FuncContract   // контракты
    LoopSpec     *specir.LoopSpec       // спецификации циклов
    Asserts      *specir.AssertSpec     // ассерт

}
```

Процесс биндинга (в данный момент не реализован):

1. Контракт функции ищется по:

   * имени (`FuncName`),
   * типу получателя (`RecvTypeName`),
   * позиции комментария (`Span`).

2. LoopSpec ищется по:

   * позиции комментария,
   * родительской функции (`ParentFuncName`),
   * узлу цикла (`*ast.ForStmt` / `*ast.RangeStmt`).

3. Найденные связи записываются в:

```go
NodeInfo[node] = &AdditionalInfo{ ... }
```

После этого любое средство анализа может получить спецификацию для нужной Go Ast ноды, обращаясь к спецификации для ноды через `AdditionalInfo`. А если потребуется узнать больше контекста спецификации, то можно обратиться к ней через `ASTData.Spec`.

---

#### plugin main

Последний этап: запуск пользовательских плагинов и анализаторов.
На вход они получают:

* `ASTData` - Go AST + SpecModule. 

Плагины могут:
* генерировать динамические проверки
* выполнять статическую верификацию,
* генерировать диагностику и отчёты.
* ...