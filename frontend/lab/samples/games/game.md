# 👾 Game

## ⭐ Intro

**Read** the code. Pretty similar to the last challenge, isn't it?

**Run** the code. Was it what you expected?

Note how we are using the `text` command to display text on the drawing area.

---

## ⭐ Move with Keys

Use the `key` event handler to move our friend left and right 👾:

- Move left with `←` or `h`
- Move right with `→` or `l`

Can you work out a way that the movement wraps around the screen? Use the modulo
operator `%`.

### [>] Hint

```evy
on key k:string
    if k == "ArrowLeft" or k == "h"
        x = (x + 99) % 100
    else if k == ❓
        x = ❓
    end
end
```

---

## ⭐ Add Gold Arrow

Add a `"gold"` colored arrow with `text "▶▶"` that moves left to right on the
screen at y-coordinate `30`. Store its x-coordinate in a global variable `gold`
and initialize with `0`.

Update the `gold` x-position inside the `animate` event handler with:

```evy
gold = (gold + 0.3) % 110
```

### [>] Hint

```evy
gold := 0

on animate
    // Update
    y = (y + 0.1) % 100
    gold = (gold + 0.3) % 110

    // Draw
    clear
    move x y
    text "👾" // size around 7x5
    color "gold" // arrows
    move gold 30
    text "▶▶"
```

---

## ⭐ Add Orange and Red Arrow

|          | Orange Arrow | Red Arrow     |
| -------- | ------------ | ------------- |
| Text     | `"◀◀"`     | `"▶▶"`      |
| Variable | `orange`     | `red`         |
| Initial  | `50`         | `0`           |
| Color    | `"orange"`   | `"orangered"` |

Use the update functions:

```evy
orange = 100 - (100 - orange + 0.5) % 120
red = (red + 0.7) % 130
```

Place along-side `gold` arrow from previous step.

Run code and ensure you see three flying arrows as well as our moving friend.

### [>] Hint

Check out a [partial solution] on the Evy Playground.

[partial solution]: https://play.evy.dev/#content=H4sIAAAAAAAAA21RzWrCQBC+5yk+Fgop0rpWgijsodBjT4U+wNKMMRizZZPWpCdfoRc99i089mF8gj5CmdmIUQwsMzvfz85kGswMEh21HHUUZa5IQ+q8LTPqYE9dNXIlbJmvbE0RAAyHeH1Pj7cWBnGLAfT96BY3GGktdXE1iCUyOhZ0FNDuJcN03CEOoasyOxH2Q2BzKwYxB8Ymgo11dGznydu15G8FWS/Zyn0SGrRyqampof5+vn8V06v8i2C9+yhTTJokKF3hPBR3KxzrvVtXJysZY6x7doft/rDdq7469K9Oqm6g5Ey32xx2mys6T2lPysNOrr1HZSorWVKL5ayqfV5mQsvnWMIYqEfu/ZnmtYLzXW0RrPlr+F82GGA67S+MioouPF7ybNE3Ka6YnC2de+PzDzXeqUNjAgAA

---

## ⭐ Add Collision detection

Add a collision detection the game to the end of `animate` event handler.

When friend's x coordinate and gold arrow's x coordinate are less than `6` apart, and their y coordinates are less than `4.5`
print a game over message and exit the program:

```evy
print "🟡 Game over."
exit 0
```

Do the same for `orange` and `red` arrows.

### [>] Hint

```evy
// Check collision
if (abs x-gold) < 6 and (abs y-30) < 4.5
  print "🟡 Game over."
  exit 0
else if (abs x-orange) < 6 and (abs y-50) < ❓
  print "🟠 Game over."
  exit 0
else if (abs x-red) ❓
  ❓
end
```

---

## ⭐ Add Level messages

Add a level message that increments every time our friend successfully gets to the top.

Use a global variable `level` and initialize with `0`.

At the beginning of the `animate` event handler check if `y` is less than `0.1` and increment and print the `level`:

```evy
level = level + 1
print "Level" level
```

### [>] Hint

Check out a [full solution] on the Evy Playground.

[full solution]: https://play.evy.dev/#content=H4sIAAAAAAAAA5VTS27bMBDd6xQPAgooMOzIdRU3hlWgaIFusigKdF0w5kQmLJMBJdtUV7lCgSJZOr1DkWUP4xPkCMVQUqw07s8LDzlv5r03FJnTmnJMUsRB4DgmcVA1+8zksl4aK3RGDWypyQZGQ2i1FCUFAHB8jPdW6RKaNsiZ2KfVBSpMEQ+Gfss/DyJtYg975NIThGcMhB0S0jJoRT5eylayQoqoQo/Zj/AMwzj2eW89ReQjoyOPDmu0GSflcvQR1aHJcnXiq5/X1TxviogDY2OPjeIHO2+t2Pj1LCdh/Wpp1gSHym9KciXC++2XHyGXF+ozQViz0hJjl9SdJjcWIbv1NcJasyn2VH6MUdyh213f7a7vwm537T/cdzUDJY/6bq52N1cH+izJTisPOz6k1w79Zk6zBTPkqlBGtx86EucFXJ/9HmGKEwgt62TVH8WcejFIfv3Y99vtN7wTS4JZkx2EDzg5VaJ2QXlBHYHa8xOJ5A8St/8pYenpCOPf83/9/ld+LQN/jY3GgiosJkVplc7aI1W6tEauZgTNR56RRmmwVlxctAe8QJoifM3344wuyhDGNrn5XtTxfXXo4fS0+yjaATscH1Q275LkB0gePazuENJsNJyc6NUSn/i/Nekkpmhezz/4cRKv8PJA+UHlnx2spi2yBAAA
