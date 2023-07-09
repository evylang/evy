# Builtins

Evy provides built-in functions and events that allow for user
interaction, graphics, animation, mathematical operations, and more.

Functions are self-contained blocks of code that perform a specific
task. Events are notifications that are sent to a program when
something happens, such as when a user moves the mouse or presses a
key.

## Input and Output

### `print`

`print` prints the arguments given to it to the output area. It separates them by
a single space and outputs a newline character at the end.

#### Example 

```evy
print "Hello"
print 2 true "blue"
print "array:" [1 2 3]
print "map:" {name:"Scholl" age:21}
```
Output
```evy:output
Hello
2 true blue
array: [1 2 3]
map: {name:Scholl age:21}
```

#### Reference

    print a:any...

The `print` function prints its arguments to the output area, each
separated by a single space and terminated by a newline character. If
no arguments are provided it only prints the newline character. 

The backslash character `\` can be used to represent special characters
in strings. For example, the `\t` escape sequence represents a tab
character, and the `\n` escape sequence represents a newline character.
Quotes in string literals must also be escaped with backslashes,
otherwise they will be interpreted as the end of the string literal.
For example: 

```evy
print "Here's a tab: 👉\t👈\nShe said: \"Thank you!\""
```

Output
```evy:output
Here's a tab: 👉	👈
She said: "Thank you!"
```

In a browser environment `print` outputs to the output area. When
running Evy from the command line interface, `print` prints to standard
out.

### `read`
 
`read` reads a line of input from the user and returns it as a
string. The newline character is not included in the returned string.

#### Example
```evy
name := read
print "Hello, "+name+"!"
```

Input
```evy:input
Mary Jackson
```

Output
```evy:output
Hello, Mary Jackson!
```

#### Reference

    read:string

The `read` function returns a string that contains the line of input
that the user entered up until, excluding the newline character. It is
a blocking functions, which means that it will not return until the
user has entered a line of input and pressed the Enter key.

In a browser environment `read` reads from the text input area. When
running Evy from the command line interface, `read` reads from standard
in.

### `cls`

`cls` clears the output area of all printed text.

#### Example
```evy
print "Hello"
sleep 1
cls
print "Bye"
```

Output
```evy:output
Bye
```

#### Reference

    cls

The `cls` function clears all text output. In a browser environment
`cls` clears the output area. When running Evy from the command line
interface, `cls` clears the terminal, similar to the Unix `clear` or
Windows `cls` commands.

### `printf`

`printf` stands for print formatted.

`printf` prints its arguments to the output area according to a _format_
string. The format string is the first argument, and it
contains _specifiers_. Specifiers start with a percent sign `%`. They
tell the `printf` function how and where to print the remaining
arguments inside the format string. The rest of the format string is
printed to the output area without changes. 

Here are some valid specifiers in Evy:

| Specifier | Description |
| --------- | ------------- |
| `%v`      | the argument in its default format  |
| `%q`      | a double-quoted string |
| `%%`      | a percent sign `%`  |

#### Example

```evy
printf "The tank is 100%% full.\n\n"

weather := "rainy"
printf "It is %v today.\n" weather
rainfall := 10
printf "There will be %vmm of rainfall.\n" rainfall
unicorns := false
printf "There will be unicorns eating lollipops: %v.\n\n" unicorns

quote := "Wow!"
printf "They said: %q\n" quote
printf "Array: %v\n" [1 2 3]
printf "Map: %v\n" {a:1 b:2}
```

Output
```evy:output
The tank is 100% full.

It is rainy today.
There will be 10mm of rainfall.
There will be unicorns eating lollipops: false.

