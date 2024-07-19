# The `print` Command

The `print` command tells the program to print something to the screen.
The **argument** is the thing that you want the program to print.
The argument must be surrounded by quotation or talking marks `"`.
For example, the following code will print the word `Hello` to the screen:

```
✅ print "Hello"
```

`"Hello"` is the argument to `print`.

Without the quotation marks, this program will generate an error.

```
❌ print Hello
```

## More Arguments

`print` takes any number of arguments: 0, 1, 2, …. Each argument is
printed _without_ the quotation marks. Arguments are separated by a single
space. After the last argument has been printed a new line or return will be
printed.

The following program

```evy
print "Bugs like hugs."
print "👾" "🐛" "🥰"
```

creates the output

```
Bugs like hugs.
👾 🐛 🥰
```

`print` without any arguments prints an empty new line.

[Exercise](README.md)
