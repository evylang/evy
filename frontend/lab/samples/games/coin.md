# 🪙 Coin Toss

## ⭐ Intro

**Read** the code. What will happen when it's run?

**Run** the code. Was it what you expected?

[Next]

## ⭐ Toss

Add the computer's coin toss.

At the beginning of the `key` event handler, add a
`coin` variable that randomly gets assigned `"h"` or `"t"`.

Just after `print "Your guess:" guess` add `print "My coin:   " coin`

### [>] Hint

```evy
on key guess:string
    coin := "h"
    if (rand1) < ❓
        coin = ❓
    end
    print "Your guess:" guess
    print "My coin:   " ❓
    sleep 1
```

There multiple variants that will get you to the same result, e.g. `rand 2`.

[Next]

## ⭐ Win or Lose?

Compare your guess to the computer's coin toss.

If they match, print `"You win!"` otherwise print `"You lose!"`.

### [>] Hint

```evy
if guess == ❓
	print "You win!"
else
	❓
end
sleep 1
```
