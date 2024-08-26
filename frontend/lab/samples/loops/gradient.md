# 🌈🫵 Gradient

## ⭐ Intro

**Read** the code. What do you think will happen when you run it?

**Run** the code. Was it what you expected?

[Next]

## ⭐ Use a Variable

Can you replace the value `30` with a variable in 3 places?
We will iterate over this variable in a following step.

### [>] Hint

```evy
x:num
x = 30
c = hsl ❓+200
color c
move ❓ 100
// One more to replace
```

[Next]

## ⭐ Use a Loop

Can you use this variable to create a loop from 0 to 100 with a step of 10?

### [>] Hint

```evy
while x <= ❓
    c = hsl __
    color c
    move __
    line __
    x = x + ❓
end
```

[Next]

## ⭐ Animation

Can you animate the lines with the `sleep` command?

### [>] Hint

```evy
while x <= __
    // ...
    sleep ❓
    x = x + __
end
```

[Next]

## ⭐ Animation Smoothing

Can you reduce the line `width`, `sleep` and loop increment to create a smooth
gradient?
