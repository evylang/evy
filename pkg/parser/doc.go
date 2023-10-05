// Package parser generates an AST for Evy programs.
//
// First, Evy source code is tokenized by the [lexer], then the generated
// tokens are parsed into an [Abstract Syntax Tree] (AST) by the
// [Parse] function. Finally, the [evaluator] walks the AST and executes
// the program. This package is concerned with the second step of parsing
// tokens into an AST.
//
// # Parsing
//
// The [Parse] function in this package uses a top-down, recursive descent
// parsing approach. This means that parsing starts with the top-level
// structure of the program and recursively descends into its components.
//
// Expression parsing is a particularly tricky part of the parsing process.
// Expressions are used in variable declarations, assignments, function
// calls, and many other places. Expressions are not simply evaluated
// top-down, left-to-right. For example, in the expression 2+3*4, the
// multiplication operator * has higher precedence than the addition
// operator +. This means that the expression 3*4 must be evaluated before
// the expression 2+3.
//
// To handle expressions with different levels of precedence, the
// [Parse] function uses a [Pratt parser]. A Pratt parser is a type of
// top-down, recursive descent parser that can handle expressions with
// different levels of precedence.
//
// # Abstract Syntax Tree
//
// The Abstract Syntax Tree (AST) is a hierarchical representation of an
// Evy program's structure. It is the result of the [Parse] function. It
// can be used for evaluation, further analysis, and code generation, such
// as formatted printing of the Evy source code with [Program.Format]. The
// AST consists of a tree of nodes, each of which implements the
// [Node] interface.
//
// Each node in the AST represents a different element of the program's
// structure. The root node of the AST represents the entire [Program]. The
// root node's direct child nodes are block statements and basic
// statements.
//
// Block statements contain further statements. They include:
//   - [FuncDefStmt]: A statement that defines a function.
//   - [EventHandlerStmt]: A statement that defines an event handler.
//   - [IfStmt]: A statement that executes a block of code if a condition is met.
//   - [ForStmt]: A statement that executes a block of code repeatedly.
//   - [WhileStmt]: A statement that executes a block of code repeatedly while a condition is met.
//
// Basic statements are statements that cannot be broken down into further
// statements. They include:
//   - [TypedDeclStmt]: A statement that declares a variable of an explicitly specified type.
//   - [InferredDeclStmt]: A statement that declares a variable with a type that is inferred from the value.
//   - [AssignmentStmt]: A statement that assigns a value to a variable.
//   - [FuncCallStmt]: A statement that calls a function.
//   - [ReturnStmt]: A statement that returns from a function.
//   - [BreakStmt]: A statement that breaks out of a loop.
//
// The components of basic statements are:
//   - Variables: [Var]
//   - Literals: [NumLiteral], [StringLiteral], [BoolLiteral], [ArrayLiteral], [MapLiteral]
//   - Expressions: [UnaryExpression], [BinaryExpression], [IndexExpression], [SliceExpression], [DotExpression], [GroupExpression], [TypeAssertion], and [FuncCall]
//
// Variables are named references to values. Literals are values that
// are directly represented in the Evy source code. Expressions are
// combinations of variables, literals, operators and function calls to
// form new values.
//
// This structure closely resembles the [grammar] of the Evy programming
// language, as defined in the [language specification].
//
// [evaluator]: https://pkg.go.dev/evylang.dev/evy/pkg/evaluator
// [Abstract Syntax Tree]: https://en.wikipedia.org/wiki/Abstract_syntax_tree
// [grammar]: https://github.com/evylang/evy/blob/main/docs/spec.md#evy-syntax-grammar
// [language specification]: https://github.com/evylang/evy/blob/main/docs/spec.md
// [Pratt parser]: https://en.wikipedia.org/wiki/Operator-precedence_parser#Pratt_parsing
package parser
