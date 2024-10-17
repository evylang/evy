# 🎨 Paint

## ⭐ Intro

**Read** the code. What will happen when it's run?

**Run** the code. Was it what you expected?

[Next]

## ⭐ Circle Color

Change the color of the circle to `"orange"` in the `down` event handler.

[Next]

## ⭐ Add a `up` Event Handler

Add an `up` event handler to draw a `"red"` circle with radius `1` at `x y`.

Run and test the program. Do you see orange and red circles?

[Next]

## ⭐ Add a `move` Event Handler

Add an `move` event handler to draw a `"black"` line to `x y`.

Run and test the program. Do you see black lines connecting orange and red
circles?

[Next]

## ⭐ Only Paint When Down

If you're using a touchscreen device, you can skip this step.

If you're _not_ on a touchscreen, you might have noticed that lines are drawn
even when you're not clicking. To fix this, we'll track whether the mouse
button is pressed.

Add a global variable `d` with the value `false`. In the `down` handler, set `d`
to `true`. In the `up` handler, set it to `false`. Only draw a line if `d` is
`true`.

### [>] Hint

Check out the [Playground Drawing Sample].

[Playground Drawing Sample]: https://play.evy.dev/#draw

[Next]

## ⭐ Change Drawing Pen Color

Add a black-and-green palette at the bottom. Clicking it changes the pen color.

![Screencast of drawing program](img/drawing.gif)

Add a `"black"` rectangle with dimensions `50 20` at `0 0`.

Add a `"green"` rectangle with dimensions `50 20` at `50 0`.

Add a global variable `pen` with the value `"black"`.

In the `down` handler, if `y` is less than `20` update the `pen` variable: If
`x` is less than `50` set `pen` to `"black"`, otherwise set it to
`"green"`.

In the `move` handler set the color to `pen`.

### [>] Hint

Check out this working [solution] on the Evy Playground.

[solution]: https://play.evy.dev/#content=H4sIAAAAAAAAA22PUQ6CMBBE//cUk54ATPgx6l2QLtpYtqSAwO1Na6EJ2qRNd3Yy+9ZzM6IqcCqoc28O34IaZ52HenhmUeSzhWajxydKop4F5yvU3dbNS5EORVvbgYmcQLtZsJxl6rCGlwBA44rRTxwL02LFLWQinTh+wboLicL5Wh6ssmx8YxllFNgOHLIWXFDlrECX4TbjsZ32i23RFG6An/q/6Gm7H9SE6Vl/o3a+Le/rPiSaFvqwac+yK9ZInrDBfQBGJrOLrAEAAA==
