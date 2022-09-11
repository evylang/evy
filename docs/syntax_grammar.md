# Evy syntax specification

`evy` is a [statically typed], [garbage collected],
[procedural] programming language. Its main design goal is to help
learn programming. `evy` strives for simplicity and directness in its
tooling and syntax. Several features typical of modern programming
languages are purposefully left out.

`evy`'s syntax is specified using a [WSN] grammar, a variant of
[EBNF] grammars, borrowing concepts from the [Go Programming Language
Specification].

For an intuitive understanding of `evy` see its [syntax by example].

[statically typed]: https://developer.mozilla.org/en-US/docs/Glossary/Static_typing
[garbage collected]: https://en.wikipedia.org/wiki/Garbage_collection_(computer_science)
[procedural]: https://en.wikipedia.org/wiki/Procedural_programming
[WSN]: https://en.wikipedia.org/wiki/Wirth_syntax_notation
[EBNF]: https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form
[Go Programming Language Specification]: https://go.dev/ref/spec
[syntax by example]: syntax_by_example.md

## WSN syntax grammar

The top level element of a WSN grammar is called _production_. For
example, `OPERATOR = "+" | "-" | "*" | "/" .` is a production
specifying an operator. A production consists of an _expression_
assigned to an _identifier_ or production name. Each production is
terminated by a period `.`. An expression consists of terms and the
following operators in increasing precedence:

    |   alternation: `a | b` stands for `a` or `b`.
    ()  grouping: `(a|b)c` stands for `ac | bc`.
    []  optionality:  `[a]b` stands for `ab | b`.
    {}  repetition (0 to n times):  `{a}` stands for `ε | a | aa | aaa | ....`.

`a … b` are single characters from `a` through `b` as alternatives.

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
    literal    = """ CHARACTER { CHARACTER } """ .
    LETTER     = "a" … "z" | "A" … "Z" | "_" .
    CHARACER   = /* an arbitrary Unicode code point */ .

By convention, upper case production names identify _terminal tokens_.
Terminal tokens are the leaves in the grammar that cannot be expanded
further. Lower case production names identify _non-terminals_, which
are production names that may be expanded further. Lexical tokens are
enclosed in double quotes `""`. Comments are fenced by `/* … */`.

## Evy syntax grammar

The `evy` source code is UTF-8 encoded. The NUL character `U+0000` is
not allowed.

    program    = { statement | func | event_handler } .
    statement  = empty_stmt | 
                 assign_stmt | typed_decl_stmt | inferred_decl_stmt |
                 func_call_stmt | 
                 return_stmt | break_stmt |
                 for_stmt | while_stmt | if_stmt  .


    /* --- Statement ---- */
    empty_stmt = NL .

    assign_stmt        = assignable "=" expr NL .
    typed_decl_stmt    = typed_decl NL .
    inferred_decl_stmt = ident ":=" toplevel_expr NL .

    func_call_stmt = func_call NL.

    return_stmt = "return" [ toplevel_expr ] NL.
    break_stmt  = "break" NL .

    /* --- Assignment --- */
    assignable     = ident { selector } .
    ident          = LETTER { LETTER | UNICODE_DIGIT } .
    selector       = index | dot_selector .
    index          = "[" expr "]" .
    dot_selector   = "." ident .

    /* --- Type --- */
    typed_decl         = ident ":" type .

    type       = BASIC_TYPE | array_type | map_type | "any" .
    BASIC_TYPE = "num" | "string" | "bool" .
    array_type = type "[]" .
    map_type   = [type] "{}" .

    /* --- Expressions --- */
    toplevel_expr = func_call | expr .

    func_call = ident args .
    args      = { term } .

    type_assertion = assignable "." "(" type ")" .

    expr       = term | expr OP expr .
    term       = operand | UNARY_OP term .
    operand    = literal | assignable | slice | type_assertion | "(" toplevel_expr ")" .
    UNARY_OP   = "+" | "-" | "!" .
    OP         = "or" | "and" | REL_OP | ADD_OP | MUL_OP .
    REL_OP     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
    ADD_OP     = "+" | "-" .
    MUL_OP     = "*" | "/" | "%" .

    /* --- Slice and Literals --- */
    slice       = assignable "[" [expr] : [expr] "]" .
    literal     = num_lit | string_lit | BOOL_CONST | array_lit | map_lit .
    num_lit     = DECIMAL_DIGIT { DECIMAL_DIGIT } |
                  DECIMAL_DIGIT { DECIMAL_DIGIT } "." { DECIMAL_DIGIT } .
    string_lit  = """ { UNICODE_CHAR } """ .
    BOOL_CONST  = "true" | "false" .
    array_lit   = type "[" array_elems "]" .
    array_elems = { term [NL] }
    map_lit     = [type] "{" map_elems "}" .
    map_elems   = { ident ":" term [NL] } .
    
    /* --- Control flow --- */
    for_stmt   = "for" range NL
                    { statement }
                 "end" NL .
    range      = ident ( ":=" | "=" ) "range" range_args .
    range_args = term [ term [ term ] ] .
    while_stmt = "while" toplevel_expr NL
                     { statement }
                 "end" NL .

    if_stmt = "if" toplevel_expr NL
                    { statement }
              { "else" "if" toplevel_expr NL
                    { statement } }
              [ "else" NL
                    { statement } ]
              "end" NL .

    /* --- Functions ---- */
    func            = "func" ident func_signature NL
                          { statement }
                      "end" NL .
    func_signature  = [ ":" type ] params .
    params          = { typed_decl } | variadic_param .
    variadic_param  = typed_decl "..." .

    event_handler   = "on" ident NL
                          { statement }
                      "end" NL .

    /* --- Terminals --- */
    LETTER         = UNICODE_LETTER | "_" .
    UNICODE_LETTER = /* a Unicode code point categorized as "Letter" (category L) */ .
    UNICODE_DIGIT  = /* a Unicode code point categorized as "Number, decimal digit" */ .
    UNICODE_CHAR   = /* an arbitrary Unicode code point except newline */ .
    DECIMAL_DIGIT  = "0" … "9" .
    NL             = "\n" .

