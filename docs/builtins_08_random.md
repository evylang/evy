# Random

## `rand`

`rand` returns a random, non-negative integer less than the argument.

### Example

```evy
print (rand 3)
print (rand 3)
```

Sample output

```
2
0
```

### Reference

    rand:num n:num

The `rand` functions returns, a non-negative pseudo-random integer
number in the half-open interval `[0,n)`. A fatal runtime error occurs
for `n <= 0`.

## `rand1`

`rand1` returns a random, non-negative floating point number less than 1.

### Example

```evy
print (rand1)
print (rand1)
```

Sample output

```
0.7679753163102002
0.6349044894123325
```

### Reference

    rand1:num

The `rand1` function returns a pseudo-random floating point number in
the half-open interval `[0.0,1.0)`.
