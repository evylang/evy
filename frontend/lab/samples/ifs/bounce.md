# 🏓 Bouncy Ball Bonus

⭐ **Your Challenge:** Can you create a bouncing ball animation?

Check out the [bouncing ball page 🏓📺](#bounce-show) or observe the
animation below.

## Demo

### [>] Animation demo

![small centered circle](img/bounce.gif)

## Animation notes 🗒

- Circle radius: `10`
- Outline width: `1`
- Outline color (stroke): `springgreen`
- Fill color: `hsl 0 0 0 3`
- Clear with transparent black: `hsl 0 0 0 4`
- Sleep one millisecond

⭐ Start by making a single green circle at position `0 50`.

⭐ Move the ball horizontally across the screen, like in the
[🟣🚚 Move challenge](#move) in the Introduction lab, don't worry about the bounce yet.

⭐ Finally, to change direction at the edges use the reversible increment trick
from the [Pulse challenge](#pulse).

### [>] Hint

Inside the loop body add:

```evy
x = x + inc
if x < 10 or x > 90
    inc = -inc
end
```
