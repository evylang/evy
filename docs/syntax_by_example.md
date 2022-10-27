# Evy Syntax by Example

The following examples give an intuitive understanding of various
aspects of `evy`'s syntax. For a formal language specification see the
[syntax grammar](syntax_grammar.md).


## Comment

    // This is comment

## Declaration

    x:num     // declaration: num, string, bool, any, []num, {}string
    y := 1    // declaration through type inference (num)

## Assignment

    z = 5

## Expression

    x := 5 * (x + z)  - 2 / 7.6           // arithmetic number expression
    b := !trace and debug or level == ""  // bool expressions

## Strings

    s1 := "quotation mark : \" " // escaping
    s2 := "abc" + "🥪123"          // concatenation
    s3 := "newline: \n indentation: \t"
    s4 := s2[0]                  // "a"
    s5 := s2[1:5]                // "bc🥪1"

## `if` statements

    if z > 0 and x != 0
        print "block 1"
    else if y != 0 or a == "abc"
        print "block 2"
    else
        print "block 3"
    end

### Nested `if`

    if z > 0 and x != 0
        if startswith str "a"
            print "nested block 1"
        else
            print "nested block 2"
        end
    end


## Loop statements

### `while` loop

    x := 0
    while x < 10
        print x
        x = x + 1
    end

### `for` … `range` number

    for x := range 5
        print x           // 0 1 2 3 4
    end

    for x := range 5 10
        print x           // 5 6 7 8 9
    end

    for x := range 1 10 2 // from to step
        print x           // 1 3 5 7 9
    end

    for x := range -10 
        print x        // nothing. step is 1 by default.
    end

### `for` … `range` array

    for x := range [1 2 3]
        print x        // 1 2 3
    end

### `for` … `range` map

    m := { name:"Mali" sport:"climbing" }
    for key := range m
        print key m[key]
    end

### `break`

    x := 0
    while true
        prints "tick..."
        sleep 1
        if x > 9
            print "💥"
            break  // `break` breaks out of the innermost loop
        end
        x = x + 1
    end


## Function definition

    func add:num a:num b:num
        return a + b
    end

### No return type

    func foxprint s:string
        print "🦊 " + s
    end

### Variadic

    func my_print args:any...
        for arg := range args
            write arg
            write " "
        end
    end

## Array

    a1:[]num
    a2:[][]string
    a1 = [1 2 3 4]             // type: num[]
    a2 = [["1" "2"]["a" "b"]]  // type: string[][]
    a3 := [true false]         // type: bool[]
    a4 := ["s1"                // line break allowed
           "s2"]               // type: string[]
    a5 := ["chars" 123]        // type: any[]
    a6:[]any                   // type: any[]

### Array element access

    a1 := [1 2 3 4]
    a2 := [["1" "2"]["a" "b"]]
    print a1[1]    // 2
    print a2[1][0] // "a"

### Concatenation, append, prepend

    z = z + [ 100 ]    // z: [1 2 3 4 5 6 100]; optional extra whitespace 
    z = z + [101]      // z: [1 2 3 4 5 6 100 101]
    append z 102       // z: [1 2 3 4 5 6 100 101 102]
    prepend z 0        // z: [0 1 2 3 4 5 6 100 101 102]

### Slicing

    x := [1 2 3]
    x1 := x[:2] // [1 2]
    x2 = x[2]   // [3]
    x2 = x[1:2] // [2]
    x2 = x[-1]  // [3]
    x2 = x[-2:] // [2 3]

## Map

    m1:{}any          // keys can only be identifiers, any value allowed
    m2.name = "fox"
    m2.age = 42

    m3 := {letters:"abc" name:"Jill"}   // type: {}string
    m4 := {}                            // type: {}any
    m5 := {
            letters:"abc"               // line break allowed
            nums:123
          }                             // type: {}any
    m6:{}[]num                          // map of array of numbers
    m6.digits = [1 2 3]
    m7:{}num
    m7.x = "y"                          // invalid, only num values allows

### Map value access

    m := {letters:"abc" name:"Jill"}   
    s := "letters"
    print m.letters    // abc
    print m[s]         // abc
    print m["letters"] // abc

