// Package bytecode contains a custom bytecode compiler and
// virtual machine.
//
// Initially, an Evy program is turned into a sequence of tokens by the
// [lexer]. Then, the [parser] generates an Abstract Syntax Tree (AST)
// from the tokens. The [Compiler] of this package walks the AST and
// writes the instructions to custom bytecode. For more on the AST
// refer to the [parser] package documentation.
//
// The [VM] can read the bytecode and execute the encoded instructions,
// providing a runtime similar to the [evaluator]. The virtual machine
// is a straight-forward stack implementation that does not use any
// additional registers.
package bytecode
