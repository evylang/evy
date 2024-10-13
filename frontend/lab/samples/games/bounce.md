# 🏓 Bouncy Ball Bonus

## ⭐ Intro

**Read** the code. Pretty similar to the last challenge, isn't it?

**Run** the code. Was it what you expected?

[Next]

## ⭐ Bounce

Make the ball bounce back and forth on the drawing area.

Remember the [Pulse Challenge](../ifs/pulse.md)?

The bounce motion uses a similar trick to change direction.

### [>] Hint

Inside `animate` add:

```evy
x = x + inc
if x < 10 or x > ❓
    inc = -❓
end
```

[Next]

## ⭐ Move with Keys

Use the `key` event handler to move the ball up and down:

- Move up with `⬆` or `k`
- Move down with `⬇` or `j`

Declare a global variable `y` that's used inside `animate` to set the balls
y-coordinate. Update it inside the `key` event handler.

### [>] Hint

```evy
y := 50

on animate
  // ...
  move x ❓
  // ...
end

on key k:string
  if k == "ArrowUp" or k == "k"
    y = y + 1
  else if ❓
    y = ❓
  end
end
```
