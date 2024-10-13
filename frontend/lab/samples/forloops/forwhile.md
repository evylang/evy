# 👯 `for` as `while` loops

⭐ Can you re-write the program using `while` instead of `for` loops?

Make sure you still generate the same output.

### [>] Hint

```evy
x:num

x = ❓ // START
while x < ❓ // STOP
  print x
  x = x + ❓ // STEP
end
```

# [>] Docs

## `for` loops

`for` loops, also known as **count-controlled** loops, are a shortcut for
writing certain `while` loops (**condition-controlled** loops).

Every `for` loop can be written as a `while` loop, but not every `while` loop
can be written as a `for` loop.

`for` loops have the following structure:

```evy
for VAR := range START STOP STEP
  // code block
end
```

This loop will execute the code block repeatedly, with `VAR` taking on
values from `START` up to (but not including) `STOP`, incrementing by `STEP`
each time. `VAR` is a new variable that only exists within the loop.

For example, this code prints the numbers `0`, `2`, `4`, and `6`:

```evy
for i := range 0 7 2
  print i
end
```

`START`, `STEP`, and `VAR` are optional.

- `START` defaults to 0.
- `STEP` defaults to 1.
- If variable `VAR` is left out, you can't access the loop counter.

The following code prints `"hello"` three times.

```evy
for range 3
    print "hello"
end
```
