# 🌈📺 Gradient

**Run** the program to draw an animated color gradient.

This program was created using the [`line`] and [`width`] commands.

[`line`]: /docs/builtins.html#line
[`width`]: /docs/builtins.html#width

How would you go about writing this program?

### [>] Hint

- Break down the problem into smaller bits.
- Create a static image with line `width 10`:
  ![thick vertical lines](img/gradient-thick.svg)
- Animate the lines with the `sleep` command.
- Reduce the line width, move, sleep and loop increment.

### [>] Docs

The following code draws a line from point `10 20` to point `80 50`:

```evy
width 2
move 10 20
line 80 50
```

Output:

![Line from 10 20 to 80 50](img/gradient-line.svg)
