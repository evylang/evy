# 🧿 Pulse

⭐ **Your Task:** Write a program that draws a small blue circle.

![small centered circle](samples/ifs/img/pulse-step-1.svg)

- Background (`clear`): `"black"`
- Circle outline width: `1`
- Circle fill color: `"none"`
- Circle outline color (`stroke`): `"blue"`
- Initial radius: `1`

### [>] Code hint 🧚

```evy
width ❓
clear ❓
fill ❓
stroke ❓
move 50 ❓
circle ❓
```

---

⭐ **Enhance your program:** Create this drawing of concentric circles:

![small centered circle](samples/ifs/img/pulse-step-2.svg)

- Use a variable `r` for the circle radius, starting at `1`.
- Within a loop:
  - Draw the circle with radius `r`.
  - Increase the radius `r` by `2` in each loop iteration.
- Keep looping as long as the radius `r` is less than `45`.

### [>] Code hint 🧚

```evy
r := ❓
while r < ❓
	circle ❓
	r = r +❓
end
```

---

⭐ Add a `sleep` of `0.1` seconds after drawing each circle to create an
animation.

### [>] Animation demo

![small centered circle](samples/ifs/img/pulse-step-3.gif)

### [>] Code hint 🧚

```evy
while r < __
	circle __
	r = r + __
	sleep ❓
end
```

---

⭐ Now, at the beginning of each loop iteration, add a nearly transparent black overlay.

Use `clear` with `hsl 0 0 0 15` to achieve the fading effect.

### [>] Animation demo

![small centered circle](samples/ifs/img/pulse-step-4.gif)

### [>] Code hint 🧚

```evy
while r < __
	clear (hsl ❓)
	circle __
	r = r + __
	sleep __
end
```

---

⭐ Make the animation smoother.

Reduce the loop increment, sleep duration, and alpha value.

### [>] Animation demo

![small centered circle](samples/ifs/img/pulse-step-5.gif)

### [>] Code hint 🧚

- increment: r = r + 0.1
- sleep: 0.001 seconds
- alpha: hsl 0 0 0 1

---

⭐ **Final step: the pulse**

Let's make the circle continuously grow and shrink.

**Loop Forever:** Change the loop condition to `true` to create an endless loop.

```evy
while true
    // ...
end
```

**Change Direction:** Instead of always increasing the radius (`r`) by `0.1`,
use a variable `inc` to control the change.

```evy
inc := 0.1  // Amount to increase/decrease the radius
while true
    r = r + inc
end
```

**Reverse the Change:** Inside the loop, check if `r` goes below 1 or above 45.
If it does, flip the sign of `inc` to reverse the animation's direction.

### [>] Code Hint 🧚

```evy
inc := 0.1
while true
    if r < 1 or r > 45
        inc = -inc  // Reverse the increment
    end
    r = r + inc
end
```
