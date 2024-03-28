# Syntax by Example

The following examples will help you understand the syntax of Evy. For a
more formal definition of the syntax, see the
[Language Specification](spec.md). Built-in functions, such as `print`
and `circle`, are documented in the [Built-ins section](builtins.md).

## Comment

```evy
// This is a comment
```

## Declaration

```evy
x:num // or string, bool, any, []num, {}string
y := 1 // type inference (num)
print x y // 0 1
```

## Assignment

```evy
z:num
z = 5
print z // 5
```

## Expression

Arithmetic, number expressions

```evy
pi := 3.1415
x := 5 * (3 + pi) - 2 / 7.6
print x // 30.44434210526316
```

Logical, boolean expressions

```evy
trace := false
debug := true
level := "error"

b := !trace and debug or level == ""
print b // true
```

## Strings

Concatenation, indexing and slicing

```evy
str := "abc" + "🥪123" // "abc🥪123" - concatenation
s2 := str[0] // "a" - indexing
s3 := str[1:5] // "bc🥪1" - slicing
print str s2 s3
```

Newline, indentation and escaping

```evy
str := "newline: \n indentation: \t"
print str
print "quotation mark : \" " // escaping
```

## `if` statements

```evy
x := 6
if x > 10
    print "huge"
else if x > 5
    print "medium"
else
    print "small"
end
```

### Nested `if`

```evy
str := "abc"
if (len str) > 2
    if (startswith str "a")
        print "string starting with 'a'"
    else
        print "string not starting with 'a'"
    end
else
    print "single character or empty string"
end
```

## Loop statements

### `while` loop

```evy
x := 0
while x < 10
    print x // 0 1 2 ... 9
    x = x + 1
end
```

### `for` … `range` number

```evy
for x := range 5
    print x // 0 1 2 3 4
end

for x := range 5 10
    print x // 5 6 7 8 9
end

for x := range 1 10 2 // from to step
    print x // 1 3 5 7 9
end

for x := range -10
    print x // nothing. step is 1 by default.
end
```

### `for` … `range` array

```evy
for x := range [1 2 3]
    print x // 1 2 3
end
```

### `for` … `range` map

```evy
m := {name:"Mali" sport:"climbing"}
for key := range m
    print key m[key]
end
```

### `break`

```evy
x := 0
while true
    print "tick... "
    sleep 1
    if x > 2
        print "💥"
        break // breaks out of the innermost loop
    end
    x = x + 1
end
```

## Function definition

```evy
func add:num a:num b:num
    return a + b
end
```

### No return type

```evy
func foxprint s:string
    print "🦊 "+s
end
```

### Variadic

```evy
func list args:any...
    for arg := range args[:-1]
        printf "%v, " arg
    end
    printf "%v" args[-1]
end
```

### Function calls

```evy
n := add 1 2
print n // 3
foxprint "🐾" // 🦊 🐾
list 2 true "blue" // [2 true blue]

// previous function definitions
func add:num a:num b:num
    return a + b
end

func foxprint s:string
    print "🦊 "+s
end

func list args:any...
    print args
end
```

## Array

Typed declaration

```evy
a1:[]num
a2:[][]string
a1 = [1 2 3 4] // type: num[]
a2 = [["1" "2"] ["a" "b"]] // type: string[][]
print a1 a2
```

Declaration with inference

```evy
a1 := [true false] // type: bool[]
a2 := ["s1" // line break allowed
    "s2"] // type: string[]
print a1 a2
```

`any` arrays

```evy
a1:[]any
a2 := ["chars" 123] // type: any[]
print a1 a2
```

### Array element access

```evy
a1 := [1 2 3 4]
a2 := [["1" "2"] ["a" "b"]]
print a1[1] // 2
print a2[1][0] // "a"
print a1[-1] // 4
```

### Concatenation

```evy
a := [1 2 3 4]
a = a + [100] // [1 2 3 4 100]; optional extra whitespace
a = [0] + a + [101 102] // [0 1 2 3 4 100 101 102]
```

### Slicing

```evy
a := [1 2 3]
b := a[:2] // [1 2]
b = a[1:2] // [2]
b = a[-2:] // [2 3]
```

## Map

Any map

```evy
m:{}any // keys used in literals or with `.` must be identifiers.
m.name = "fox"
m.age = 42
m["key with space"] = "🔑🪐"
print m // {name:fox age:42 key with space:🔑🪐}
```

Typed map

```evy
m1 := {letters:"abc" name:"Jill"} // type: {}string
m2 := {
    letters:"abc" // line break allowed
    name:"Jill"
}
print m1 m2
```

Empty map

```
m1:{}string // {}string
m2 := {} // {}any
print m1 m2 // {} {}
```

Nested map

```
m1:{}[]num
m2 := {a:{}}
print m1 m2 // {} {a:{}}
```

### Map value access

```evy
m := {letters:"abc" name:"Jill"}
s := "letters"
print m.letters // abc
print m[s] // abc
print m["letters"] // abc
```

## `any`

Zero value of any is `false`.

```evy
x:any
m1:{}any
m2 := {letter:"a" number:1} // {}any
print x m1 m2 // false {} {letter:a number:1}

a1:[]any
a2 := ["b" 2] // []any
print a1 a2 // [] [b 2]
```

## Type inspection with `typeof`

```evy
print (typeof "abc") // "string"
print (typeof true) // "bool"
print (typeof [1 2]) // "[]num"
print (typeof [[1 2] [3 4]]) // "[][]num"
```

## Type assertion

```evy
x:any
print x (typeof x) // flase bool
x = [1 2 3 4]
s := x.([]num) // type assertion
print s (typeof s) // [1 2 3 4] []num
```

## Type inspection and assertion

```evy
v:any
v = "🐐"
if (typeof v) == "string"
    s := v.(string) // type assertion
    print s+s // 🐐🐐
end
```

## Event handling

```evy
on key k:string
    print "key:" k
end
```

Evy can only handle a limited set of events, such as key presses,
pointer movements, or periodic screen redraws.
