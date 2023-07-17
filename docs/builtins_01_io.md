# Input and Output

## `print`

`print` prints the arguments given to it to the output area. It separates them by
a single space and outputs a newline character at the end.

### Example

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

### Reference

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

## `read`

`read` reads a line of input from the user and returns it as a
string. The newline character is not included in the returned string.

### Example

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

### Reference

    read:string

The `read` function returns a string that contains the line of input
that the user entered up until, excluding the newline character. It is
a blocking functions, which means that it will not return until the
user has entered a line of input and pressed the Enter key.

In a browser environment `read` reads from the text input area. When
running Evy from the command line interface, `read` reads from standard
in.

## `cls`

`cls` clears the output area of all printed text.

### Example

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

### Reference

    cls

The `cls` function clears all text output. In a browser environment
`cls` clears the output area. When running Evy from the command line
interface, `cls` clears the terminal, similar to the Unix `clear` or
Windows `cls` commands.

## `printf`

`printf` stands for print formatted.

`printf` prints its arguments to the output area according to a _format_
string. The format string is the first argument, and it
contains _specifiers_. Specifiers start with a percent sign `%`. They
tell the `printf` function how and where to print the remaining
arguments inside the format string. The rest of the format string is
printed to the output area without changes.

Here are some valid specifiers in Evy:

| Specifier | Description                        |
| --------- | ---------------------------------- |
| `%v`      | the argument in its default format |
| `%q`      | a double-quoted string             |
| `%%`      | a percent sign `%`                 |

### Example

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

### Reference

    printf format:string a:any...

The `printf` function prints its arguments to the output area according
to the _format_ string that is the first argument. The _specifiers_
that start with `%` and are contained in the format string are replaced
by the remaining arguments in the given order. For example, the
following code `printf "first: %s, second: %s" "A" "B"` prints `first:
A, second: B`.

Full list of valid specifiers in Evy:

| Specifier | Description                                            |
| --------- | ------------------------------------------------------ |
| `%v`      | the argument in a default format                       |
| `%t`      | the word `true` or `false`                             |
| `%f`      | decimal point (floating-point) number, e.g. 123.456000 |
| `%e`      | scientific notation, e.g. -1.234456e+78                |
| `%s`      | string value                                           |
| `%q`      | a double-quoted string                                 |
| `%%`      | a literal percent sign `%`; consumes no value          |

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

| Verb    | Description                      |
| ------- | -------------------------------- |
| `%f`    | default width, default precision |
| `%7f`   | width 7, default precision       |
| `%.2f`  | default width, precision 2       |
| `%7.2f` | width 7, precision 2             |
| `%7.f`  | width 7, precision 0             |

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
