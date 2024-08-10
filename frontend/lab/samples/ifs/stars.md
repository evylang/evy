# ✨ Random Stars

⭐ Write a program that draws:

- A black background using `clear`
- A `"white"` circle with radius `1` anywhere on the canvas.

⭐ Use the [`rand`] function to generate random coordinates for your white
circle.

`rand n` generates a random whole numbers between 0 and `n`.

## [>] Code hint 🧚

```evy
x := rand 100
y := ❓
move ❓ ❓
circle ❓
```

[`rand`]: /docs/builtins.html#rand

---

⭐ Draw `200` circles with radius `1` in random positions.

- You'll need a loop for this.
- Make sure to generate new random coordinates **inside** the loop.

## [>] Code hint 🧚

```evy
i := 0
while i ❓
    x := rand ❓
    // ...
    i = i + 1
end
```

---

⭐ Make the circle sizes random between `0` and `0.6`.

- Use the [`rand1`] function to get a random number between 0 and 1.
- Multiply that random number by `0.6` to scale it to the desired range.

[`rand1`]: /docs/builtins.html#rand1

## [>] Code hint 🧚

```evy
size := (rand1) * 0.6
circle size
```

---

⭐ Change the color of 10% or your circles to `"gold"`

## [>] Code hint 🧚

```evy
c := rand1
if c < 0.1
    color "gold"
else
    ❓
end
```

---

⭐ Make it your own. Change the number of stars, sizes and colors to create your
favorite night sky.
