# Monkey

This is an implementation of Monkey programming language from Thorsten Ball's [Writing an Interpreter in Go](https://interpreterbook.com/).
It closely follows the implementation and specification but is not exactly the same as in the book, because where would the fun be in that?
Among the features that deviate from the book's implementation are:

- Booleans support addition and multiplication (`+` means OR, `*` means AND)
- Arrays support concatenation using `+` operator
- Arrays support slicing using either `slice` builtin or using `a[start:end]` syntax (half-open range *[start, end)*)
