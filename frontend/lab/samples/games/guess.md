# 🎲 Guess my Number

Create a number guessing game. The player guesses a number between 0 and 9 and
gets hints if their guess is too high or too low.

## ⭐ Intro

**Read** the code. What will happen when it's run?

**Run** the code. Was it what you expected?

[Next]

## ⭐ Check Guess

Delete the two `print` statements at the bottom of the `key` event handler.

Print the message `"You win!"` if the player guesses the number correctly,
otherwise print `"Try again."`.

### [>] Hint

```evy
if guess == ❓
  print ❓
else
  print ❓
end
```

[Next]

## ⭐ New Game

Start a new game after winning:

- Sleep for a second
- Print the message stored in `msg` again
- Generate a new random number

### [>] Hint

```evy
if guess == ...
  print ...
  sleep ❓
  print ❓
  number = ❓
else
```

[Next]

## ⭐ Add Hints

Print `"Too low."` if the guess is less than the number, `"Too high."` if the
guess is greater.

Start a new game after winning:

Add a `guess` variable to store the player's guess. Print the guess to the

### [>] Hint

```evy
else if guess < ❓
  print guess "is too ❓."
else
  print ❓
end
```
