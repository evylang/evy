# Evy language specification

Evy is a [statically typed], [garbage collected],
[procedural] programming language. Its main design goal is to help
learn programming. Evy aims for simplicity and directness in its
tooling and syntax. Several features typical of modern programming
languages are purposefully left out.

To get an intuitive understanding of Evy, you can either look at its
[syntax by example](syntax_by_example.md) or read through the
[built-ins documentation](builtins.md).

[statically typed]: https://developer.mozilla.org/en-US/docs/Glossary/Static_typing
[garbage collected]: https://en.wikipedia.org/wiki/Garbage_collection_(computer_science)
[procedural]: https://en.wikipedia.org/wiki/Procedural_programming

#### Table of Contents

<!-- gen:toc -->

1. [**Syntax Grammar**](#syntax-grammar)  
   [WSN Syntax Grammar](#wsn-syntax-grammar), [Evy Syntax Grammar](#evy-syntax-grammar)
2. [**Comments**](#comments)
3. [**Types**](#types)
4. [**Variables and Declarations**](#variables-and-declarations)
5. [**Zero Values**](#zero-values)
6. [**Assignments**](#assignments)
7. [**Copy and Reference**](#copy-and-reference)
8. [**Variable Names**](#variable-names)
9. [**Scope**](#scope)
10. [**Strings**](#strings)
11. [**Arrays**](#arrays)
12. [**Maps**](#maps)
13. [**Index and Slice**](#index-and-slice)
14. [**Operators and Expressions**](#operators-and-expressions)  
    [Arithmetic and Concatenation Operators](#arithmetic-and-concatenation-operators), [Logical Operators](#logical-operators), [Comparison Operators](#comparison-operators), [Unary Operators](#unary-operators)
15. [**Precedence**](#precedence)
16. [**Statements**](#statements)
17. [**Whitespace**](#whitespace)  
    [Vertical Whitespace](#vertical-whitespace)
18. [**Horizontal Whitespace**](#horizontal-whitespace)
19. [**Functions**](#functions)  
    [Bare Returns](#bare-returns), [Function Names](#function-names), [Anonymous Parameters](#anonymous-parameters), [Variadic Functions](#variadic-functions)
20. [**Break and Return**](#break-and-return)
21. [**Typeof**](#typeof)
22. [**Type Assertion**](#type-assertion)
23. [**Assignability**](#assignability)  
    [Assignability of variable values](#assignability-of-variable-values), [Assignability of constant values](#assignability-of-constant-values), [Assignability of empty composite literals](#assignability-of-empty-composite-literals)
24. [**Run-time Panics and Recoverable Errors**](#run-time-panics-and-recoverable-errors)
25. [**Execution Model and Event Handlers**](#execution-model-and-event-handlers)
26. [**Runtimes**](#runtimes)

<!-- genend:toc -->

## Syntax Grammar

The Evy syntax grammar is a [WSN] grammar, which is a formal set of
rules that define how Evy programs are written. The Evy interpreter uses
the syntax grammar to parse Evy source code, which means that it checks
that the code follows the rules of the grammar.

[WSN]: https://en.wikipedia.org/wiki/Wirth_syntax_notation

### WSN Syntax Grammar

Evy's syntax is specified using a WSN grammar, a variant of
[EBNF] grammars, borrowing concepts from the [Go Programming Language
Specification].

**Productions** are the top-level elements of a WSN grammar. For example,
the production `OPERATOR = "+" | "-" | "*" | "/" .` specifies that an
operator can be one of the characters `+`, `-`, `*`, or `/`.

A production consists of an **expression** assigned to an **identifier** or
production name. Each production is terminated by a period `.`. An
expression consists of **terms** and the following **operators** in
increasing precedence:

- **Alternation:** `|` stands for "or". For example, `a | b` stands for `a` or `b`.
- **Grouping:** `()` stands for grouping. For example, `(a|b)c` stands for `ac` or `bc`.
- **Optionality:** `[]` stands for optionality. For example, `[a]b` stands for `ab` or `b`.
- **Repetition:** `{}` stands for repetition. For example, `{a}` stands for the empty string, `a`, `aa`, `aaa`, ...".

`a … b` stands for a range of single characters from `a` to `b`,
inclusive.

Here is a WSN defining itself:

    syntax     = { production } .
    production = identifier "=" expression "." .
    expression = terms { "|" terms } .
    terms      = term { term } .
    term       = identifier |
                 literal |
                 "[" expression "]" |
                 "(" expression ")" |
                 "{" expression "}" .
    identifier = LETTER { LETTER } .
    literal    = """ CHARACTER { CHARACTER } """ . /* """" is a literal `"` */
    LETTER     = "a" … "z" | "A" … "Z" | "_" .
    CHARACTER  = /* an arbitrary Unicode code point */ .

**Terminals** are the leaves in the grammar that cannot be expanded
further. By convention, terminals are identified by production names
in uppercase.

**Non-terminals**, on the other hand, can be expanded into other
productions. This means that they can be replaced by a more complex
expression. By convention, non-terminals are identified by production
names in lowercase.

**Literals** or lexical tokens are enclosed in double quotes `""`.
Comments are fenced by `/*` … `*/`.

There are two special fencing tokens in Evy's grammar related to
**horizontal whitespace**, `<-` … `->` and `<+` … `+>`. `<-` … `->` means
no horizontal whitespace is allowed between the terminals of the
enclosed expression, e.g. `3+5` inside `<-` … `->` is allowed, but
`3 + 5` is not. The fencing tokens `<+` … `+>` are the default and mean
horizontal whitespace is allowed (again) between terminals.

See the section on [whitespace](#whitespace) for further details.

[EBNF]: https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form
[Go Programming Language Specification]: https://go.dev/ref/spec

### Evy Syntax Grammar

Evy source code is UTF-8 encoded, which means that it can contain any
Unicode character. The NUL character `U+0000` is not allowed, as it is
a special character that is used during compilation.

The `WS` abbreviation in the grammar comments below refers to horizontal
whitespace, which is any combination of spaces and tabs. The following
listing contains the complete syntax grammar for Evy.

    program    = { statement | func | event_handler | nl } eof .
    statements = { nl } statement { statement | nl } .
    statement  = typed_decl_stmt | inferred_decl_stmt |
                 assign_stmt |
                 func_call_stmt |
                 return_stmt | break_stmt |
                 if_stmt | while_stmt | for_stmt .

    /* --- Functions and Event handlers ---- */
    func            = "func" ident func_signature nl
                          statements
                      "end" nl .
    func_signature  = [ ":" type ] params .
    params          = { typed_decl } | variadic_param .
    variadic_param  = typed_decl "..." .

    event_handler   = "on" ident params nl
                          statements
                      "end" nl .

    /* --- Control flow --- */
    if_stmt         = "if" toplevel_expr nl
                            statements
                      { "else" "if" toplevel_expr nl
                            statements }
                      [ "else" nl
                            statements ]
                      "end" nl .

    while_stmt      = "while" toplevel_expr nl
                          statements
                      "end" nl .

    for_stmt   = "for" range nl
                      statements
                 "end" nl .
    range      = [ ident ":=" ] "range" range_args .
    range_args = <- expr -> [ <- expr -> [ <- expr -> ] ] .

    return_stmt = "return" [ toplevel_expr ]  nl .
    break_stmt  = "break" nl .

    /* --- Statement ---- */
    assign_stmt        = target "=" toplevel_expr nl .
    typed_decl_stmt    = typed_decl nl .
    inferred_decl_stmt = ident ":=" toplevel_expr nl .
    func_call_stmt     = func_call nl .

    /* --- Assignment --- */
    target         = <- ident | index_expr | dot_expr -> . /* no WS before `[` and around `.` */
    ident          = LETTER { LETTER | UNICODE_DIGIT } .
    index_expr     = target "[" <+ toplevel_expr +> "]" .
    dot_expr       = target "." ident .

    /* --- Type --- */
    typed_decl     = ident ":" type .
    type           = BASIC_TYPE | DYNAMIC_TYPE | composite_type .
    BASIC_TYPE     = "num" | "string" | "bool" .
    DYNAMIC_TYPE   = "any" .
    composite_type = array_type | map_type .
    array_type     = "[" "]" type .
    map_type       = "{" "}" type .

    /* --- Expressions --- */
    toplevel_expr = func_call | expr .

    func_call = ident args .
    args      = { tight_expr } . /* no WS within single arg, WS is arg separator */

    tight_expr = <- expr -> . /* no WS allowed unless within `(…)`, `[…]`, or `{…}` */
    expr       = operand | unary_expr | binary_expr .

    operand    = literal | target | slice | type_assertion | group_expr .
    group_expr = "(" <+ toplevel_expr +> ")" . /* WS can be used freely within `(…)` */
    type_assertion = <- target ".(" -> type ")" . /* no WS around `.` */

    unary_expr = <- UNARY_OP -> expr .  /* no WS after UNARY_OP */
    UNARY_OP   = "-" | "!" .

    binary_expr   = expr binary_op expr .
    binary_op     = LOGICAL_OP | COMPARISON_OP | ADD_OP | MUL_OP .
    LOGICAL_OP    = "or" | "and" .
    COMPARISON_OP = "==" | "!=" | "<" | "<=" | ">" | ">=" .
    ADD_OP        = "+" | "-" .
    MUL_OP        = "*" | "/" | "%" .

    /* --- Slice and Literals --- */
    slice       = <- target "[" slice_expr "]" -> .
    slice_expr  = <+ [expr] ":" [expr] +> .
    literal     = num_lit | string_lit | BOOL_CONST | array_lit | map_lit .
    num_lit     = DECIMAL_DIGIT { DECIMAL_DIGIT } |
                  DECIMAL_DIGIT { DECIMAL_DIGIT } "." { DECIMAL_DIGIT } .
    string_lit  = """ { UNICODE_CHAR } """ .
    BOOL_CONST  = "true" | "false" .
    array_lit   = "[" <+ array_elems +> "]" . /* WS can be used freely within `[…]`, but not inside the elements */
    array_elems = { tight_expr [nl] } .
    map_lit     = "{" <+ map_elems +> "}" . /* WS can be used freely within `{…}`, but not inside the values */
    map_elems   = { ident ":" tight_expr [nl] } .
    nl          = [ comment ] NL .
    eof         = [ comment ] EOF .
    comment     = "//" { UNICODE_CHAR } .

    /* --- Terminals --- */
    LETTER         = UNICODE_LETTER | "_" .
    UNICODE_LETTER = /* a Unicode code point categorized as "Letter" (category L) */ .
    UNICODE_DIGIT  = /* a Unicode code point categorized as "Number, decimal digit" */ .
    UNICODE_CHAR   = /* an arbitrary Unicode code point except newline */ .
    DECIMAL_DIGIT  = "0" … "9" .
    NL             = "\n"  . /* end of file */
    EOF            = "" . /* end of file */

## Comments

There is only one type of comment, the line comment which starts with
`//` and stops at the end of the line. Line comments cannot start
inside string literals.

## Types

Evy has a static **type system** where the types of variables, parameters
and expressions are known at parse time. This means that the parser
can check for type errors before the program is run.

There are three **basic types**: `num`, `string` and `bool` as well as two
**composite types**: arrays `[]` and maps `{}`. The **dynamic** type `any`
can hold any of the previously listed types.

Composite types can nest further composite types, for example
`[]{}string` is an array of maps with string values.

A `bool` value is either `true` or `false`.

A number value can be expressed as integer `1234` or decimal `56.78`.
Internally a number is represented as a [double-precision floating-point number]
according to the IEEE-754 64-bit floating point standard.

[double-precision floating-point number]: https://en.wikipedia.org/wiki/Double-precision_floating-point_format

## Variables and Declarations

Variables hold values of a specific type. They must be **declared**
before they can be used. A declared variable must be used at least once,
meaning it must be used in the right hand side of an assignment or
passed as an argument to a function call. There are two types of
variable declarations: inferred declarations and typed declarations.

**Inferred declarations** do not specify the type of the variable
explicitly. The type of the variable is inferred from the value that
it is initialized to. For example, the following code declares a
variable `n` and initializes it to the value `1`. The type of `n` is
inferred to be `num`.

    n := 1

**Typed declarations** explicitly specify the type of the variable. The
variable is initialized to the type's zero value. For example, the
following code declares a variable `s` of type `string` and
initializes it to the empty string `""`.

    s:string

`arr := []` infers an array of type any, `[]any`. `map := {}` infers a
map of type any, `{}any`. The strictest possible type is inferred for
composite types:

```evy
arr1 := [1 2 3] // []num
arr2 := [1] + [] // []num
print 1 (typeof arr1) (typeof arr2)

arr3 := [1 "a"] // []any
arr4 := [[1] ["a"]] // [][]any
arr5 := [] // []any
print 2 (typeof arr3) (typeof arr4) (typeof arr5)

map1 := {} // {}any
map2 := {age:10} // {}num
print 3 (typeof map1) (typeof map2)
```

The [`typeof`](#typeof) function returns the type as string
representation, so the code above outputs:

```evy:output
1 []num []num
2 []any [][]any []any
3 {}any {}num
```

## Zero Values

Variables declared via typed declaration are initialized to the zero
value of their type:

- Number: `0`
- String: `""`
- Boolean: `false`
- Any: `false`
- Array: `[]`
- Map: `{}`

The empty array becomes `[]any` in inferred declarations. Otherwise the
empty array literal assumes the array type `[]TYPE` required by the
assigned variable or parameter. For example, the following code

```evy
arr:[]num
print 1 arr (typeof arr)
arr = []
print 2 arr (typeof arr)
print 3 (typeof [])
```

generates the output

```evy:output
1 [] []num
2 [] []num
3 []any
```

Similarly, the empty map literal becomes `{}any` in inferred
declarations. Otherwise the empty map literal assumes the map type
`{}TYPE` required.

## Assignments

Assignments are defined by an equal sign `=`. The left-hand side of the `=`
must contain an **assignment target**, a variable, an indexed array, or a map
field. The assignment target must be declared before the assignment, either
implicitly via type inference or explicitly via a type declaration. It can
also be a parameter of a function or event handler definition.
[Assignability](#assignability) provides rules on which value types can be
assigned to which target types.

For example, the following code declares a string variable named `s` and
initializes it to the value `"a"` through inference. Then, it assigns
the value `"b"` to `s`. Finally, it tries to assign the value `100` to
`s`, which will cause a parse error because `s` is a string
variable and `100` is a number.

```evy
s := "a"
print 1 s
s = "b"
print 2 s
// s = 100 // parse error, wrong type
```

Output

```evy:output
1 a
2 b
```

## Copy and Reference

When a variable of a basic type `num`, `string`, or `bool` is the value
of an assignment, a **copy** of its value is made. A copy is also made
when a variable of a basic type is used as the value in an inferred
declaration or passed as an argument to a function.

```evy
a := 1
b := a
print a b
a = 2 // `b` keeps its initial value
print a b
```

generates the output

```evy:output
1 1
2 1
```

By contrast, composite types - maps and arrays - are **passed by
reference** and no copy is made. Modifying the contents of an array
referenced by one variable also modifies the contents of the array
referenced by another variable. This is also true for argument passing
and inferred declarations:

```evy
a := [1]
b := a
print a b
a[0] = 2 // the value of `b` is also updated
print a b
```

generates the output

```evy:output
[1] [1]
[2] [2]
```

For the dynamic type `any`, a copy is made if the value is a basic type.
The variable is passed by reference if the value is a composite type.

## Variable Names

Variable names in Evy must start with a letter or underscore, and can
contain any combination of letters, numbers, and underscores. They
cannot be the same as keywords, such as `if`, `func`, or any built-in
or defined function names.

## Scope

**Scope** refers to the visibility of a variable or function.

Functions can only be defined at the top level of the program, known
as **global scope**. A function does not have to be defined before it
can be called; it can also be defined afterwards. This allows for
[mutual recursion], where function `a` calls function `b` and function
`b` calls function `a`.

Variables, by contrast, must be declared and given an unchangeable type
before they can be used. Variables can be declared at the top level of
the program, at **global scope**, or within a block-statement, at _block
scope_.

A **block-statement** is a block of statements that ends with the keyword
`end`. A function's parameter declaration and the function body
following the line starting with `func` is a block-statement. The
statements between `if` and `else` are a block. The statements between
`while`/`for`/`else` and `end` are a block. Blocks can be nested within
other blocks.

A variable declared inside a block only exists until the end of the
block. It cannot be used outside the block.

Variable names in an inner block can **shadow** or override the same
variable name from an outer block, which makes the variable of the
outer block inaccessible to the inner block. However, when the inner
block is finished, the variable from the outer block is restored and
unchanged:

```evy
x := "outer"
print 1 x
for range 1
    x := true
    print 2 x
end
print 3 x
```

This program will print

```evy:output
1 outer
2 true
3 outer
```

[mutual recursion]: https://en.wikipedia.org/wiki/Mutual_recursion

## Strings

A **string** is a sequence of [Unicode code points]. Unicode is a
standard that defines a unique code point for every character in every
language. This means that a string can contain characters from any
language, including English, French, Spanish, Chinese, Japanese, and
Korean.

A **string literal** is a sequence of characters enclosed by double
quotes. The characters in a string literal are interpreted as Unicode
code points. This means that a string literal can contain any character
that has a Unicode code point, including letters, numbers, punctuation
marks, and emojis.

The example code `str := "Hallöchen Welt 👋🌍"` defines a string variable
`str` and initializes it with a string literal that contains the German
words "Hallöchen Welt" and the emojis "👋🌍".

The `len str` function returns the number of Unicode code points, or
**characters**, in the string. The loop `for ch := range str` iterates
over all characters of the string. Individual characters of a string can
be read by index, starting at `0`. Strings can be concatenated with the
`+` operator.

The backslash character `\` can be used to represent special characters
in strings. For example, the `\t` escape sequence represents a tab
character, and the `\n` escape sequence represents a newline character.
Quotes in string literals must also be escaped with backslashes. To
print a backslash character, use `\\`.

For example the following code

```evy
str := "hello"
str = str + ", " + str // hello, hello
str = "H" + str[1:] // Hello, hello
str = "She said, \"" + str + "!\""
print str
```

outputs

```evy:output
She said, "Hello, hello!"
```

[Unicode code points]: https://en.wikipedia.org/wiki/Unicode

## Arrays

**Arrays** are collections of elements that have the same type. They are
declared with brackets `[]`, and the elements are separated by a space.
For example, the following code declares two arrays of numbers

```evy
arr1 := [1 2 3]
arr2:[]num
print arr1 arr2
```

The output is

```evy:output
[1 2 3] []
```

Arrays can also be nested, meaning that they can contain other arrays or
maps. For example, the following code declares an array of maps of
strings `arr:[]{}string`.

An array composed of different types becomes an array of type any,
`[]any`, for example

```evy
arr := ["abc" 123] // []any
print "Type of arr:" (typeof arr)
```

outputs

```evy:output
Type of arr: []any
```

The function `len arr` returns the length of the array, which is the
number of elements in the array. The loop `for el := range arr`
iterates over all elements of the array in order. Arrays can be
concatenated with the `+` operator, for example `arr2 := arr + arr`.

The elements of an array can be accessed via index starting at `0`. In
the example `arr := ["abc" 123]` the first element in the array
`arr[0]` is `"abc"`.

The empty array becomes `[]any` in inferred declarations, otherwise the
empty array literal assumes the array type required by the assigned
variable or parameter. `arr:[]any` and `arr := []` are equivalent.

In order to distinguish between array literals and array indices, there
cannot be any whitespace between array variable and index. For example,
the following code

```evy
arr := ["a" "b"]
print 1 arr[1] // index
print 2 arr [1] // literal
arr[0] = "A"
print 3 arr
// arr [1] = "B" // whitespace before `[` is invalid
```

outputs

```evy:output
1 b
2 [a b] [1]
3 [A b]
```

## Maps

**Maps** are key-value stores, where the values can be looked up by their
key, for example `map := { key1:"value1" key2:"value2" }`.

Map values can be accessed with the dot expression, for example
`map.key1`. If maps are accessed via the dot expression the key must
match the grammars `ident` production. Map values can also be accessed
with an index expression which allows for evaluation, non-ident keys
and variable usage. For example the following code

```evy
m := {letters:"abc"}
print 1 m.letters
print 2 m["letters"]

key := "German letters"
m[key] = "äöü"
print 3 m[key]
print 4 m["German letters"]
```

outputs

```evy:output
1 abc
2 abc
3 äöü
4 äöü
```

The `has` function tests for the existence of a key in a map. The
following code

```evy
m := {letters:"abc"}
print 1 (has m "letters")
print 2 (has m "digits")
```

outputs

```evy:output
1 true
2 false
```

The `del` function deletes a key from a map if it exists and does
nothing if the key does not exist. The following code

```evy
m := {letters:"abc"}
del m "letters"
print m
```

outputs

```evy:output
{}
```

The loop `for key := range map` iterates over all map keys. It is safe
to delete values from the map with the built-in function `del` while
iterating. The keys are iterated in the order in which they are
inserted. Any values inserted into the map during the iteration will
not be included in the iteration.

The function `len m` returns the number of key-value pairs in the map.

The empty map literal becomes `{}any` in inferred declarations,
otherwise the empty map literal assumes the type required by the map
type of the assigned variable or parameter. `m:{}any` and `m := {}` are
equivalent.

No whitespace is allowed around the dot expression `.`, and before the
index expression `[`.

## Index and Slice

An array or string **index** in Evy is a number that is used to access a
specific element of an array or character of a string. Array indices
start at `0`, so the first element of an array is `arr[0]`. A negative
index `-i` is a shorthand for `(len arr) - i`, so `arr[-1]` refers to the
last element of arr.

For example, the following code

```evy
arr := ["a" "b" "c"]
print 1 arr[0]
print 2 arr[-1]
```

will print the first and last elements of the array

```evy:output
1 a
2 c
```

A **slice** is a way to access portions of an array or a string. It is a
substring or subarray that is copied from the original array or string.
The slice expression `arr[start:end]` copies a substring or subarray
starting with the value at index `arr[start]`. The length of the slice
is `end-start`. The end index `arr[end]` is not included in the slice.
If `start` is left out, it defaults to `0`. If `end` is left out, it
defaults to `len arr`. For example, the following code

```evy
s := "abcd"
print 1 s[1:3]
print 2 s[:2]
print 3 s[2:]
print 4 s[:]
print 5 s[:-1]
```

outputs

```evy:output
1 bc
2 ab
3 cd
4 abcd
5 abc
```

If you try to access an element of an array or string that is out of
bounds, a [runtime panic](#run-time-panics-and-recoverable-errors) will occur. Slice
expressions must not be preceded by whitespace before the `[` character,
just like indexing an array or string. For more details, see the section
on [whitespace](#whitespace).

## Operators and Expressions

**Operators** are special symbols or identifiers that combine the values of
their operands into a single value. **Operands** are the variables or
literal values that the operator acts on. The combination of operands
and operators is called expression. An **expression** is a combination
of literal values, operators, variables, and further nested expressions
that evaluates to a single value.

In Evy, there are two types of operators: unary operators and binary
operators:

- **Unary operators** act on a single operand. For example, the unary operator `-` negates the value of its operand.
- **Binary operators** act on two operands. For example, the binary operator `+` adds the two operands together.

Operators can be combined to form larger expressions, for example, the
expression `-delta + 3` would first negate the value of the variable `delta` and
then add literal number `3` to the result.

Binary expressions can only be evaluated if the operands are of the
same type. For example, you cannot add a string to a number. There is
no automated type conversion of operands.

There are a variety of binary operators: arithmetic, concatenation,
logical, and comparison operators.

| Operator            | Operands  | Result   | Description   |
| ------------------- | --------- | -------- | ------------- |
| `+` `-` `*` `/` `%` | `num`     | `num`    | arithmetic    |
| `+`                 | `string`  | `string` | concatenation |
| `+`                 | array     | array    | concatenation |
| `and` `or`          | `bool`    | `bool`   | logical       |
| `<` `<=` `>` `>=`   | `num`     | `bool`   | comparison    |
| `<` `<=` `>` `>=`   | `string`  | `bool`   | comparison    |
| `==` `!=`           | all types | `bool`   | comparison    |

### Arithmetic and Concatenation Operators

The **arithmetic operators** `+`, `-`, `*`, `/`, and `%` stand for addition,
subtraction, multiplication, division, and the [modulo operator]. The
symbol `+` can also be used to concatenate strings and arrays.

The **modulo operator** `%`, also known as the remainder operator, returns the
remainder of a division operation. For example, `10 % 3` results in
`1`, because `10` divided by `3` has a remainder of `1`.

The **concatenation operator** `+`, combines two strings or two arrays
together. For example, `"fire"` + `"engine"` combines into the string
`"fireengine"`.

[modulo operator]: https://en.wikipedia.org/wiki/Modulo_operation

### Logical Operators

The **logical operators** `and` and `or` are used to perform
[logical conjunction] and [logical disjunction]. They are used to
perform logical operations on boolean values with type `bool`.

The `and` operator evaluates to `true` if both operands are `true`. The
`or` operator evaluates to `true` if either operand is `true`.

These operators perform [short-circuit evaluation], which means that the
right-hand side of the operator is not evaluated if the result of the
operation can be determined from the left-hand side alone.

For example, the expression `false and true` evaluates to `false`
because the first operand is `false`. The second operand does not need
to be evaluated because the result of the expression is already known
to be `false`.

[logical conjunction]: https://en.wikipedia.org/wiki/Truth_table#Logical_conjunction_(AND)
[logical disjunction]: https://en.wikipedia.org/wiki/Truth_table#Logical_disjunction_(OR)
[short-circuit evaluation]: https://en.wikipedia.org/wiki/Short-circuit_evaluation

### Comparison Operators

The **comparison operators** `<` `<=` `>` `>=` stand for less, less or equal,
greater, greater or equal. Their operands may be `num` or `string`
values. For `string` types [lexicographical comparison] is used.

The **comparison operators** `==` and `!=` compare two operands of the
same type for equality and inequality. The operands of these operators
can be basic types, such as numbers and strings, or composite types,
such as arrays and maps. The result of a comparison operation is the
boolean value `true` or `false`.

[lexicographical comparison]: https://en.wikipedia.org/wiki/Lexicographic_order

### Unary Operators

**Unary operators** are operators that operate on a single operand. In
Evy, there are two unary operators: `-` and `!`.

- The unary operator `-` negates the value of a numeric operand. For example, `-delta` negates the value of `delta`.
- The unary operator `!` performs [logical negation] on a boolean operand. For example, `!true` evaluates to `false`.

Unary operators must not be immediately followed by whitespace. For example, the
`-delta` is valid, but `- delta` is not.

The following sample illustrates the care needed with operators and
whitespace

```evy
a := 10
b := 3
print 1 a-b
print 2 (a - b)
print 3 a -b
// print a - b // parse error
```

Output:

```evy:output
1 7
2 7
3 10 -3
```

For more information about whitespace, see the
[whitespace](#whitespace) section.

[logical negation]: https://en.wikipedia.org/wiki/Truth_table#Logical_negation

## Precedence

Operators in Evy are evaluated in a specific order, called **precedence**.
The order of precedence is as follows:

1. Indexing, dot notation and grouped expressions: `a[i]` `a.b` `(` … `)`
2. Unary operators: `-delta` `!true`
3. Binary operators
   1. Multiplication, division, and modulo: `*` `/` `%`
   2. Addition and subtraction: `+` `-`
   3. Comparison operators: `<` `<=` `>` `>=`
   4. Equality operators: `==` `!=`
   5. Logical conjunction: `and`
   6. Logical disjunction: `or`

Operators of the same precedence are evaluated from left to right. For
example, the expression `a[i] - 5 * 2` will be evaluated as follows:

- The index `a[i]` will be evaluated first.
- The multiplication `5 * 2` will be evaluated next.
- The subtraction `- 5 * 2` will be evaluated last.

If you want to change the order of precedence, you can use parentheses
to group expressions. For example, the expression `(a[i] - 5) * 2` will
be evaluated as follows:

- The expression `a[i] - 5` will be evaluated first.
- The multiplication `* 2` will be evaluated next.

## Statements

A **statement** is a unit of code that performs an action. Statements are
the building blocks of programs, and they can be used to control the
flow of execution.

Statements can be divided into two categories: block statements and basic statements.

- **Basic statements** are statements that cannot be broken down into further statements.
- **Block statements** are statements that contain further statements.

There are 5 types of block statements in Evy:

- Function definition
- Event Handler definitions
- If statements
- For statements
- While statements

There are 5 types of basic statements in Evy:

- Variable declaration statement
- Assignment statement
- Function call statement
- Return statement
- Break statement

Not all statements are allowed in all contexts. For example, a return
statement may only be used within a function definition.

## Whitespace

Whitespace in Evy is used to separate different parts of a program.
There are two types of whitespace in Evy: vertical whitespace and
horizontal whitespace.

### Vertical Whitespace

Vertical whitespace is a sequence of one or more newline characters that
can optionally contain comments. It is used to terminate or end basic
statements in Evy. A basic statement is a statement that cannot be
broken up into smaller statements, such as a variable declaration,
an assignment or a function call.

Evy does not allow multiple statements on a single line. For example,
the following code is invalid because it contains two statements, a
declaration and a function call, on one line:

`x := 1      print x`

It is also not possible to break up a single basic statement over more
than one line. For example, the following code is invalid because the
arithmetic expression `1 + 2` is split over two lines:

```
x := 1 +
     2
```

The rule that basic statements cannot be split across multiple lines has
one exception: Array literals and map literals can be broken up over
multiple lines, as long as each line is a complete expression. For
example, the following code is valid because it is a declaration with a
multiline map literal:

```evy
person := {
    name:"Jane Goddall"
    born:1934
}
print person
```

## Horizontal Whitespace

**Horizontal whitespace** is a sequence of tabs or spaces that is used to
separate elements in lists. Lists include the argument list to a
function call, the element list of an array literal, and the value in
the key-value pairs of a map literal. However, horizontal whitespace
is not allowed _within_ this list elements.

Horizontal whitespace is not allowed around the dot expression `.` or
before the opening brace `[` in an index expression or slice
expression. However, it is allowed _within_ the grouping expression,
index expression, and slice expression, even if the expression is an
element of a list such as an argument to a function call.

Assignments, inferred variable declarations, return statements and the
the expression inside an index expression `[ … ]` _can_ have whitespace
around their binary operators. The whitespace around the operators is
optional, but it is often used to improve the readability of the code,
for example:

```evy
x := 5 + 3
x = 7 - 2

arr := [1 2 3]
arr[3 - 2] = 10

func fn:num
    return 7 + 1
end

print x arr (fn)
```

More formally, horizontal whitespace `WS` between tokens or terminals as
defined in the grammar is ignored and can be used freely with the
addition of the following rules:

1. `WS` is not allowed around dot `.` in dot expressions.
2. `WS` is not allowed before the `[` in index or slice expressions.
3. `WS` is not allowed following the unary operators `-` and `!`.
4. `WS` is not allowed within arguments to a function call
5. `WS` is not allowed within elements of an array literal
6. `WS` is not allowed within the values of a map literal's key-value pairs.
7. `WS` is allowed within any grouping expression `(` … `)`.
8. `WS` is allowed within an index expression `[` … `]`.
9. `WS` is allowed within a slice expression `[` … `:` … `]`.

Here are some examples of incorrect uses of horizontal whitespace, along
with their correct uses.

Invalid:

```
print - 5
len "a" + "b"

arr := [1 + 1]
arr [0] = 3 + 2
print 2 + arr [0]

map := {address: "10 Downing " + "Street"}
map.  address = "221B Baker Street"

print len map
```

Valid:

```evy
print -5
len "a"+"b"

arr := [1+1]
arr[0] = 3 + 2
print 2+arr[0]

map := {address:"10 Downing "+"Street"}
map.address = "221B Baker Street"

print (len map)
```

## Functions

**Functions** are blocks of code that are used to perform a specific task.
They are often used to encapsulate code that is used repeatedly, so
that it can be called from different parts of a program.

A **function definition** binds an identifier, the function name, to a
function. As part of the function definition, the **function signature**
declares the number, order and types of input parameters as well as the
result or return type of the function. If the return type is left out,
the function does not return a value.

For example, the following code defines a function called `validate`
that takes two parameters, `s` and `maxl`, and returns a boolean
result. The `s` parameter is of type `string` and the `maxl`
parameter is of type `num`. The return type of the function is `bool`.

```evy
func validate:bool s:string maxl:num
    return (len s) <= maxl
end
```

### Bare Returns

**Bare returns** are return statements without values. They can be used in
functions without result type. For example, the following code defines a
function called `reverse` that takes a string array as an argument and
does not return a value. The return statement in the if statement simply
exits the function early.

```evy
func reverse arr:[]string
    if arr == []
        return
    end
    // ...
end
```

Function calls used as arguments to other function calls must be
parenthesized to avoid ambiguity, for example:

```evy
print "length of abc:" (len "abc")
```

Output

```evy:out
length of abc: 3
```

Function names must be unique within an Evy program. This means that no
two functions can have the same name. Function names also cannot be the
same as a variable name.

### Function Names

Function names in Evy must start with a letter or underscore, and can
contain any combination of letters, numbers, and underscores. They
cannot be the same as keywords, such as `if`, `func`, or any built-in
or other defined function names.

### Anonymous Parameters

The anonymous parameter `_` is a special parameter in Evy that can be
used as a placeholder for a named parameter. It can be used for
multiple parameters in a single function, but it cannot be read. For
example, the following code defines an event handler for the pointer down
event that only uses the `y` parameter:

```evy
on down _:num y:num
    print "y:" (round y)
end
```

### Variadic Functions

**Variadic functions** in Evy are functions that can take zero or more
arguments of a specific type. The type of the variadic parameter is an
array with the element type of the parameter. The length of the array
is the number of arguments passed to the function.

For example, the following code defines a variadic function called
`quote` that can take any number of arguments of any type

```evy
func quote args:any...
    words:[]string
    for arg := range args
        word := sprintf "«%v»" arg
        words = words + [word]
    end
    print (join words " ")
end

quote "Life, universe and everything?" 42
```

Output

```evy:output
«Life, universe and everything?» «42»
```

Unlike other languages, arrays cannot be turned into variadic arguments
in Evy. The call arguments must be listed individually.

## Break and Return

`break` and `return` are **terminating statements** in Evy. They interrupt
the regular flow of control.

- `break` is used to exit from the innermost loop body. This means that
  it will skip the rest of the loop body and continue with the next
  statement after the loop.
- `return` is used to exit from a function. It can be followed by an
  expression whose value is returned by the function call.

For example, the following code shows how the `break` statement can be
used to exit from a loop:

```evy
for x := range 2
    y := 0
    while y < 10
        if y == 2
            print "break" y
            break
        end
        print "no break" y
        y = y + 1
    end
    print "x" x "y" y
    print
end
```

This code will print the following output:

```evy:output
no break 0
no break 1
break 2
x 0 y 2

no break 0
no break 1
break 2
x 1 y 2

```

As you can see, the `break` statement causes the loop to exit when the
value of `y` is equal to 2. The next statement after the loop is then
executed. Note how `break` only exits the innermost loop.

The following code shows how the return statement can be used to exit
from a function:

```evy
func foo:string
    if (rand1) < 0.7
        return "bar"
    else
        return "baz"
    end
end
```

This code will return the value of `"bar"` 70% of the time and `"baz"`
otherwise. The names [foo, bar, and baz] are common placeholder names
used in code.

[foo, bar, and baz]: https://en.wikipedia.org/wiki/Foobar

## Typeof

The `typeof` function returns the concrete type of a value held by a
variable as a string. It returns a string that is the same as the type
in an Evy program, such as `"num"`, `"bool"`, `"string"`, `"[]num"`,
`"{}[]any"`, etc. It is particularly useful to determine the concrete
type of an `any` variable together with type assertions.

Here is an example of how the `typeof` function works

```evy
print (typeof "abc")
print (typeof true)
print

arr := ["abc" 1]
print (typeof arr)
print (typeof arr[0])
print (typeof arr[1])
```

The output of this code is

```evy:output
string
bool

[]any
string
num
```

Empty composite literals, `[]` and `{}`, can be assigned to variables or
parameters of any subtype, such as `[]string` or `{}num`. This is
because empty composite literals are untyped, meaning that they can be
matched to any subtype.

```evy
func fn nums:[]num
    print nums
end

fn []
```

The `typeof` functions will return `"[]"` or `"{}"` for an empty
composite literal.

An array literal, such as `[1 2 3]`, has a type of `[]num`. However, it
is possible to assign an array literal of any type to a variable of
type `[]any`. It is important to note that this only applies to array
literals. A variable of type `[]num` cannot be assigned to a variable
of type `[]any`.

```evy
x := [1 2 3]
print "x" (typeof x)
y:[]any
y = [1 2 3]
print "y" (typeof y)
// y = x // parse error
// x = y // parse error
```

will output

```evy:output
x []num
y []any
```

## Type Assertion

A type assertion `x.(TYPE)` asserts that the value of the variable `x`
is of the given `TYPE`. `TYPE` can be any basic or composite type, such
as `num` or `[]string`. If the assertion does not hold, a
[run-time panic](#run-time-panics-and-recoverable-errors) occurs.

```evy
x:any
x = [1 2 3 4]
num_array := x.([]num)
print "typeof x:" (typeof x)
print "typeof num_array" (typeof num_array)
print

x = "abc"
str := x.(string)
print "typeof x:" (typeof x)
print "typeof str:" (typeof str)
```

Will generate the output

```ev:output
typeof x: any
typeof num_array: []num

typeof x: any
typeof str: string
```

Only values of type `any` can be type asserted. That means an array of
type any, `[]any`, _cannot_ be type asserted to be an array of type
`[]num` or any other concrete type. However, the elements of an array of
type `[]any` can be type assert, for example `arr[0].(num)`,

```evy
x:[]any
x = [1 2 3 true]
x = [1 2 3]
print "x:" x "typeof x:" (typeof x)
// print x.([]num) // parse error
// print x[0].(string) // run-time panic
```

outputs

```evy:output
x: [1 2 3] typeof x: []any
```

## Assignability

**Assignability** determines whether a value of one type can be assigned to a
variable of another type. This means that the variable _accepts_ the value.
Assignability rules apply to:

- assignments
- function parameters
- return values

In the assignment `target = val`, `val` can be a variable, a constant, or an
expression. A _constant_ is either a literal of type `num`, `string`, or
`bool`, or it is a composite literal that does not contain any variables. For
example, `[1 2 {}]` is a constant, but `[1 2 x]` is not. If `val` in
`target = val` is an expression that only contains constants, it is treated
like a constant; otherwise, it is treated like a variable.

### Assignability of variable values

If `target` is of type `t` and `val` is a _variable_ of type `t2`, `target`
accepts `val` if:

- `t` and `t2` are identical, or
- `t` is of type `any`

### Assignability of constant values

If `target` is of type `t` and `val` is a _constant_ of type `t2`, `target`
accepts `val` if:

- `t` and `t2` are identical, or
- `t` is of type `any`, or
- `t` is a composite with basic subtype `any` and `t2` can be **converted** to it.

A constant of type `t2` can be **converted** to type `t` if both types are
composite types of the same structure and the final subtype of `t` is `any`.
This means, for instance, that the literal array `[1 2 3]` of type
`[]num` can be assigned to a variable of type `[]any`.

The following code:

```evy
arr:[]{}any
arr = [{a:1} {b:[1 2 {}]} {}]
print (typeof arr)
print (typeof arr[0])
print (typeof arr[0].a)
```

will output:

```evy:output
[]{}any
{}any
num
```

### Assignability of empty composite literals

Empty composite literals `[]`, `{}`, or nested emtpy composite literals of
them, such as `[[]]`, follow the same rules as
[inferred declarations](#variables-and-declarations): `[]` gets converted to
type `[]any`, `{}` gets converted to type `{}any` and `[[]]` to type `[][]any`.

## Run-time Panics and Recoverable Errors

**Run-time panics** are unrecoverable errors that can occur during the
execution of an Evy program. They can be caused by a variety of things,
such as trying to index an array out of bounds, accessing a map value
for a key that does not exist, or a failed type assertion. When a
run-time panic occurs, the Evy program will stop and error details will
be printed. You can trigger a panic in your own code by calling the
built-in function `panic "msg"`.

**Recoverable errors** are errors that can be handled by the Evy program.
They are typically caused by user input or external factors that the
Evy program cannot control. Functions that can cause recoverable errors
set the global `err` variable to `true` and the string variable `errmsg`
to a description of the error. The Evy program can then check the value
of `err` and handle the error accordingly. You can trigger a recoverable
error in your own code by setting `err` and `errmsg`.

For more information on run-time panics and recoverable errors, see the
built-in documentation on the [panic function](builtins.md#panic) and
the [errors section](builtins.md#errors).

## Execution Model and Event Handlers

Evy first executes all top-level code in the order it appears in the
source code. If there is at least one event handler, Evy then enters an
event loop. In the event loop, Evy waits for external events, such as a
key press or a pointer down event. When an event occurs, Evy calls the
corresponding event handler function if it has been implemented. The
event handler function can optionally receive arguments, such as the
key character or the pointer coordinates. Once the event handler
function has finished, Evy returns to the event loop and waits for the
next event.

Event handlers are declared using the `on` keyword. Only predefined
events can be handled: `key`, `down`, `up`, `move`, `animate`, and
`input`. The parameters to the event handlers must match the expected
signature. The parameters can be fully omitted or fully specified. If
only some parameters are needed, use the anonymous `_` parameter.

For more information on individual event handlers, see the
[built-in documentation](builtins.md#event-handlers).

## Runtimes

Evy has two runtimes: the **terminal runtime** and the **browser runtime**.

The browser runtime can be tried at [play.evy.dev]. It fully
supports all built-in functions and event handlers as described in the
[built-in documentation](builtin.md).

To use the terminal runtime, first install Evy and then run

    evy run FILE.evy

in the terminal. This will execute the source code in the
given file. You can also use the evy command to format your source code
with

    evy fmt FILE.evy

For more details, run `evy run --help` or `evy fmt --help`. The terminal
runtime does not support event handlers or graphics functions.

[play.evy.dev](https://play.evy.dev)
