# Error

Evy programs in execution can report two types of errors:

- Fatal errors
- Non-fatal errors

## Fatal Errors

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

## Non-fatal Errors

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
if no error has occurred. If an error does occur, `errmsg` is set to a
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
