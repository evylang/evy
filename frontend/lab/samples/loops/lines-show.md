# 🌐📺 Lines

Run the program to draw and check out the cool animation.

For this program we have two **nested loops** 🤯.

## [>] Nested loops 📖

Explore the following code:

```evy
a:num
b:num
while a <= 1
    while b <= 10
        print a b
        b = b + 10
    end
    b = 0 // reset inner loop counter
    a = a + 1
end
```

Its output is:

```
0 0
0 10
1 0
1 10
```

## Implementation

How would you go about writing this program?

## [>] 💡 Some tips

- Break down your problem into smaller bits.
- Create a static with image with step `10` first.
  ![target board](samples/loops/img/lines.svg)
- Make the `step` a variable.
- Wrap the code in a second `while` loop with a `sleep`.
