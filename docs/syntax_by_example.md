# Evy Syntax by Example

The following examples will help you understand the syntax of Evy. For a
more formal language specification, please see the [syntax grammar](syntax_grammar.md).

## Comment

    // This is a comment

## Declaration

    x:num     // declaration: num, string, bool, any, []num, {}string
    y := 1    // declaration through type inference (num)

## Assignment

    z = 5

## Expression

    x := 5 * (y + z)  - 2 / 7.6           // arithmetic number expression
    b := !trace and debug or level == ""  // bool expressions

## Strings

    s1 := "quotation mark : \" "          // escaping
    s2 := "abc" + "🥪123"                 // concatenation
    s3 := "newline: \n indentation: \t"
    s4 := s2[0]                           // "a"
    s5 := s2[1:5]                         // "bc🥪1"

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
        print "tick... "
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

    func list args:any...
        for arg := range args[:-1]
            printf "%v, " arg
        end
        printf "%v" args[-1]
    end

### Function calls

    n := add 1 2        // 3
    foxprint "🐾"       // 🦊 🐾  
    list 2 true "blue"  // 2, true, blue

## Array

    a1:[]num
    a2:[][]string
    a1 = [1 2 3 4]              // type: num[]
    a2 = [["1" "2"] ["a" "b"]]  // type: string[][]
    a3 := [true false]          // type: bool[]
    a4 := ["s1"                 // line break allowed
           "s2"]                // type: string[]
    a5 := ["chars" 123]         // type: any[]
    a6:[]any                    // type: any[]

### Array element access

    a1 := [1 2 3 4]
    a2 := [["1" "2"] ["a" "b"]]
    print a1[1]                  // 2
    print a2[1][0]               // "a"
    print a1[-1]                  // 4

### Concatenation

    a := [1 2 3 4]
    a = a + [ 100 ]          // [1 2 3 4 100]; optional extra whitespace 
    a = [0] + a + [101 102]  // [0 1 2 3 4 100 101 102]

### Slicing

    a := [1 2 3]
    b := a[:2]         // [1 2]
    b = a[1:2]         // [2]
    b = a[-2:]         // [2 3]

## Map

    m1:{}any // keys used in literals or with `.` must be identifiers.
    m1.name = "fox"
    m1.age = 42
    m1["key with space"] = "🔑🪐"
    
    m2 := {letters:"abc" name:"Jill"} // type: {}string
    m3 := {}                          // type: {}any
    m4 := {
        letters:"abc"                 // line break allowed
        nums:123
    }                                 // type: {}any
    m5:{}[]num                        // map of array of numbers
    m5.digits = [1 2 3]
    m6:{}num
    //m6.x = "y"                      // invalid, only num values allowed

### Map value access

    m := {letters:"abc" name:"Jill"}   
    s := "letters"
    print m.letters    // abc
    print m[s]         // abc
    print m["letters"] // abc

## `any`

    x:any     // any type, default value: false
    m1:{}any  // map with any value type
    m2 := { letter:"a" number:1 }
    arr1:[]any
    arr2 := [ "b" 2 ]
   
## Type assertion

    x:any
    x = [ 1 2 3 4 ]  // concrete type num[]
    s := x.([]num)

## Type reflection

    typeof "abc"          // "string"
    typeof true           // "bool"
    typeof [ 1 2 ]        // "[]num"
    typeof [[1 2] [3 4]]  // "[][]num"

    v:any
    v = "🐐"
    if (typeof v) == "string"
        print "v is a string:" v
        s := v.(string) // type assertion        
        print s+s       // 🐐🐐 
    end

## Event handling 

    on key
        print "key pressed"
    end

Evy can only handle a limited set of events, such as key presses, mouse
movements, or periodic screen redraws.

### Event handlers with parameters

    on key k:string
        printf "%q pressed\n" k
    end

