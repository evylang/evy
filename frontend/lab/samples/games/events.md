# 🎪 Events

In Evy, **event handlers** react to events like keyboard input (`on key`).

### [>] Docs

An event handler starts with `on`, then the event type, parameters, code, and
ends with `end`. This example prints the pressed key:

```evy
on key k:string
  print k
end
```

`k:string` is the parameter (must be a string).

There are six types of events in Evy: `key`, `down`, `up`, `move`, `animate`,
and `input`.

## ⭐ Intro

**Read** the code. What will happen when it's run?

**Run** the code. Was it what you expected?

---

## ⭐ Update `down`

Update the `down` event handler to draw a circle with radius `1` at `x y`.

Can you see circles on click or tap?

---

## ⭐ Update `key`

Update the `key` event handler:

- If the key is `"r"`, clear with `"red"`.
- If the key is `"g"`, clear with `"green"`.
- If the key is `"b"`, clear with `"blue"`.
- Otherwise, just use `clear`.

Can you change the background color?
