# Map

## `has`

`has` returns whether a map has a given key or not.

### Example

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

### Reference

    has:bool map:{} key:string

The `has` function takes two arguments: a map and a key. It returns true
if the map has the key, and false if the map does not have the key. The
map can be of any value type, such as `{}num` or `{}[]any` and the key
can be any string.

## `del`

`del` deletes a key-value entry from a map.

### Example

```evy
map := {a:1 b:2}
del map "b"
print map
```

Output

```evy:output
{a:1}
```

### Reference

    del map:{} key:string

The `del` function takes two arguments: a map and a key. It deletes the
key-value entry from the map if the key exists. If the key does not
exist, the function does nothing. The map can have any value type, and
the key can be any string.