### Map value existence

    p := { name: "Le Petit Prince", pet: "sheep" }
    if has p "age" {
        print "age" p.age
    } 

## Any

    x:any     // any type
    m1:{}any  // map with any value type
    m2 := { letter:"a" number:1 }
    arr1:[]any
    arr2 := [ "b" 2 ]
   
## Type reflection

    reflect "abc"         // {type: "string"}
    reflect true          // {type: "bool"}
    reflect [ 1 2 ]       // {type: "array",
                          //  sub:  {type: "num"}
                          // }
    reflect [[1 2] [3 4]] // {
                          //   type: "array",
                          //   sub:  {
                          //     type: "array"
                          //     sub: {
                          //       type: "num"
                          //     }
                          //   }
                          // }

### Type reflection Usage Example

    v:any
    v = "asdf"
    if (reflect v) == {type: "string"}
        print "v is a string:" v
    end

## Type conversion

    b := str2bool "true" // true
    n := str2num "123"   // 123
    s1 := bool2str true  // "true"
    s2 := num2str 42     // "42"

## Type assertion

    x:any
    x = [ 1 2 3 4 ]  // concrete type num[]
    s := x.([]num)

## Variadic functions

    func addmany:num arr:num...
        result := 0
        for x := range arr
            result = result + x
        end
        return result
    end

    print addmany 1 2 3
    arr := [ 4 5 6 ]
    print (addmany arr...)

## Event handling 

    on frame
        draw
    end

    on key_press 
        print key
    end

    on mouse_down
        print mouse_x mouse_y
    end

## Builtin

### Print
    
    print  "abc" 123 // abc 123\n
    prints "abc" 123 // print string: abc123
    printq "abc" 123  // print quoted, reuse as value: "abc"

Returning a string:

    sprint  "abc" 123 // returns "abc 123\n"
    sprints "abc" 123 // returns "abc123"
    sprintq "abc"     // returns "\"abc\""

### Strings
    
    "Hello"[2]                // "l"
    "Hello world"[1:5]        // "ello"
    join [ "one" "two" ] ", " // "one, two"
    split "hi there"          // [ "hi" "there" ]

### Length
 
    len for strings, arrays and maps

### Arrays

    append x 100   // [ 1 2 3 100 ]
    prepend x -100 // [ -100 1 2 3 100 ]

### Maps
 
    m := {name: "abc"}
    has m "abc" // true
    del m "abc"
    has m "abc" // false

### Conversion

    str2bool "true" // true
    str2num "123"   // 123
    num2str 123     // "123"
    bool2str false  // "false"

### Error

    error  // global string error message of last error
    errnum // error num to check for error type 0 ... 10 reserved 
           // conversion error, index out of bounds, assertion error,
    panic "error message" // terminates the program and prints "error message"

### Time
    
    now                                 // return unix time in seconds
    format_time now                     // "2022-08-28T23:59:05Z" Z is time zone zero
    format_timef now "06/01/02 15:04"   // "22/01/02 15:04" 
    parse_time "2022-08-28T23:59:05Z"   // internal representation as unix seconds
    parse_timef value format
    sleep 10                            // sleep 10 seconds

See Go's [time.Layout] for further details on formatting and parsing.

[time.Layout]: https://pkg.go.dev/time#pkg-constants

### Events
    
    mouse_down mouse_up mouse_move
    key_press

Used in event handlers, e.g. `on mouse_down`.
No support for custom events.

### UI

    move x y

    circle radius
    line end_x end_y
    rect width height
    curve // TBD
    polygon [x1 y2] [x2 y2] [x3 y3]
    polyline [x1 y1] [x2 y2] [x3 y3]

    color  900   // CSS    #ff0000
    colors "red" // CSS color keywords:     #ff0000
    linewidth 1

    text "some text"
    textsize 12

### Math
   
    div 7 3 // integer division
    pow 2 3 // exponentiation
    sqrt 2
    logn 10
    sin 45
    asin 0.707
    cos 45
    acos 0.707
    tan 45
    atan 1
    atan 100
    atan2 100 0
    pi
    abs -21.34
    floor 2.15
    number "114.2"
    random 10
    randomf // [0 1.0)

### Read

    str := read
    str := readln
    str := readid query_selector // event.target.value

