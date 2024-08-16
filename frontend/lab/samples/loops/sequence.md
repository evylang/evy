# 🔢 Sequences

**⭐ Warm up:** Can you write a program that prints the numbers from 0 to 9?

## [>] `while` loop structure

```evy
while loop_condition
    loop_body
    // …
end
```

### [>] Code hint 🧚

```evy
x:num
while x <= ❓
    print x
    x = x + ❓
end
```

---

Write programs that generate the first 20 numbers of the following sequences:

- ⭐️ 0, 2, 4, 6, …
- ⭐️ 1, 4, 7, 10, …
- ⭐️ 20, 18.5, 17, …
- ⭐️ 1, 2, 4, 8, …
- ⭐️ 1, 10, 100, 1000, …
- ⭐️ 1, 0.5, 0.25, 0.125, …
- ⭐️ 1, 3, 6, 10, …

Use _two different_ variables to track the count used in the loop condition and the printed
sequence number.

## [>] Code hint 🧚

```evy
x:num
a:num
while x <= ❓
    print a
    a =  ❓
    x = x + ❓
end
```

## [>] Solution: 20th number

- 38
- 58
- -8.5

- 524288
- 10000000000000000000
- 0.0000019073486328125

- 210

## [>] Number Sequences 📖

The first 3 sequences are **arithmetic sequences** where you add the same amount to
get from one number to the next.

The next 3 sequences are **geometric sequences** where you multiply by the same
amount to get from one number to the next.

The last sequence is the **triangle sequence** where each number is the sum of the
previous number and its position in the sequence.
