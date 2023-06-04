# Evy syntax specification

`evy` is a [statically typed], [garbage collected],
[procedural] programming language. Its main design goal is to help
learn programming. `evy` aims for simplicity and directness in its
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
    CHARACTER  = /* an arbitrary Unicode code point */ .

By convention, upper case production names identify _terminal tokens_.
Terminal tokens are the leaves in the grammar that cannot be expanded
further. Lower case production names identify _non-terminals_, which
are production names that may be expanded further. Lexical tokens are
enclosed in double quotes `""`. Comments are fenced by `/* … */`.

There are two special fencing tokens in evy's grammar related to
horizontal whitespace, `<-` … `->` and `<+` … `+>`. `<-` … `->` means
no horizontal whitespace is allowed between the terminals of the
enclosed expression, e.g. `3+5` inside `<-` … `->` is allowed, but
`3 + 5` is not. The fencing tokens `<+` … `+>` are the default and mean
horizontal whitespace is allowed (again) between terminals. 

See section [whitespace](#whitespace) for further details.

## Evy syntax grammar

The `evy` source code is UTF-8 encoded. The NUL character `U+0000` is
not allowed.

    program    = { statement | func | event_handler | NL } .
    statements = statement { statement } .
    statement  = typed_decl_stmt | inferred_decl_stmt |
                 assign_stmt | 
                 func_call_stmt | 
                 return_stmt | break_stmt |
                 if_stmt | for_stmt | while_stmt .

    /* --- Functions and Event handlers ---- */
    func            = "func" ident func_signature NL
                          statements
                      "end" NL .
    func_signature  = [ ":" type ] params .
    params          = { typed_decl } | variadic_param .
    variadic_param  = typed_decl "..." .

    event_handler   = "on" ident NL
                          statements
                      "end" NL .

    /* --- Control flow --- */
    if_stmt = "if" toplevel_expr NL
                    statements
              { "else" "if" toplevel_expr NL
                    statements }
              [ "else" NL
                    statements ]
              "end" NL .

    for_stmt   = "for" range NL
                    statements
                 "end" NL .
    range      = [ ident ":=" ] "range" range_args .
    range_args = <- expr -> [ <- expr -> [ <- expr -> ] ] .
    while_stmt = "while" toplevel_expr NL
                     statements
                 "end" NL .

    return_stmt = "return" [ toplevel_expr ] NL .
    break_stmt  = "break" NL .

    /* --- Statement ---- */
    assign_stmt        = assignable "=" toplevel_expr NL .
    typed_decl_stmt    = typed_decl NL .
    inferred_decl_stmt = ident ":=" toplevel_expr NL .
    func_call_stmt     = func_call NL .

    /* --- Assignment --- */
    assignable     = <- ident | index_expr | dot_expr -> . /* no WS around `[…]` and `.` */
    ident          = LETTER { LETTER | UNICODE_DIGIT } .
    index_expr     = assignable "[" expr "]" .
    dot_expr       = assignable "." ident .

    /* --- Type --- */
    typed_decl     = <- ident ":" type -> . /* no WS allowed. */
    type           = BASIC_TYPE | DYNAMIC_TYPE | COMPOSITE_TYPE .
    BASIC_TYPE     = "num" | "string" | "bool" .
    DYNAMIC_TYPE   = "any" .
    COMPOSITE_TYPE = array_type | map_type .
    array_type     = "[]" type .
    map_type       = "{}" type .

    /* --- Expressions --- */
    toplevel_expr = func_call | expr .

    func_call = ident args .
    args      = { tight_expr } .  /* no WS within single arg, WS is arg separator */

    tight_expr = <- expr -> .     /* no WS allowed unless within `(…)`, `[…]`, or `{…}` */
    expr       = operand | unary_expr | binary_expr .

    operand    = literal | assignable | slice | type_assertion | group_expr .
    group_expr = "(" <+ toplevel_expr +> ")" . /* WS can be used freely within `(…)` */
    type_assertion = <- assignable "." "(" type ")" -> .
    
    unary_expr = <- UNARY_OP -> expr .  /* WS not allowed after UNARY_OP */
    UNARY_OP   = "-" | "!" .

    binary_expr   = expr BINARY_OP expr .
    BINARY_OP     = LOGICAL_OP | COMPARISON_OP | ADD_OP | MUL_OP .
    LOGICAL_OP    = "or" | "and" .
    COMPARISON_OP = "==" | "!=" | "<" | "<=" | ">" | ">=" .
    ADD_OP        = "+" | "-" .
    MUL_OP        = "*" | "/" | "%" .

    /* --- Slice and Literals --- */
    slice       = assignable "[" [expr] ":" [expr] "]" .
    literal     = num_lit | string_lit | BOOL_CONST | array_lit | map_lit .
    num_lit     = DECIMAL_DIGIT { DECIMAL_DIGIT } |
                  DECIMAL_DIGIT { DECIMAL_DIGIT } "." { DECIMAL_DIGIT } .
    string_lit  = """ { UNICODE_CHAR } """ .
    BOOL_CONST  = "true" | "false" .
    array_lit   = "[" <+ array_elems +> "]" . /* WS can be used freely within `[…], but not inside the elements` */
    array_elems = { tight_expr [NL] } .
    map_lit     = "{" <+ map_elems +> "}" .   /* WS can be used freely within `{…}, but not inside the values` */
    map_elems   = { ident ":" tight_expr [NL] } .

    /* --- Terminals --- */
    LETTER         = UNICODE_LETTER | "_" .
    UNICODE_LETTER = /* a Unicode code point categorized as "Letter" (category L) */ .
    UNICODE_DIGIT  = /* a Unicode code point categorized as "Number, decimal digit" */ .
    UNICODE_CHAR   = /* an arbitrary Unicode code point except newline */ .
    DECIMAL_DIGIT  = "0" … "9" .
    NL             = "\n" {"\n"} .
    WS             = " " | "\t" {" " | "\t"} .


## Comments

There is only one type of comment, the line comment which starts with
`//` and stops at the end of the line. Line comments cannot start
inside string literals.

## Types

There are three basic types: `string`, `bool` and `num` as well as two
composite types: [arrays](#arrays) `[]` and [maps](#maps) `{}`.
The _dynamic_ type `any` can hold any of the previously listed
types.

Composite types can nest further composite types, for example 
`[]{}string` is an array of maps with string values.

A `bool` value is either `true` or `false`.

A number value can be expressed as integer `1234` or decimal `56.78`.
Internally a number is represented as a [double-precision floating-point number]
according to the IEEE-754 64-bit floating point standard.

[double-precision floating-point number]: https://en.wikipedia.org/wiki/Double-precision_floating-point_format

## Variables and Declarations

Variables hold values of the type specified in the variable declaration.

Variables must be _declared_ before they can be used. A variable
declaration can either be an _inferred declaration_ or a _typed
declaration_. With the inferred declaration the type is not given but
inferred from the value. With typed declaration the type is explicitly
specified and the variable is initialised to its type's zero value.

    a1 := 1 // inferred declaration of variable 'a1' 
            // with type 'num' and value 1.
    a2:num  // typed declaration of variable 'a2'
            // with type 'num' and zero value 0.

`arr := []` infers an array of type any, `[]any`. `map := {}` infers a
map of type any, `{}any`. The strictest possible type is inferred for
composite types:

    arr := [ 1 2 3 ]     // type: []num
    arr := [ 1 "a" ]     // type: []any
    arr := [ [1] ["a"] ] // type: [][]any
    arr := []            // type: []any
    arr := [ 1 ] + []    // type: []num
    m := {}              // type: {}any
    m := {age: 10}       // type: {}num

## Zero Values

Variables declared via typed declaration are initialised to the zero
value of their type:

    Type        Zero
    num         0
    string      ""
    bool        false
    []ANY       [] // empty array of given type, if no type given: array of any
    {}ANY       {} // empty map of given type, if no type given: map of any

## Assignments

Assignments are defined by an equal sign `=`. The left hand side of the
`=` must contain an _assignable_, a variable, an indexed array 
(`arr[1] = "abc"`) or a map field (`person.age = 42`), see [Arrays](#Arrays) 
and [Maps](#Maps) for further details. Before the assignment
the variable must be declared via inferred (`:=`)or typed declaration
(`:TYPE`). Only values of the correct type can be assigned to a
variable.

    a := 1
    print a // prints 1
    a = 2
    print a // prints 2
    a = "abc" // compile time error, wrong type

## Copy and reference

When a variable of a basic type `num`, `string`, or `bool` is the value
of an assignment, a copy of its value is made. A copy is also made when
a variable of basic type is used as the value in an inferred
declaration or passed as an argument to a function.

    a := 1
    b := a
    print a b // 1 1
    a = 2
    print a b // 2 1 - `b` keeps its initial value

By contrast, composite types - maps and arrays - are passed by
reference and no copy is made. Modifying the contents of an array
referenced by one variable also modifies the contents of the array
referenced by another variable. This is also true for argument passing
and inferred declarations:

    a := [1]
    b := a
    print a b // [1] [1]
    a[0] = 2
    print a b // [2] [2] - b also has updated contents

For the dynamic type `any`, a copy is made if the value is of basic type
and the variable is passed by reference if the value is a composite type.

See [Functions](#Functions), [Maps](#Maps) and [Arrays](#Arrays) for
further details.

## Scope

Functions can only be defined at the top level of the program, known
as _global scope_. A function does not have to be defined before it can
be used. This allows for [mutual recursion] of functions - `func a`
calling `b` and `func b` calling func `a`.

Variables by contrast must be declared and given an unchangeable type
before they can be used. Variables can be declared at the top level of
the program, at _global scope_, or within a block-statement, at _block
scope_. 

A _block-statement_ is a block of statements that ends with the keyword
`end`. A function body following the line starting with `func` is a
block-statement. The statements between `if` and `else` are a block
(statement). The statements between `while`/`for`/`else` and `end` are
a block. Blocks can be nested within other blocks.

A variable declared inside a block only exists until the end of the
block and may not be used outside the block. 

Variable names in an inner block can shadow or override the same
variable name from an outer block, which makes the variable of the
outer block inaccessible to the inner block. However, when the inner
block is finished the variable from the outer block is restored and
unchanged:

    x := "outer"
    print "1" x
    for i := range 1
        x := true
        print "2" x
    end
    print "3" x

This program will print
    
    1 outer
    2 true
    3 outer

[mutual recursion]: https://en.wikipedia.org/wiki/Mutual_recursion

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
example `arr := [1 2 3]` and `arr:[]num` are arrays of type `num`.
Arrays can be nested `arr:[]{}string`.

An array composed of different types becomes an array of `any`:

    arr := ["abc" 123] // type: []any

`len arr` returns the length of the array. `for el := range arr`
iterates over all elements of the array in order. `append arr 1` and
`prepend arr 0` add a new element to end or beginning of the array.
Arrays can be concatenated with the `+` operator `arr2 := arr + arr`.

The elements of an array can be accessed via index starting at 0. In the
example above the first element in the array `arr[0]` is `"abc"`.

The empty arrays becomes `[]any` in inferred declarations, otherwise the
empty array literal assumes the type required. `arr:[]any` and `arr :=
[]` are equivalent.

In order to distinguish between array literals and array indices, there
cannot be any whitespace between array variable and index.

    arr := ["a" "b"]
    print arr[1]  // index: b
    print arr [1] // literal: [a b] [1]
    arr[0] = "A"
    arr [1] = "B" // invalid

See section [whitespace](#whitespace) for further details.

## Maps

Maps are key-value stores, where the values can be looked up by their
key

    m := { key1:"value1" key2:"value2" }

Map keys must be strings that match the grammars `ident` production. Map
values can be accessed with the dot expression, for example `map.key`.
Map values can also be accessed with an index which allows for
evaluation and variable usage:

    m := { letters: "abc" }
    print m.letters    // abc
    print m["letters"] // abc

    s := "letters"
    print m[s]         // abc

The `has` function tests for the existence of a key in a map:

    has m "letters"     // true
    has m "digits"      // false

The `del` function deletes a key from a map if it exists:

    del m "letters"    // m == {}

`for key := range map` iterates over all map keys. It is safe to delete
values from the map with the builtin function `del` while iterating.
The keys are iterated in the order in which they are inserted. Any
values inserted into the map during the iteration will not be included
in the iteration.

`len m` returns the number of values in the map.

The empty map literal becomes `{}any` in inferred declarations,
otherwise the empty map literal assumes the type required. 
`m:{}any` and `m := {}` are equivalent.

The dot expression `.` and the index expression `[ "key" ]` must not be
surrounded by any whitespace. See section [whitespace](#whitespace) for
further details.

## Index and Slice

The first index of an array or string is `0`. A negative index `-i` is a
short hand for `(len a) - i`. Therefore `arr[-1]` references the last
element of `arr`. When trying to index an array or string out of bounds
a [run-time panic](#run-time-panics-and-recoverable-errors) occurs.

Portions of an array or string can be copied with the slice selector,
for example `a[1:3]`. `a[start : end]` copies a substring or subarray,
a _slice_, starting with the value at `a[start]`. The length of the
slice is `end - start`. The end index `a[end]` is not included in the
slice. If `start` is left out it defaults to 0. If `end` is left out it
defaults to `len a`.

    s := "abcd"
    print s[1:3] // bc
    print s[:2]  // ab
    print s[2:]  // cd
    print s[:]   // abcd
    print s[:-1] // abc

Slices may not be sliced further, `a[:2][1:]` is illegal.

Slice expressions must nut be preceded by whitespace, like array or
string indexing. See section [whitespace](#whitespace) for further
details.

## Operators

Binary operations can only be executed with operands of the same type.
There is no automated type conversion of operands.

    operands   operators      result
    num        + - * / %      num
    string     +              string
    array      +              array
    bool       and or         bool
    num        <  <=  >  >=   bool
    string     <  <=  >  >=   bool

`==` and `!=` compare two operands of the same type for equality and
have a `bool` result.

`+` `-` `*` `/` `%` stand for addition, subtraction, multiplication,
division and the [modulo operator]. `+` may also be used as
concatenation operator for `string` and `array` types. 

Boolean operators `and`, `or` stand for [logical conjunction (AND)] and
[logical disjunction (OR)]. They perform [short-circuit evaluation]
where the right-hand side of the operator is not evaluated if the result
of the operation can be determined from the left-hand side alone.
Comparison operators `<`  `<=`  `>`  `>=` stand for less, less or equal,
greater, greater or equal. Their operands may be `num` or `string`
values. For `string` types [lexicographical comparison] is used.

The unary operator `-` stands for the negative sign and can only be used
with `num`. The unary operator `!` stands for [logical negation] and
can only be used with `bool`. Unary operators `-` and `!` must not be
followed by horizontal whitespace.

    a := 1
    b := 2
    print a-b     // -1
    print (a - b) // -1
    print a -b    // 1 -2
    print a - b   // compile time error

See section [whitespace](#whitespace) for further details.

[modulo operator]: https://en.wikipedia.org/wiki/Modulo_operation
[logical conjunction (AND)]: https://en.wikipedia.org/wiki/Truth_table#Logical_conjunction_(AND)
[logical disjunction (OR)]: https://en.wikipedia.org/wiki/Truth_table#Logical_disjunction_(OR)
[short-circuit evaluation]: https://en.wikipedia.org/wiki/Short-circuit_evaluation
[logical negation]: https://en.wikipedia.org/wiki/Truth_table#Logical_negation
[lexicographical comparison]: https://en.wikipedia.org/wiki/Lexicographic_order

## Precedence

The index `a[i]`, dot `a.b` and group `(` … `)`expressions have the
highest precedence, followed by the unary operators `-` and `!`.
Finally, binary operators have the following order of precedence:

    precedence    Operator
        6             *  /  %
        5             +  -
        4             <  <=  >  >=
        3             ==  !=  
        2             and
        1             or

## Whitespace

Vertical whitespace is newlines, `NL` in the grammar. It delimits
statements. Array and map literals are the exception as they can be
very large and allow for whitespace `NL` within:

    person := {
        name: "Jane Goddall"
        born: 1934
    }               // valid map declaration
    x := 1 +
         2          // compile time error

Horizontal whitespace is tabs or spaces, `WS` in the grammar. `WS` is
used as a separator between elements in lists and cannot be
used _within_ an element. These lists include argument list to a
function call and the element lists of array literals or map literals.
To further avoid confusion, whitespace within index expression, dot
expressions or unary expressions is not allowed.

More formally, `WS` between tokens or terminals as defined in the
grammar is ignored except for the following cases:

1. `WS` is not allowed in assignables, around `DOT`, or before array or
map index. Invalid: `person .name`, `person. name`, `array [1] = 2`.
Valid: `person.name`, `array[1] = 2`.

2. `WS` is not allowed following the unary operators `-` and `!`.

3. `WS` is used as the separator in expression lists in function call
arguments and array elements. `WS` is therefore _not_ allowed within
the expressions of an expression list, including the values of map
literal definitions.

4. `WS` is allowed within the expression of an expression list if the
expression is surrounded by `()`, `[]` or `{}`, `e.g. [ ( 2 + 3 ) ]`,
not `WS` directly after `[` and within `( … )`

5. `WS` can be freely used in single expressions for assignments,
inferred declarations, return statements, `if` conditions and `while`
conditions, as well as within parentheses `(…)`


Examples:

    print -5      // -5
    print - 5     // invalid
    print 2-1     // 1
    print 2 -1    // 2 -1
    print 2 - 1   // invalid
    a := 2 - 1    // valid!

    arr := ["a" "b"]
    print arr[1]        // b
    print arr [1]       // [a b] [1]
    arr [0] = "A"       // invalid
    arr2 :=[ 1   ]      // valid 
    arr3 := [[1][2]]    // valid
    arr3 := [[1] [ 2] ] // valid
    arr3 := [1 + 1 ]    // invalid

    m1 := { age:3+6 name:"mary"+"anne" }     // valid
    m2 := {age:  12 name:"mary"}             // valid
    m1.address = "10 Downing" + "Street"     // valid
    m3 := {address: "10 Downing" + "Street"} // invalid

    func add:num n1:num n2:num
        return n1 + n2 // valid
    end
    print (add 1 2)  // 3
    print add 1 2    // invalid

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

A variable may not use the same name as a function.

Function calls used as arguments to another function call must be
parenthesized to avoid ambiguity, for example:

    print "valid:" (is_valid "abc" 5)

Bare returns in functions without result types are allowed

    func validate m:{}any
        if m == {}
            return
        end
        // further validation
    end


## Variadic functions

A function with a single parameter may have a type suffixed with `...`.
A function with such a parameter is called variadic and may be invoked
with zero or more arguments for that parameter.

If `f` is variadic with a parameter `p` of type `T...`, then within `f`
the type of `p` is equivalent to type `[]T`. The length of the array is
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

## Break and Return

`break` and `return` are terminating statements. They interrupt the
regular flow of control. `break` is used to exit from the inner-most
loop body. `return` is used to exit from a function and may be followed
by an expression whose value is returned by the function call.

## Typeof

`typeof` returns the concrete type of a value held by a variable as a
string. It returns one of `"num"`, `"string"`, `"bool"`, `"array"`, or
`"map"`.

    typeof "abc"        // string
    typeof true         // bool
    arr := [ "abc" 1 ]
    typeof arr          // array
    typeof arr[0]       // string
    typeof arr[1]       // num
    typeof {}           // map

## Type assertion

A type assertion `ident.(type)` asserts that the value of the variable
`ident` is of the given `type`. If the assertion does not hold a
[run-time panic](#run-time-panics-and-recoverable-errors) occurs.

    x:any
    x = [ 1 2 3 4 ]  
    num_array := x.([]num)
    x = "abc"
    str := x.(string)

Only values of type `any` can be type asserted. That means an array of
type any, `[]any`, _cannot_ be type assert to be an array of type `num`
or other concrete type:

    x:[]
    x = [1 2]
    // x.([]num) // compile time error
    x[1] = [3 4 5]
    x[0].(num)    // valid
    x[0].(string) // run time panic

However, the elements of `x` can be type assert, e.g. `x[0].(num)`, 
`x[1].([]num)`.

## Event Handler

An event handler starts with `on`, followed by an event name and a block
of statements. The statements get executed when the given event is
triggered. Events can be triggered by user interaction, for example
clicking the mouse or tapping the keyboard or by the system, for
example `frame` when a new frame is painted.

There is a limited, predefined set of events. It is not possible to
create custom events.

    on mouse_down
        print mouse_x mouse_y
    end

    on frame
        draw
    end

The `frame` event is triggered every 2 Milliseconds, 50 times per
second.

## Run-time Panics and Recoverable Errors

Execution errors such as trying to index an array out of bounds or
access a map value for a key that does not exist or a failed type
assertion trigger a run-time panic. The execution of the `evy` program
stops and error details are printed.

A panic can be triggered with `panic "message"`.

Functions that can cause recoverable errors set the global string
variable `error` and the error classification number `errno`.
