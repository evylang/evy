# 💬 Let's Chat

## ⭐ Intro

**Read** the code. What do you think will happen when you run it?

**Run** the code. Was it what you expected?

---

## ⭐ Type Inference

Note the new `:=` operator. Have a guess how it works.

### [>] Docs

In Evy, you can declare a variable and assign it a value in one step using `:=`.
This is called **Declaration with Type Inference**.

Instead of

```evy
s:string
s = "banana"
```

you can use the shortcut

```evy
s := "banana"
```

---

## ⭐ Add `else if`

Can you add a different response to the program if the answer is `"no"`? Use `else
if` to create an alternative message.

### [>] Hint

```evy
if answer == "yes"
    print "🍪 Sweet!"
else if answer == ❓
    ❓❓
else
    print "I'm confused"
end
```

---

## ⭐ Your Turn

Can you create your own chat bot?

### Some ideas

🍦 Ask about their favorite ice cream flavor instead of cookies.

🎁 Ask if they want to open a surprise.

- If they say `"yes"`, reveal the surprise (let your imagination run wild 🐉).
- If they say `"no"`, respond with something like `"And so it remains my secret
🔒"`.
- If they say anything else, respond with `"I don't understand."`

You could also ask about their favorite color, band, or football team and
respond accordingly!
