# String

## `sprint`

`sprint` stands for print to string.

It returns a string representation of the arguments given to it. It
separates them by a single space. Unlike `print`, there is no newline
added to the end.

### Example

```evy
s := sprint "a" [true] {a:1 b:2}
printf "%q\n" s
printf "%q\n" (sprint)
```

Output

```evy:output
"a [true] {a:1 b:2}"
""
```

### Reference

    sprint:string a:any...

The `sprint` function takes any number of arguments and returns a string
that represents them, separated by a single space. The arguments can be
of any type, including strings, numbers, booleans, and maps. Unlike the
`print` function, there is no newline added to the end of the string.

## `sprintf`

`sprintf` stands for print formatted to string.

`sprintf` returns a string representation of its arguments according to
a _format_ string. Formatting in `sprintf` and `printf` work the same
way, see [`printf`](#printf).

### Example

```evy
s := sprintf "%10q: %.f" "val" 123.45
print s
```

Output

```evy:output
     "val": 123
```

### Reference

    sprintf:string format:string a:any...

The `sprintf` function returns a string representation of its arguments
according to a _format_ string. The format string controls how the
arguments are formatted. The `sprintf` function works the same way as
the `printf` function, and the formatting syntax is the same, see
[`printf`](#printf).

## `join`

`join` concatenates the elements of an array of strings into a single
string, with the given separator string placed between elements.

### Example

```evy
s := join ["a" "b" "c"] ", "
print s
```

Output

```evy:output
a, b, c
```

### Reference

    join:string elems:[]string sep:string

The `join` function takes two arguments: an array of strings and a
separator string. The array of strings is the list of elements to be
concatenated. The separator string is the string that will be placed
between elements in the resulting string.

The `join` function returns a single string that is the concatenation of
the elements in the list of strings, with the separator string placed
between elements.

## `split`

`split` splits a string into a list of substrings separated by the given
separator. The separator can be any string, including the empty
string.

### Example

```evy
print (split "a,b,c" ",")
print (split "a,b,c" ".")
print (split "a,b,c" "")
```

Output

```evy:output
[a b c]
[a,b,c]
[a , b , c]
```

### Reference

    split:[]string s:string sep:string

The `split` function takes two arguments: the string to be split and the
separator string. The string to be split is the string that will be
split into substrings. The separator string is the string that will be
used to split the string.

The `split` function returns a list of substrings. The list of substrings
contains all of the substrings of the original string that are
separated by the separator string.

If the string does not contain the separator, the `split` function returns
an array of length 1 containing the original string.

If the separator is the empty string, the `split` function splits the
string after each character (UTF-8 sequence).

If both the string and the separator are empty, the `split` function
returns an empty list.

## `upper`

`upper` returns a string with all lowercase letters converted to
uppercase.

### Example

```evy
s := upper "abc D e ü"
print s
```

Output

```evy:output
ABC D E Ü
```

### Reference

    upper:string s:string

The `upper` function takes a single argument: the string to be converted
to uppercase. The function returns a new string with all lowercase
letters converted to uppercase. All other characters are left unchanged.

The `upper` function uses the Unicode character database to determine
which characters are lowercase and their equivalent uppercase form.

## `lower`

`lower` returns a string with all uppercase letters converted to
lowercase.

### Example

```evy
s := lower "abc D e ü"
print s
```

Output

```evy:output
abc d e ü
```

### Reference

    lower:string s:string

The `lower` function takes a single argument: the string to be converted
to lowercase. The function returns a new string with all uppercase
letters converted to lowercase. All other characters are left
unchanged.

The `lower` function uses the Unicode character database to determine
which characters are uppercase and their equivalent lowercase form.

## `index`

`index` returns the position of a substring in a string, or -1 if the
substring is not present.

### Example

```evy
n := index "abcde" "de"
print n
```

Output

```evy:output
3
```

### Reference

    index:num s:string sub:string

The `index` function finds the index of a substring `sub` in a string
`s`. It returns the index of the first occurrence of a `sub` within
`s`, or -1 if the substring is not present.

## `startswith`

`startswith` tests whether a string begins with a given prefix.

### Example

```evy
b := startswith "abcde" "ab"
print b
```

Output

```evy:output
true
```

### Reference

    startswith:bool s:string prefix:string

The `startswith` function tests whether the string `s` begins with
`prefix` and returns `true` if `s` starts with `prefix`, `false`
otherwise.

## `endswith`

`endswith` tests whether a string ends with a given suffix.

### Example

```evy
b := endswith "abcde" "ab"
print b
```

Output

```evy:output
false
```

### Reference

    endswith:bool s:string suffix:string

The `endswith` function tests whether the string `s` ends with `suffix`
and returns `true` if `s` ends with `suffix`, `false` otherwise.

## `trim`

`trim` removes leading and trailing characters from a string.

### Example

```evy
s := trim ".,..abc.de." ".,"
print s
```

Output

```evy:output
abc.de
```

### Reference

    trim:string s:string cutset:string

The `trim` function removes any characters in `cutset` from the
beginning and end of string `s`. It returns a copy of the resulting
string.

## `replace`

`replace` replaces all occurrences of a substring with another substring
in a string.

### Example

```evy
s := replace "abc123xyzabc abc" "abc" "ABC"
print s
```

Output

```evy:output
ABC123xyzABC ABC
```

### Reference

    replace:string s:string old:string new:string

The `replace` function replaces all occurrences of the substring `old`
in the string `s` with the substring `new`.
