# 💬 Let's Chat

⭐ **Before you run the code:** Can you predict what will happen?

Now, hit **Run** and see if you were right!

⭐ **Think about it:** What's the purpose of the `:=` operator?

## [>] `:=` Declaration with type inference 📖

In Evy, you can declare a variable and assign it a value in one step using `:=`.

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

⭐ **Challenge:** Can you add a different response if the answer is "no"? Use
`else if` to create an alternative message.

### [>] Code hint 🧚

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

⭐ **Your turn:** Can you create your own chat bot?

### [>] Some ideas

- 🍦 Ask about their favorite ice cream flavor instead of cookies.
- 🎁 Ask if they want to open a surprise.
  - If they say `"yes"`, reveal the surprise (let your imagination run wild 🐉).
  - If they say `"no"`, respond with something like "And so it remains my secret
    🔒".
  - If they say anything else, respond with "I don't understand."

You could also ask about their favorite color, band, or football team and
respond accordingly!
