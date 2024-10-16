# 🫧 Bubble

⭐ Create a program of a rising, `"skyblue"` bubble with radius `5`.

- move it from `50 0` to `50 110` using a `for` loop.
- increment the y-coordinate by `0.5`
- animate it with a sleep for `0.02` seconds.

## [>] Hint

```evy
for y := range 0 110 ❓
  clear
  move 50 y
  circle 5
  sleep ❓
end
```

[Next]

⭐ Add randomness to the x-coordinate of the bubble.

Change the x-coordinate by a maximum of ± 2 of its previous value.

![Animated bubble](img/bubble.gif)

## [>] Hint

```evy
//...
r := rand 5 // 0..4
x = x + r - 2
move x y
//...
```
