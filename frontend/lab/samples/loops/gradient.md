# 🌈🫵 Gradient

**Run** the program on the right and see if you understand its code.

⭐ Can you replace the value `30` with a variable in 3 places?

### [>] Code hint 🧚

```evy
x:num
x = 30
c = hsl ❓+200
color c
move ❓ 100
// One more to replace
```

---

⭐ Can you use this variable to create a loop from 0 to 100 with a step of 10?

### [>] Code hint 🧚

```evy
while x <= ❓
    c = hsl __
    color c
    move __
    line __
    x = x + ❓
end
```

---

⭐ Can you animate the lines with the `sleep` command?

### [>] Code hint 🧚

```evy
while x <= __
    // ...
    sleep ❓
    x = x + __
end
```

---

⭐ Can you reduce the line `width`, `sleep` and loop increment to create a
smooth gradient?
