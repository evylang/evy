# ✨ Random Stars

## ⭐ Introduction

Write a program that draws:

- A black background using `clear`
- A `"white"` circle with radius `1` anywhere on the canvas.

[Next]

## ⭐ Random Position

Use the [`rand`] function to generate random coordinates for your white circle.

`rand n` generates a random whole numbers between `0` and `n`.

## [>] Hint

```evy
x := rand 100
y := ❓
move ❓ ❓
circle ❓
```

[`rand`]: /docs/builtins.html#rand

[Next]

## ⭐ 200 Circles

Draw `200` circles with radius `1` in random positions.

- You'll need a loop for this.
- Make sure to generate new random coordinates **inside** the loop.

## [>] Hint

```evy
i := 0
while i ❓
    x := rand ❓
    // ...
    i = i + 1
end
```

[Next]

## ⭐ 200 Stars

Make the circle sizes random between `0` and `0.6`.

- Use the [`rand1`] function to get a random number between 0 and 1.
  (`rand1` has _no_ space before the`1` !)
- Multiply that random number by `0.6` to scale it to the desired range.

[`rand1`]: /docs/builtins.html#rand1

## [>] Hint

```evy
size := (rand1) * 0.6
circle size
```

[Next]

## ⭐ Sprinkle a Bit of Gold

Change the color of 10% or your circles to `"gold"`

## [>] Hint

```evy
c := rand1
if c < 0.1
    color "gold"
else
    ❓
end
```

[Next]

## ⭐ Make It Your Own

Change the number of stars, sizes and colors to create your
favorite night sky - here's [mine].

[mine]: https://play.evy.dev/#content=H4sIAAAAAAAAA2WO0QqDMAxF3/sVlz5tDqQ+yGDox7g2m8HOSnVT/360HUyRQMJNTi5XW2o85N02upOCcauhxNyyJTAqlEoJAFjCwTe9KaJc91I7694ecm55IhlXL/chLFmhFNbQE7d/4wc0Kqg8ya3T01mTjMiOtCHLA2qIhoH77o8fkE0u6k2cY4hyilnOyKDya3LW7LUljNmvUlDUYFxQiPD+BRnfw+AzAQAA
