# Evy Syntax by Example

The following examples on various aspects of `evy`'s syntax give an
intuitive understanding of the language. For a formal language
specification see the [syntax grammar].

[syntax grammar]: syntax_grammar.md

## Comment

    // This is comment

## Declaration

    x:num     // declaration: num, string, bool, any, [] {}
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
        print x        // nothing. step always 1 by default.
    end

### `for` … `range` array

    for x := range num[ 1 2 3 ]
        print x        // 1 2 3
    end

### `for` … `range` map

    m := string{ name:"Mali" sport:"climbing" }
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

    a1:num[]
    a2:string[][]
    a1 = num[ 1 2 3 4 ]
    a2 = string[ ["1" "2"] ["a" "b"] ]
    a3 := bool[ true false ]
    a4 := num[ "s1"
            "s2" ] // linebreak allowed
    a5 := any[ "chars" 123 ]

### Array element access

    print a1[1]    // 2
    print a2[1][0] // "a"

### Concatenation, append, prepend

    z = z + num[ 100 ] // [ 1 2 3 4 5 6 100 ]
    z = append z 101       // [ 1 2 3 4 5 6 100 101 ]
    z = prepend z 0        // [ 0 1 2 3 4 5 6 100 101 ]

### Slicing

    x1 := x[:2] // [ 1 2 ]
    x2 = x[2]   // [3]
    x2 = x[1:2] // [2]
    x2 = x[-1]  // [3]`
    x2 = x[-2:] // [ 2 3 ]

## Map

    m:any{}          // keys can only be identifiers
    m.name = "fox"
    m.age = 42

    m1 := string{ letters:"abc" nums:"asdf" }
    m2 := {} // short for any{}
    m3 := any{
            letters:"abc"
            nums:123
          } // linebreak allowed

### Map value access

    s := "letters"
    print m2.letters    // abc
    print m2[s]         // abc
    print m2["letters"] // abc

### Map value existence

    p := { name: "Le Petit Prince", pet: "sheep" }
    if has p "age" {
        print "age" p.age
    } 

## Any

    x:any  // any type
    m1:{}  // any map value type
    m2 := { letter:"a" number:1 }
    arr := any[ "b" 2 ]
   
## Type reflection

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

### Type reflection Usage Example

    v:any
    v = "asdf"
    if (reflect v).type == "string" 
        print "v is a string:" v
    end

## Type conversion

    b := str2bool "true" // true
    n := str2num "123"   // 123
    s1 := bool2str true  // "true"
    s2 := num2str 42     // "42"

## Type assertion

    x:any
    x = num[ 1 2 3 4 ]  // concrete type num[]
    s := num[] x

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

    on animate
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
    
    print  "abc" 123 // abc 123 \n
    prints "abc" 123 // print string: abc123
    printq "abc"     // print quoted, reuse as value: "abc"

### Strings
    
    "Hello"[2]                // "l"
    "Hello world"[1:5]        // "ello"
    join [ "one" "two" ] ", " // "one, two"
    split "hi there" " "      // [ "hi" "there" ]

### Length
 
    len for strings, arrays and maps

### Arrays

    append x 100   // [ 1 2 3 100 ]
    prepend x -100 // [ -100 1 2 3 100 ]

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
    format_time2 now "06/01/02 15:04"   // "22/01/02 15:04" 
    parse_time "2022-08-28T23:59:05Z"   // internal representation as unix seconds
    parse_time2 value format
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
    curve ?
    polygon num[ x1 y2 x2 y2 x3 y3 ]
    polyline num[ x1 y1 x2 y2 x3 y3 ]

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