## Comments

There is only one type of comment, the line comment which starts with
`//` and stops at the end of the line. Line comments cannot start
inside string literals.

## Variables and Zero values

A variable is a storage location for holding a value. The set of
permissible values is determined by the variable's [type](#types).

A variable must be _declared_ before it can be used either in
an _inferred declaration_ or in a _typed declaration_.

With the inferred declaration the type is not given but inferred from
the value. For example `a := 2` declares variable `a` of type `num`
holding the value `2`. `num` is the inferred type.

Variables declared via typed declaration are not given a value, for
example `x:num`. They are initialised to the zero value of their type:

    Type        Zero
    num         0
    string      ""
    bool        false
    any[]       [] // empty array
    any{}       {} // empty map

## Assignment and Expressions

When a variable is assigned to a new variable or passed as an argument
to a function, a copy of its value is made. Modifying the copy does not
change the original, for example:

    arr1 := [ 1 2 ]
    arr2 := arr1
    arr2[0] = 2
    print arr1 arr2 // [ 1 2 ] [ 2 2 ]

The following assignment statement is ambiguous `a := b - 1`. Does the
expression `b - 1` represent a function call of function `b` with
argument `-1`? Or is `b` a variable so that `b - 1` becomes the
arithmetic expression "`b` minus `1`"?

`evy` resolves this ambiguity by tracking identifiers in a symbol table
and annotating them as _variable_ or _function_ names. If `b` has not
been declared as a variable or function by the point where the
expression `b - 1` is seen it is assumed that `b` is the name of a
function. This allows for [mutual recursion] of functions.

[mutual recursion]: https://en.wikipedia.org/wiki/Mutual_recursion

## Types

