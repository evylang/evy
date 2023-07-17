# Math

## `min`

`min` returns the smaller of the two given numbers.

### Example

```evy
print (min 3 1)
```

Output

```evy:output
1
```

### Reference

    min:num n1:num n2:num

The `min` function returns the smaller of the two given number
arguments.

## `max`

`max` returns the greater of the two given numbers.

### Example

```evy
print (min 3 1)
```

Output

```evy:output
1
```

### Reference

    max:num n1:num n2:num

The `max` function returns the greater of the two given number
arguments.

## `floor`

`floor` returns the greatest integer value less than or equal to the given
number.

### Example

```evy
print (floor 2.7)
print (floor 3)
```

Output

```evy:output
2
3
```

### Reference

    floor:num n:num

The `floor` function returns the greatest integer value less than or
equal to its number argument `n`.

## `ceil`

`ceil` returns the smallest integer greater than or equal to the given
number.

### Example

```evy
print (ceil 2.1)
print (ceil 4)
```

Sample output

```
3
4
```

### Reference

    ceil:num n:num

The `ceil` function returns the smallest integer greater than or equal
to its number argument `n`.

## `round`

`round` returns the nearest integer to the given number, rounding half
away from 0.

### Example

```evy
print (round 2.4)
print (round 2.5)
```

Sample output

```
2
3
```

### Reference

    round:num n:num

The `round` function returns the nearest integer to the given number
argument `n`, rounding half away from 0.

## `pow`

`pow` returns the value of the first number raised to the power of the
second number.

### Example

```evy
print (pow 2 3)
```

Output

```evy:output
8
```

### Reference

    pow:num b:num exp:num

The `pow` function returns `b` to the power of `exp`. The first number
argument `b` is the base. The second number argument `exp` is the
exponent.

## `log`

`log` returns the logarithm of the given number, to the base of e.

### Example

```evy
printf "%.2f\n" (log 1)
printf "%.2f\n" (log 2.7183) // e
```

Output

```evy:output
0.00
1.00
```

### Reference

    log:num n:num

The `log` function returns the _natural logarithm_, the logarithm of the
given number argument `n`, to the base of `e`.

## `sqrt`

`sqrt` returns the square root of the given number.

### Example

```evy
print (sqrt 9)
```

Output

```evy:output
3
```

### Reference

    sqrt:num n:num

The `sqrt` function returns the positive square root of the number
argument `n`.

## `sin`

`sin` returns the sine of the given angle in radians.

### Example

```evy
pi := 3.14159265
print (sin 0.5*pi)
```

Output

```evy:output
1
```

### Reference

    sin:num n:num

The `sin` function returns the sine of the given angle `n` in radians.

## `cos`

`cos` returns the cosine of the given angle in radians.

### Example

```evy
pi := 3.14159265
print (cos pi)
```

Output

```evy:output
-1
```

### Reference

    cos:num n:num

The `cos` function returns the cosine of the given angle `n` in radians.

## `atan2`

`atan2` returns the angle in radians between the positive x-axis and the
ray from the origin to the point `x y`.

### Example

```evy
pi := 3.14159265
rad := atan2 1 1
degrees := rad * 180 / pi
printf "rad: %.2f degrees: %.2f" rad degrees
```

Output

```evy:output
rad: 0.79 degrees: 45.00
```

### Reference

    atan2:num y:num x:num

The `atan2` function returns the angle in radians between the positive
x-axis and the ray from the origin to the point `x y`. More formally,
it returns the arc tangent of `y/x` for given arguments `y` and `x`,
using the signs of the two to determine the quadrant of the return
value.