They said: "Wow!"
Array: [1 2 3]
Map: {a:1 b:2}
```

#### Reference

    printf format:string a:any...

The `printf` function prints its arguments to the output area according
to the _format_ string that is the first argument. The _specifiers_
that start with `%` and are contained in the format string are replaced
by the remaining arguments in the given order. For example, the
following code `printf "first: %s, second: %s" "A" "B"` prints `first:
A, second: B`.

Full list of valid specifiers in Evy:

| Specifier | Description |
| --------- | ------------- |
| `%v`      | the argument in a default format  |
| `%t`      | the word `true` or `false` |
| `%f`      | decimal point (floating-point) number, e.g. 123.456000 |
| `%e`      | scientific notation, e.g. -1.234456e+78 |
| `%s`      | string value |
| `%q`      | a double-quoted string |
| `%%`      | a literal percent sign `%`; consumes no value  |

If the arguments for the `%s`, `%q`, `%f`, `%e`, and `%t` specifiers do
not match the required type, a fatal runtime error will occur.

The _width_ and _precision_ of a floating-point number can be specified
with the `%f` and `%v` format specifiers.

- Width is the number of characters that will be used to print the
  number. If the width is not specified, it will be calculated based on
  the size of the number. It can be useful for padding and aligned
  output.
- Precision is the number of decimal places that will be displayed. If
  the precision is not specified, it will be set to 6 for `%f`. 

Here is a table that shows the different ways to specify the width and
precision of a floating-point number:

| Verb    | Description |
| ------- | ------------- |
| `%f`    | default width, default precision |
| `%7f`   | width 7, default precision |
| `%.2f`  | default width, precision 2 |
| `%7.2f` | width 7, precision 2 |
| `%7.f`  | width 7, precision 0 |

If the width/precision is preceded by a `-`, the value is padded with
spaces on the right rather than the left. If it is preceded by a 0, the
value is padded with leading zeros rather than spaces.

The width, precision and alignment prefix (`-` or `0`) can be used with
all valid specifiers. For example:

```evy
printf "right:  |%7.2f|\n" 1
printf "left:   |%-7.2v|\n" "abcd"
printf "zeropad:|%07.2f|\n" 1.2345
```
Output
```evy:output
right:  |   1.00|
left:   |ab     |
zeropad:|0001.23|
```

## Types

### `len`

`len` returns the number of characters in a string, the number of
elements in an array or the number of key-value pairs in a map.

#### Example

```evy
l := len "abcd"
print "len \"abcd\":" l

l = len [1 2]
print "len [1 2]:" l

l = len {a:3 b:4 c:5}
print "len {a:3 b:4 c:5}:" l
```

Output
```evy:output
len "abcd": 4
len [1 2]: 2
len {a:3 b:4 c:5}: 3
```

#### Reference

    len:num a:any

The `len` function takes a single argument, which can be a string, an
array, or a map. If the argument is a string, `len` returns the number
of characters in the string. If the argument is an array, `len` returns
the number of elements in the array. If the argument is a map, `len`
returns the number of key-value pairs in the map. If the argument is of
any other type, a fatal runtime error will occur.

### `typeof`

`typeof` returns the type of the argument as string value.

#### Example

```evy
a:any
a = "abcd"
t := typeof a
print "typeof \"abcd\":" t

t = typeof {kind:true strong:true}
print "typeof {kind:true strong:true}:" t

t = typeof [[1 2] [3 4]]
print "typeof [[1 2] [3 4]]:" t

t = typeof [1 2 true]
print "typeof [1 2 true]:" t

print "typeof []:" (typeof [])
```

Output
```evy:output
typeof "abcd": string
typeof {kind:true strong:true}: {}bool
typeof [[1 2] [3 4]]: [][]num
typeof [1 2 true]: []any
typeof []: []
```

#### Reference

    typeof:string a:any

The `typeof` function takes a single argument, which can be of any type.
The function returns a string that represents the type of the argument.
The string returned by `typeof` is the same as the type in an Evy
program, for example `num`, `bool`, `string`, `[]num`, `{}[]any`. For
an empty composite literal, `typeof` returns `[]` or `{}` as it can be
matched to any subtype, e.g. `[]` can be passed to a function that
takes an argument of `[]num`, or `[]string`.

## Map

### `has`

`has` returns whether a map has a given key or not.

#### Example

```evy
map := {a:1}
printf "has %v %q: %t\n" map "a" (has map "a")
printf "has %v %q: %t\n" map "X" (has map "X")
```

Output
```evy:output
has {a:1} "a": true
has {a:1} "X": false
```

#### Reference

    has:bool map:{} key:string

The `has` function takes two arguments: a map and a key. It returns true
if the map has the key, and false if the map does not have the key. The
map can be of any value type, such as `{}num` or `{}[]any` and the key
can be any string.

### `del`

`del` deletes a key-value entry from a map.

#### Example

```evy
map := {a:1 b:2}
del map "b"
print map
```

Output
```evy:output
{a:1}
```

#### Reference

    del map:{} key:string

The `del` function takes two arguments: a map and a key. It deletes the
key-value entry from the map if the key exists. If the key does not
exist, the function does nothing. The map can have any value type, and
the key can be any string.

## Program control

### `sleep`

`sleep` pauses the program for the given number of seconds.

`sleep` can be used to create delays in Evy programs. For example, you
could use sleep to create a countdown timer.

#### Example

```evy
print "2"
sleep 1
print "1"
```

Output
```evy:output
2
1
```

#### Reference

    sleep seconds:num

The `sleep` function pauses the execution of the current Evy program for
at least the given number of seconds. Sleep may also pause for a
fraction of a second, e.g. `sleep 0.1`.

### `exit`

`exit` terminates the program with the given status code.

#### Example

```evy
input := "not a number"
n := str2num input

