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

    program    = { statements | func | event_handler } .
    statements = { statement NL } .
    statement  =  assignment | declaration | 
                  loop | if | return | 
                  BREAK | EMPTY_STATEMENT .

    EMPTY_STATEMENT = .
    BREAK           = "break" .

    /* --- Assignment --- */
    assignment = ident "=" expr .
    ident      = LETTER { LETTER | UNICODE_DIGIT } .

    /* --- Declarations --- */
    decl          = typed_decl | inferred_decl .
    typed_decl    = ident ":" type .
    inferred_decl = ident ":=" toplevel_expr .

    type       = BASIC_TYPE | array_type | map_type | "any" .
    BASIC_TYPE = "num" | "string" | "bool" .
    array_type = type "[]" .
    map_type   = [type] "{}" .

    /* --- Expressions --- */
    toplevel_expr = func_call | type_assertion | expr.

    func_call = ident args .
    args      = { unary_expression } .

    type_assertion = type ( ident | ident_selector ) .
    ident_selector = ident selector { selector } .
    selector       = index | slice | dot_selector .
    index          = "[" expr "]" .
    slice          = "[" [expr] : [expr] "]" .
    dot_selector   = "." ident .

    expr       = unary_expr | expr OP expr .
    unary_expr = operand | UNARY_OP unary_expr .
    operand    = literal | ident | ident_selector | "(" toplevel_expr ")" .
    UNARY_OP   = "+" | "-" | "!" .
    OP         = "or" | "and" | REL_OP | ADD_OP | MUL_OP .
    REL_OP     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
    ADD_OP     = "+" | "-" .
    MUL_OP     = "*" | "/" | "%" .

    /* --- Literals --- */
    literal     =  num_lit | string_lit | BOOL_CONST | array_lit | map_lit .
    num_lit     =  DECIMAL_DIGIT { DECIMAL_DIGIT } |
                   DECIMAL_DIGIT { DECIMAL_DIGIT } "." { DECIMAL_DIGIT } .
    string_lit  = """ { UNICODE_CHAR } """ .
    BOOL_CONST  = "true" | "false" .
    array_lit   = type "[" array_elems "]" .
    array_elems = { unary_expr [NL] }
    map_lit     = [type] "{" map_elems "}" .
    map_elems   = { ident ":" unary_expr [NL] } .
    
    /* --- Control flow --- */
    loop       = for | while .
    for        = "for" range NL
                     statements
                 END .
    range      = ident ( ":=" | "=" ) "range" range_args .
    range_args = unary_expr [ unary_expr [ unary_expr ] ] .
    while      = "while" toplevel_expr NL
                     statements
                 END .

    if = "if" toplevel_expr NL
                statements
          { "else" "if" toplevel_expr NL
                statements }
          [ "else" NL
                statements ]
          END .

    /* --- Functions ---- */
    func            = "func" func_signature NL
                          statements
                      END .
    func_signature  = ident [ ":" type ] params .
    params          = { param } | variadic_param .
    param           = ident ":" type .
    variadic_param  = param "..." .
    return          = "return" [toplevel_expr] .

    event_handler   = "on" ident NL
                          statements
                      END .

    /* --- Terminals --- */
    LETTER         = UNICODE_LETTER | "_" .
    UNICODE_LETTER = /* a Unicode code point categorized as "Letter" (category L) */ .
    UNICODE_DIGIT  = /* a Unicode code point categorized as "Number, decimal digit" */ .
    UNICODE_CHAR   = /* an arbitrary Unicode code point except newline */ .
    DECIMAL_DIGIT  = "0" … "9" .
    END            = "end" .
    NL             = "\n" .

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
[run-time panic] occurs.

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

[run-time panic]:#run-time-panics-and-recoverable-errors

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

    reflect "abc"           // {type: "string"}
    reflect true            // {type: "bool"}
    reflect num[1 2]        // {type: "array", 
                            //  sub:  {type: "num"}
                            // }
    reflect num[[1 2] [3 4] // {
                            //   typ e: "array", 
                            //   sub:  {
                            //     type: "array"
                            //     sub: {
                            //       type: "num"
                            //     }
                            //   }
                            // }

## Type assertion

A type assertion `type ident` asserts that the value of the variable
`ident` is of the given `type`. This is particularly useful for a
variable of type `any`, `any[]`, `any{}`, `any[][]` etc. The value
returned by the assertion is of given `type` and can be used in a
declaration, assignment or function call. If the assertion does not hold
a run-time panic occurs.

    x:any
    x = num[ 1 2 3 4 ]  
    num_array := num[] x // concrete type num[] 

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

A panic can be triggered with `panic "messsage"`.

Functions that can cause recoverable errors set the global string
variable `error` and the error classification number `errno`.

