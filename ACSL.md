# ACSL Overview

В этом документе приводится краткий обзор примитивов [ACSL](https://github.com/acsl-language) и оценка возможности переиспользования этих притимивов в языке спецификаций для go.

## Логические выражения и предикаты

Ниже представлена грамматика, задающая предикаты в языке ACSL. Предикаты здесь соответствуют предикатам в классической логике первого порядка и не зависят от анализируемого языка. Поэтому множество предикатов можно оставить без изменений.

```(ebnf)
rel-op ::= == | != | <= | >= | > | <

pred ::= \true | \false
       | term (rel-op term)+                  (* comparisons (see remark) *)
       | ident ( term (, term)* )             (* predicate application *)
       | ( pred )                             (* parentheses *)
       | pred && pred                         (* conjunction *)
       | pred || pred                         (* disjunction *)
       | pred ==> pred                        (* implication *)
       | pred <==> pred                       (* equivalence *)
       | ! pred                               (* negation *)
       | pred ^^ pred                         (* exclusive or *)
       | term ? pred : pred                   (* ternary condition *)
       | pred ? pred : pred
       | \let id = term ; pred                (* local binding *)
       | \let id = pred ; pred
       | \forall binders ; pred               (* universal quantification *)
       | \exists binders ; pred               (* existential quantification *)
       | id : pred                            (* syntactic naming *)
       | string : pred                        (* syntactic naming *)
```

В предикатах с помощью ключевых слов `\forall` и `\exists` могут вводится дополнительные связанные переменные. Ниже представлена грамматика, задающая связывания. Эта грамматика почти в неизменном виде может быть использована в GSL, если заменить в ней `C-type-name` и `C-type-expr` на аналоги из Go.

```(ebnf)
binders ::= binder (, binder)*

binder ::= type-name variable-ident (, variable-ident)*

type-name ::= logic-type-name | C-type-name

type-expr ::= logic-type-name | C-type-expr

logic-type-name ::= built-in-logic-type
                  | id                         (* type identifier *)

built-in-logic-type ::= boolean | integer | real

variable-ident ::= id
                 | * variable-ident
                 | variable-ident []
                 | ( variable-ident )
```

Грамматика логических выражений представлена ниже.

```(ebnf)
literal ::= \true | \false                    (* boolean constants *)
           | integer                          (* lexical integer constants *)
           | real                             (* lexical real constants *)
           | string                           (* lexical string constants *)
           | character                        (* lexical character constants *)

bin-op ::= + | - | * | / | %
         | == | != | <= | >= | > | <
         | && | || | ^^                       (* boolean operations *)
         | << | >>
         | & | | | --> | <--> | ^             (* bitwise operations *)

unary-op ::= + | -                            (* unary plus and minus *)
           | !                                (* boolean negation *)
           | ~                                (* bitwise complementation *)
           | *                                (* pointer dereferencing *)
           | &                                (* address-of operator *)

term ::= literal                              (* literal constants *)
       | ident                                (* variables, function names *)
       | unary-op term
       | term bin-op term
       | term [ term ]                        (* array access *)
       | { term \with [ term ] = term }       (* array functional modifier *)
       | term . id                            (* structure field access *)
       | { term \with . id = term }           (* field functional modifier *)
       | term -> id
       | ( type-expr ) term                   (* cast *)
       | ident ( term (, term)* )             (* function application *)
       | ( term )                             (* parentheses *)
       | term ? term : term                   (* ternary condition *)
       | \let id = term ; term                (* local binding *)
       | sizeof ( term )
       | sizeof ( C-type-expr )
       | alignof ( C-type-expr )
       | _Alignof ( C-type-expr )             (* discouraged, equivalent to alignof *)
       | id : term                            (* syntactic naming *)
       | string : term                        (* syntactic naming *)

poly-id ::= id
ident ::= id
```

В отличие от предикатов, выражения (термы, если говорить в терминах логики первого порядка) соответствуют выражениям из языка C, поэтому предлагается внести ряд изменений в грамматику.
*   Добавить объекты в множество литералов
*   Удалить C-специфичные операторы `sizeof` и `alignof`

Язык ACSL предоставляет некоторые встроенные функции и предикаты. Например, предикаты для проверки значения типа `double` на $\pm\infty$ и NaN. Все перечисленные предикаты актуальны и для языка go.

```(ascl)
predicate \is_finite(double x);          // is a finite double
predicate \is_plus_infinity(double x);   // is equal to +infinity
predicate \is_minus_infinity(double x);  // is equal to -infinity
predicate \is_infinite(double x);        // is equal to +infinity or -infinity
predicate \is_NaN(double x);             // is a NaN double
```

Методы для округления чисел с плавающей точкой:
```
type rounding_mode = \Up | \Down | \ToZero | \NearestAway | \NearestEven;

logic float \round_float(rounding_mode m, real x);
logic double \round_double(rounding_mode m, real x);
```

В следующих частях будут описаны другие встроенные предикаты ACSL с указанием применимости их для языка спецификаций Go.

## Контракты

```
function-contract a ::= requires-clause∗ terminates-clause?
                        decreases-clause? simple-clause∗
                        named-behavior∗ completeness-clause∗

clause-kind ::= check | admit

requires-clause ::= clause-kind? requires pred ;

terminates-clause ::= terminates pred ;

decreases-clause ::= decreases term (for ident)? ;

simple-clause ::= assigns-clause | ensures-clause
                | allocation-clause | abrupt-clause

assigns-clause ::= assigns locations ;

locations ::= locations-list | \nothing

locations-list ::= location (, location)∗

location ::= tset b

ensures-clause ::= clause-kind? ensures pred ;

named-behavior ::= behavior id : behavior-body

behavior-body ::= assumes-clause∗ requires-clause∗ simple-clause∗

assumes-clause ::= assumes pred ;

completeness-clause ::= complete behaviors (id (, id)∗)? ;
                      | disjoint behaviors (id (, id)∗)? ;

term ::= \old ( term ) old value
       | \result result of a function

pred ::= \old ( pred )
```

Язык описания контрактов в требует следующих изменений:
*   Удалить `allocation-clause`, поскольку автоматическое управление памятью в go избавляет от необходимости отслеживать выделенные участки памяти

## Множества значений

```(ebnf)
tset ::= \empty                                    (* empty set *)
       | tset -> id
       | tset . id
       | * tset
       | & tset
       | tset [ tset ]
       | tset [ range ]
       | ( range )                                 (* a range as a set of integers *)
       | \union ( tset (, tset)∗ )                 (* union of location sets *)
       | \inter ( tset (, tset)∗ )                 (* intersection of location sets *)
       | tset + tset
       | ( tset )
       | { tset | binders (; pred)? }              (* set comprehension *)
       | { ( term (, term )∗ )? }                  (* explicit set *)
       | term                                      (* implicit singleton *)

pred ::= \subset ( tset , tset )                   (* set inclusion *)
       | term \in tset                             (* set membership *)

range ::= term? .. term?
```

ACSL позволяет выполнять логические операции не только с конкретными значениями, но и с множествами значений (`tset` в терминах ACSL). Каждая операция применяется ко всем элементам множества. Результатом применения такой операции является новый tset, содержащий результаты применения операции к каждому элементу исходного.

`tset` слабо зависит от языка реализации, поэтому может быть перенесён в GSL почти в неизменном виде.

## Аннотации инструкций и циклов

```(ebnf)
C-compound-statement a ::= { C-declaration∗
                             C-statement∗ assertion+ }

C-statement b ::= assertion
                  C-statement

assertion-kind ::= assert assertion
                 | clause-kind non-blocking assertion

assertion ::= /*@ assertion-kind pred ; */
            | /*@ for id (, id)∗ :
                  assertion-kind pred ; */
```

```(ebnf)
C-statement a ::= /*@ loop-annot */
                  C-iteration-statement

loop-annot b ::= loop-clause∗ loop-behavior∗
                 loop-variant?

loop-clause ::= loop-invariant | loop-assigns
              | loop-allocation

loop-invariant ::= clause-kind?
                   loop invariant pred ;

loop-assigns ::= loop assigns locations ;

loop-behavior ::= for id (, id)∗ : loop-clause+    (* annotation for behavior id *)

loop-variant ::= loop variant term ;
               | loop variant term for ident ;     (* variant for relation ident *)
```

```(ebnf)
assertion ::= /*@ clause-kind? invariant pred ; */
            | /*@ for id (, id)∗ : clause-kind? invariant pred ; */
```

```(ebnf)
term ::= \at ( term , label-id )
pred ::= \at ( pred , label-id )

label-id ::= Here | Old | Pre | Post
           | LoopEntry | LoopCurrent | Init
           | id                                  (* id is id of label in program *)
```

## Контракты инструкций

```(ebnf)
C-statement a ::= /*@ statement-contract */ C-statement

statement-contract b ::= (for id (, id)∗ :)? requires-clause∗
                         simple-clause-stmt∗ named-behavior-stmt∗
                         completeness-clause∗

simple-clause-stmt ::= simple-clause | abrupt-clause-stmt

named-behavior-stmt ::= behavior id : behavior-body-stmt

behavior-body-stmt c ::= assumes-clause∗
                         requires-clause∗ simple-clause-stmt∗
```

## Логические конструкции

TBD: эти конструкции не привязаны к C, поэтому в обзоре не рассматривались

## Указатели и память

```(ebnf)
term ::= \null
       | \base_addr one-label? ( term )
       | \block_length one-label? ( term )
       | \offset one-label? ( term )
       | \allocation one-label? ( term )

pred ::= \allocable one-label? ( term )
       | \freeable one-label? ( term )
       | \fresh two-labels? ( term, term )
       | \valid one-label? ( locations-list )
       | \valid_read one-label? ( locations-list )
       | \separated ( location , locations-list )
       | \object_pointer one-label? ( locations-list )
       | \pointer_comparable one-label? ( term , term )
       | \aligned ( term , term )

pred ::= \valid_function ( locations-list )

one-label ::= { label-id }
two-labels ::= { label-id, label-id }
```

Поскольку управление памятью в Go автоматическое, описанные выше предикаты не имеют смысла для Go.

```(ebnf)
allocation-clause ::= allocates dyn-allocation-addresses ;
                    | frees dyn-allocation-addresses ;

loop-allocation ::= loop allocates dyn-allocation-addresses ;
                  | loop frees dyn-allocation-addresses ;

dyn-allocation-addresses ::= locations
```