if err
    print errmsg
    exit 1
end

print n
```

Output
```evy:output
str2num: cannot parse "not a number"
```

#### Reference

    exit status:num

The `exit` function takes a single argument, which is the status code
that the program will terminate with. The status code can be any
number, but it is typically used to indicate whether the program
terminated successfully or with an error. A status code of 0 means that
the program terminated successfully, while any other status code is
considered an error.

## Conversion

### `str2num`

`str2num` converts a string to a number. If the string is not a valid
number, it returns `0` and sets the global `err` variable to `true`.

#### Example

```evy
n:num
n = str2num "1"
print "n:" n "err:" err
n = str2num "NOT-A-NUMBER"
print "n:" n "err:" err
```

Output
```evy:output
n: 1 err: false
n: 0 err: true
```

#### Reference

    str2num:num s:string

The `str2num` function converts a string to a number. It takes a single
argument, which is the string to convert. If the string is a valid
number, the function returns the number. Otherwise, the function
returns 0 and sets the global `err` variable to `true`. For more
information on `err`, see the [Non-fatal Error section](#non-fatal-errors).

### `str2bool`

`str2bool` converts a string to a boolean. If the string is not a valid
boolean, it returns `false` and sets the global `err` variable to
`true`.

#### Example

```evy
b:bool
b = str2bool "true"
print "b:" b "err:" err
b = str2bool "NOT-A-BOOL"
print "b:" b "err:" err
```

Output
```evy:output
b: true err: false
b: false err: true
```

#### Reference

    str2bool:bool s:string

The `str2bool` function converts a string to a bool. It takes a single
argument, which is the string to convert. The function returns `true`
if the string is equal to `"true"`, `"True"`, `"TRUE"`,  or `"1"`, and
`false` if the string is equal to `"false"`, `"False"`, `"FALSE"`, or
`"0"`. The function returns `false` and sets the global `err` variable
to `true` if the string is not a valid boolean. For more information
on `err`, see the [Non-fatal Error section](#non-fatal-errors).

## Error

Evy programs in execution can report two types of errors:

- Fatal errors
- Non-fatal errors

### Fatal Errors

Fatal errors cause the program to exit immediately with an error
message. They typically occur when the program encounters a situation
that it cannot handle, such as trying to access an element of an array
that is out of bounds. Fatal errors cannot be intercepted by the
program, so it is important to take steps to prevent them from
occurring in the first place. 

One way to do this is to use _guarding code_, which is code that checks
for potential errors and takes steps to prevent them from occurring.
For example, guarding code could be used to check the length of an
array before trying to access an element to avoid an out of bounds
error. If the access index is out of bounds, the guarding code could
report the error.

Here is an example of a fatal error:

```evy
arr := [0 1 2]
i := 5 // e.g. user input
print arr[i] // out of bounds
print "This line will not be executed"
```

This code will cause a fatal error because the index 5 is out of bounds
for the array `arr`. The program will exit with the error message 

```
line 3: index out of bounds: 5
```

### Non-fatal Errors

The global `err` variable is used to indicate whether a non-fatal
error has occurred. The global `errmsg` variable stores a detailed
message about the error that occurred.

Non-fatal errors are caused by code that could not be prevented from
running, such as converting a user input string to a number if the
string is not a number. Non-fatal errors will set the global `err`
variable to `true` and the program will continue executing. If there is
no error, `err` is set to `false`.

The global `errmsg` variable stores a detailed message about the error
which is set alongside `err`. `errmsg` is set to the empty string `""`
if no error has occurred.  If an error does occur, `errmsg` is set to a
message that describes the error.

When a function that could potentially cause an error finishes executing
without an error, the `err` variable is reset to `false` and the
`errmsg` variable is set to the empty string. This is done even if the
`err` variable was previously set to `true` or the `errmsg` variable
was not empty. Therefore, it is up to the program to check the `err`
variable after any possible error occurrence.

Here is an example of a non-fatal error:

```evy
n := str2num "NOT A NUM"
print "num:" n
print "err:" err
print "errmsg:" errmsg
```

Output
```evy:output
num: 0
err: true
errmsg: str2num: cannot parse "NOT A NUM"
```
