# Conversion

## `str2num`

`str2num` converts a string to a number. If the string is not a valid
number, it returns `0` and sets the global `err` variable to `true`.

### Example

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

### Reference

    str2num:num s:string

The `str2num` function converts a string to a number. It takes a single
argument, which is the string to convert. If the string is a valid
number, the function returns the number. Otherwise, the function
returns 0 and sets the global `err` variable to `true`. For more
information on `err`, see the [Non-fatal Error section](#non-fatal-errors).

## `str2bool`

`str2bool` converts a string to a boolean. If the string is not a valid
boolean, it returns `false` and sets the global `err` variable to
`true`.

### Example

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

### Reference

    str2bool:bool s:string

The `str2bool` function converts a string to a bool. It takes a single
argument, which is the string to convert. The function returns `true`
if the string is equal to `"true"`, `"True"`, `"TRUE"`, or `"1"`, and
`false` if the string is equal to `"false"`, `"False"`, `"FALSE"`, or
`"0"`. The function returns `false` and sets the global `err` variable
to `true` if the string is not a valid boolean. For more information
on `err`, see the [Non-fatal Error section](#non-fatal-errors).
