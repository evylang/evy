# 🌐🫵 Lines

⭐ Can you complete the program on the right to create this output?

![target board](samples/loops/img/lines.svg)

⭐ Can you do it using a loop?

## [>] Code hint 🧚

```evy
while x < ❓
    move ❓
    line ❓
    x = x + ❓
end
```

---

⭐ Can you make the step `10` a variable and run the program with different step
values?

## [>] Code hint 🧚

```evy
step = ❓
while x < __
    // ...
    x = x + ❓
end
```

## [>] ⭐⭐⭐ Animation!

Add a second, wrapping, outer loop that ranges over the `step` variable.

- Set `step` to `10` at the beginning
- Reduce the `step` variable by `-0.05`.
- Keep looping as long as `step` is greater then `2`
- Use the `clear` command.
- Use the `sleep 0.01` command.
- Reset the position variable `x` to `0`.
