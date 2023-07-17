# Program control

## `sleep`

`sleep` pauses the program for the given number of seconds.

`sleep` can be used to create delays in Evy programs. For example, you
could use sleep to create a countdown timer.

### Example

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

### Reference

    sleep seconds:num

The `sleep` function pauses the execution of the current Evy program for
at least the given number of seconds. Sleep may also pause for a
fraction of a second, e.g. `sleep 0.1`.

## `exit`

`exit` terminates the program with the given status code.

### Example

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

### Reference

    exit status:num

The `exit` function takes a single argument, which is the status code
that the program will terminate with. The status code can be any
number, but it is typically used to indicate whether the program
terminated successfully or with an error. A status code of 0 means that
the program terminated successfully, while any other status code is
considered an error.