There are three basic types: `string`, `bool` and `num` as well as two
composite types: [arrays](#arrays) `[]` and [maps](#maps) `{}`.
The _dynamic_ type `any` can hold any of the previously listed
types.

Composite types can nest further composite types, for example `num[][]{}`.

A `bool` value is either `true` or `false`.

A number value can be expressed as integer `1234` or decimal `56.78`.
Internally a number is represented as a [double-precision floating-point number]
according to the IEEE-754 64-bit floating point standard.

[double-precision floating-point number]: https://en.wikipedia.org/wiki/Double-precision_floating-point_format

## Strings

A `string` is a sequence of [Unicode code points]. A string literal is
enclosed by double quotes `"`, for example str := "Hallöchen Welt 🌏".

`len str` returns the number of Unicode code points, _characters_, in
the string. `for ch := range str` iterates over all characters of the
string. Individual characters of a string can be addressed and updated
by index. Strings can be concatenated with the `+` operator.

    str := "hello!"
    str[0] = "H"             // Hello
    str1 := str + ", " + str // Hello, Hello

[Unicode code points]: https://en.wikipedia.org/wiki/Unicode

## Arrays

Arrays are declared with brackets `[]`. Their elements have a type, for
example `arr := num[ 1 2 3 ]` is an array of `num`. Arrays can be
nested `arr:num[][]{}`.

An array of type `any` can be composed of different types:

    arr := any[ "abc" 123 ]

`len arr` returns the length of the array. `for el := range arr`
iterates over all elements of the array. `arr = append arr 1` and `arr
= prepend arr 0` add a new element to beginning or end of the array.
Arrays can be concatenated with the `+` operator `arr2 := arr + arr`.

## Maps

Map keys must be strings that match the grammars `ident` production. Map
values can be accessed with the dot selector, for example `map.key`.
They can also be accessed with a key expression and the index selector:

    m := { letters: "abc" }
    print m.letters    // abc
    print m["letters"] // abc

    s := "letters"
    print m[s]         // abc

The `has` function tests for the existence of a key in a map:

    has m "letters"    // true
    has m "digits"     // false

`for key := range map` iterates over all map keys.

`len m` returns the number of values in the map.

When leaving out the value type of a map, `any` is inferred. Therefore
`m:any{}`, `m:{}`, `m := any{}` and `m := {}` are all equivalent.

## Operators

Binary operations can only be executed with operands of the same type.
There is no automated type conversion of operands.

    operands   operators      result
    num        + - * / % ^    num
    string     +              string
    array      +              array
    bool       and or         bool
    num        <  <=  >  >=   bool
    string     <  <=  >  >=   bool

`==` and `!=` compare to operands of the same type for equality and
have a `bool` result.

`+` `-` `*` `/` `%` stand for addition, subtraction, multiplication,
division and the [modulo operator]. `+` may also be used as
concatenation operator with `string` and `array`.

Boolean operators `and`, `or` stand for [logical conjunction (AND)] and
[logical disjunction (OR)]. Comparison operators `<`  `<=`  `>`  `>=`
stand for less, less or equal, greater, greater or equal. Their
operands may be `num` or `string` values. For `string`
[lexicographical comparison] is used.

The unary operator `-` stands for the negative sign and can only be used
with `num`. The unary operator `!` stands for [logical negation] and
can only be used with `bool`.

[modulo operator]: https://en.wikipedia.org/wiki/Modulo_operation
[logical conjunction (AND)]: https://en.wikipedia.org/wiki/Truth_table#Logical_conjunction_(AND)
[logical disjunction (OR)]: https://en.wikipedia.org/wiki/Truth_table#Logical_disjunction_(OR)
[logical negation]: https://en.wikipedia.org/wiki/Truth_table#Logical_negation
[lexicographical comparison]: https://en.wikipedia.org/wiki/Lexicographic_order

## Precedence

Unary operators, `( … )` and selectors `a[i]`, `a.b` have the highest
precedence, followed by binary operators. The binary operators have the
following order of precedence:

    precedence    Operator
        5             *  /  %
        4             +  -
        3             ==  !=  <  <=  >  >=
        2             and
        1             or

## Comments

There is only one type of comment, the line comment which starts with
`//` and stops at the end of the line. Line comments cannot start
inside string literals.

## Variables and Zero values

Variables hold values of type `num`, `string`, `bool`, array or map.
They must be _declared_ before they can be used. A variable declaration
can either be an _inferred declaration_ or a _typed delaration_.

With the inferred declaration the type is not given but inferred from
the value. For example `a := 2` declares variable `a` of type `num`
holding the value `2`. `num` is the inferred type.

Variables declared via typed declaration are not given a value, for
example `x:num`. They are initialised to the zero value of their type:

    Type        Zero
    num         0
    string      ""
    bool        false
    any[]       [] // empty array
    any{}       {} // empty map

## Arrays

Arrays are declared with brackets `[]`. Their elements have a type, for
example `arr := int[ 1 2 3]` is an array of `num`. Arrays can be nested
`arr:num[][]{}`.

An array of type `any` can be composed of different types:

    arr := any[ "abc" 123]

`len arr` returns the length of the array. `for el := range arr`
iterates over all elements of the array. `arr = append arr 1` and `arr
= prepend arr 0` add a new element to beginning or end of the array.
Arrays can be concatenated with the `+` operator `arr2 := arr + arr`.

## Strings

`len str` returns the number of Unicode code points, _characters_, in
the string. `for ch := range str` iterates over all characters of the
string. Individual characters of a string can be addressed and updated
by index. Strings can be concatenated with the `+` operator.

    str := "hello!"
    str[0] = "H"             // Hello
    str1 := str + ", " + str // Hello, Hello

## Index and Slice

The first index of an array or string is `0`. A negative index `-i` is a
short hand for `(len a) - i`, for example `a[-1]` references the last
element. When trying to index an array or string out of bounds a
[run-time panic](#run-time-panics-and-recoverable-errors) occurs.

Portions of an array or string can be copied with the slice selector,
for example `a[1:3]`. `a[start : end]` copies a substring or subarray,
a _slice_, starting with the value at `a[start]`. The length of the
slice is `end - start`. The end index `a[end]` is not included in the
slice. If `start` is left out it defaults to 0. If `end` is left out it
defaults to `len a`, for example:

    s := "abcd"
    print s[1:3] // bc
    print s[:2]  // ab
    print s[2:]  // cd
    print s[:]   // abcd
    print s[:-1] // abc

Slices may not be sliced further, `a[:2][1:]` is illegal.

## Maps

Map keys can only be strings shaped like identifiers so that map values
can be accessed via dot selector, `map.key`. They may also
use a key expression with the index selector:

    m := { letters: "abc" }
    print m.letters    // abc
    print m["letters"] // abc

    s := "letters"
    print m[s]         // abc

The `has` function tests for the existence of a key in a map:

    has m "letters"    // true
    has m "digits"     // false

`for key := range map` iterates over all map keys.

`len m` returns the number of values in the map.

When leaving out the value type of a map, `any` is inferred. Therefore
`m:any{}`, `m:{}`, `m := any{}` and `m := {}` are all equivalent.

## Assignment

When a variable of any type is assigned to a new variable or passed as
an argument to a function, a copy is made. Modifying the copy does not
change the original, for example:

    arr1 := [ 1 2 ]
    arr2 := arr1
    arr2[0] = 2
    print arr1 arr2 // [ 1 2 ] [ 2 2 ]

## Break and Return

`break` and `return` are terminating statements. They interrupt the
regular flow of control. `break` is used to exit from the inner-most
loop body. `return` is used to exit from a function and may be followed
by an expression whose value is returned by the function call.

## Functions

A function declaration binds an identifier, the function name, to a
function. As part of the function declaration, the function signature
defines the number, order and types of input parameters as well as the
result or return type of the function. If the return type is left out
the function does not return a value.

    func is_valid:bool text:string cap:num
        return (len text) <= cap
    end

The example above has a function name of `is_valid`, input parameters
`text` of type `string` and `cap` of type `num`. The return type is
`bool`.

## Variadic functions

A function with a single parameter may have a type suffixed with `...`.
A function with such a parameter is called variadic and may be invoked
with zero or more arguments for that parameter.

If `f` is variadic with a parameter `p` of type `T...`, then within `f`
the type of `p` is equivalent to type `T[]`. The length of the array is
the number of arguments bound to `p` and may differ for each call.

For example, `my_print` is a variadic function

    func my_print args:any...
        for arg := range args
            write arg
            write " "
        end
    end

It can be called as `my_print "hello" "world" true 42`

Unlike other languages, arrays cannot be turned into variadic arguments
at the call site. The call arguments must be listed individually.

## Reflection

Reflection retrieves the type of the value held by a variable. This is
particularly useful for a variable of type `any`, `any[]`, `any
{}`, `any[][]` etc. The function call `reflect val` returns a map with
keys `type` and optionally `sub`. The `type` value is one of `"num"`,
`"string"`, `"bool"`, `"any"`, `"array"`, `"map"`. For `"array"` and
`"map"` the key `"sub"` contains another map with keys `type` and
optionally `sub` representing the type of the array elements or map
values.

    reflect "abc"              // {type: "string"}
    reflect true               // {type: "bool"}
    reflect num[ 1 2 ]         // {type: "array",
                               //  sub:  {type: "num"}
                               // }
    reflect num[ [1 2] [3 4] ] // {
                               //   type: "array",
                               //   sub:  {
                               //     type: "array"
                               //     sub: {
                               //       type: "num"
                               //     }
                               //   }
                               // }

## Type assertion

A type assertion `ident.(type)` asserts that the value of the variable
`ident` is of the given `type`. This is particularly useful for a
variable of type `any`, `any[]`, `any{}`, `any[][]` etc. The value
returned by the assertion is of given `type` and can be used in a
declaration, assignment or function call. If the assertion does not hold
a [run-time panic](#run-time-panics-and-recoverable-errors) panic occurs.

    x:any
    x = num[ 1 2 3 4 ]  
    num_array := x.(num[])
    x = "abc"
    str := x.(string)

Only variables or variable selectors of type `any` can be type asserted.
That means an array of type `any`, _cannot_ be type assert
to be an array of type `num` or other concrete type:

    x:any[]
    x = [1 2]
    // x.([]num) // compile time error
    x[1] = [3 4 5]

However, the elements of `x` can be type assert, e.g. `x[0].(num)`, `x
[1].([]num)`.

## Event Handler

An event handler starts with `on`, followed by an event name and a block
of statements. The statements get executed when the given event is
triggered. Events can be triggered by user interaction, for example
clicking the mouse or tapping the keyboard or by the system, for
example `animate` when a new frame is painted.

There is a limited, predefined set of events. It is not possible to
create custom events.

    on mouse_down
        print mouse_x mouse_y
    end

    on animate
        draw
    end

The `animate` event is triggered every 2 Milliseconds, 50 times per
second.

## Run-time Panics and Recoverable Errors

Execution errors such as trying to index an array out of bounds or
access a map value for a key that does not exist or a failed type
assertion trigger a run-time panic. The execution of the `evy` program
stops and error details are printed.

A panic can be triggered with `panic "message"`.

Functions that can cause recoverable errors set the global string
variable `error` and the error classification number `errno`.
