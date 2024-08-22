<h1>🟣🚚 Moving Dot</h1>

⭐ Can you draw 6 purple circles at x coordinates 0, 20, 40, 60, 80, and 100?

## [>] Result

![6 circles on horizontally center aligned](samples/intro/img/move-6circles.svg)

---

⭐ Can you draw these 6 purple circles using a loop?

## [>] Loop structure

```evy
while loop_condition
    loop_body
    // …
end
```

### [>] Code hint 🧚

```evy
x:num
while x <= ❓
    move x 50
    circle 10
    x = x + ❓
end
```

---

⭐ Can you change the program to make the circle move from left to right, using the
[`clear`] and [`sleep`] commands?

[`clear`]: /docs/builtins.html#clear
[`sleep`]: /docs/builtins.html#sleep

## [>] Result

![one horizontally moving circle](samples/intro/img/1-circle.gif)

### [>] Code hint 🧚

```evy
while // …
   clear
   // …
   sleep 0.2
end
```

---

⭐ Make 2 circles move in opposite direction

## [>] Result

![two horizontally moving circles](samples/intro/img/2-circles.gif)

### [>] Code hint 🧚

```evy
move x 40
circle 10
move 100-x 60
// …
```

---

⭐ Make 4 circles move in opposite direction

## [>] Result

![four moving circles](samples/intro/img/4-circles.gif)

### [>] Code hint 🧚

```evy
move 100-x 60
circle 10
move  40 x
// …
```
