# Types

## `len`

`len` returns the number of characters in a string, the number of
elements in an array or the number of key-value pairs in a map.

### Example

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

### Reference

    len:num a:any

The `len` function takes a single argument, which can be a string, an
array, or a map. If the argument is a string, `len` returns the number
of characters in the string. If the argument is an array, `len` returns
the number of elements in the array. If the argument is a map, `len`
returns the number of key-value pairs in the map. If the argument is of
any other type, a fatal runtime error will occur.

## `typeof`

`typeof` returns the type of the argument as string value.

### Example

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

### Reference

    typeof:string a:any

The `typeof` function takes a single argument, which can be of any type.
The function returns a string that represents the type of the argument.
The string returned by `typeof` is the same as the type in an Evy
program, for example `num`, `bool`, `string`, `[]num`, `{}[]any`. For
an empty composite literal, `typeof` returns `[]` or `{}` as it can be
matched to any subtype, e.g. `[]` can be passed to a function that
takes an argument of `[]num`, or `[]string`.
