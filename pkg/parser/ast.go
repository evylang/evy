package parser

import (
	"fmt"
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

// Node represents a node in the AST.
type Node interface {
	// Token returns the token of the Evy source program associated with the node.
	Token() *lexer.Token
	// String returns a string representation of the node.
	String() string
	// Type returns the Evy type of the node, such as num, []string, NONE.
	Type() *Type
}

// Program is the top-level or root AST node. It represents the entire
// Evy program.
//
// Program implements the [Node] interface.
type Program struct {
	token *lexer.Token
	// Statements is the ordered list of top level block and basic statements of the given Evy program.
	Statements []Node
	// EventHandlers maps event names to their event handler statements.
	// It is used in web interface to set connect up relevant handlers with JS event handlers.
	EventHandlers map[string]*EventHandlerStmt
	// CalledBuiltinFuncs is a list of builtin functions that are called
	// in the program. It is used in web interface to hide or show
	// Canvas and input widgets, such as sliders or readline text field.
	CalledBuiltinFuncs []string

	alwaysTerms bool
	formatting  *formatting
}

// EmptyStmt is an AST node that represents an empty statement. An empty
// statement is a statement that does nothing. Empty statement is used
// for formatting, such as to add a blank line between statements.
//
// EmptyStmt implements the [Node] interface.
type EmptyStmt struct {
	token *lexer.Token // The NL token
}

// FuncCallStmt is an AST node that represents a standalone function
// call statement. It is a statement that calls a function without any
// surrounding expressions.
//
// FuncCallStmt implements the [Node] interface.
type FuncCallStmt struct {
	token    *lexer.Token // The IDENT of the function
	FuncCall *FuncCall
}

// FuncCall is an AST node that represents a function call. It can be
// used either as a standalone statement or as part of an expression.
//
// FuncCall implements the [Node] interface.
type FuncCall struct {
	token     *lexer.Token // The IDENT of the function
	Name      string
	Arguments []Node
	FuncDef   *FuncDefStmt
}

// UnaryExpression is an AST node that represents a unary expression,
// such as: -n.
//
// UnaryExpression implements the [Node] interface.
type UnaryExpression struct {
	token *lexer.Token // The unary operation token, e.g. !
	Op    Operator
	Right Node
}

// BinaryExpression is an AST node that represents a binary expression.
// A binary expression is an expression that has two operands and an
// operator, such as a + b.
//
// BinaryExpression implements the [Node] interface.
type BinaryExpression struct {
	T     *Type
	token *lexer.Token // The binary operation token, e.g. +
	Op    Operator
	Left  Node
	Right Node
}

// IndexExpression is an AST node that represents an indexing
// expression. It accesses the value of an element in an array, map or
// string. For example: array[i].
//
// IndexExpression implements the [Node] interface.
type IndexExpression struct {
	T     *Type
	token *lexer.Token // The [ token
	Left  Node
	Index Node
}

// SliceExpression is an AST node
// that represents a slice expression. A slice expression is used
// to access a subsequence of an array or string, such as: array[1:4].
//
// SliceExpression implements the [Node] interface.
type SliceExpression struct {
	T     *Type
	token *lexer.Token // The [ token
	Left  Node
	Start Node
	End   Node
}

// DotExpression is an AST node that represents a field access
// expression. A field access expression is an expression that accesses
// the value of a field in a map, such as person.age.
//
// DotExpression implements the [Node] interface.
type DotExpression struct {
	T     *Type
	token *lexer.Token // The . token
	Left  Node
	Key   string // m := { age: 42}; m.age => key: "age"
}

// TypeAssertion is an AST node that represents a type assertion
// expression. A type assertion expression is used to enforce the
// specific type of an any value. For example:
//
//	val:any
//	val = 1
//	print val.(num)+2 // 3
//
// TypeAssertion implements the [Node] interface.
type TypeAssertion struct {
	T     *Type
	token *lexer.Token
	Left  Node
}

// GroupExpression is an AST node that represents a parenthesized
// expression. It groups together an expression so that it can be
// evaluated as a single unit, such as:(a+b)*3.
//
// GroupExpression implements the [Node] interface.
type GroupExpression struct {
	token *lexer.Token
	Expr  Node
}

// Decl is an AST node that represents a variable declaration. A
// variable declaration is a statement that creates a new variable and
// assigns it a value. Variable declarations are used in
// [TypedDeclStmt] and [InferredDeclStmt] statements.
//
// Decl implements the [Node] interface.
type Decl struct {
	token *lexer.Token
	Var   *Var
	Value Node // literal, expression, assignable, ...
}

// TypedDeclStmt is an AST node that represents a typed declaration
// statement. A typed declaration statement declares a variable of an
// explicitly specified type, such as n:num.
//
// TypedDeclStmt implements the [Node] interface.
type TypedDeclStmt struct {
	token *lexer.Token
	Decl  *Decl
}

// InferredDeclStmt is an AST node that represents an inferred
// declaration statement. It declares a variable with a type that is
// inferred from the value that is assigned to it. For example: n :=
// 1.
//
// InferredDeclStmt implements the [Node] interface.
type InferredDeclStmt struct {
	token *lexer.Token
	Decl  *Decl
}

// AssignmentStmt is an AST node that represents an assignment
// statement. An assignment statement assigns a value to a variable,
// such as n = 2.
//
// AssignmentStmt implements the [Node] interface.
type AssignmentStmt struct {
	token  *lexer.Token
	Target Node // Variable, index or field expression
	Value  Node // literal, expression, assignable, ...
}

// ReturnStmt is an AST node that represents a return statement. A
// return statement terminates the execution of a function and can
// return a value. For example:
//
//	func square:num n:num
//	    return n * n
//	end
//
// ReturnStmt implements the [Node] interface.
type ReturnStmt struct {
	token *lexer.Token
	Value Node // literal, expression, assignable, ...
	T     *Type
}

// BreakStmt is an AST node that represents a break statement. A break
// statement is used to terminate the current loop statement, for
// example:
//
//	while true
//	    break
//	end
//
// BreakStmt implements the [Node] interface.
type BreakStmt struct {
	token *lexer.Token
}

// Token returns the token of the Evy source program associated with the
// FuncDefStmt node.
func (f *FuncDefStmt) Token() *lexer.Token {
	return f.token
}

// FuncDefStmt is an AST node that represents a function definition. It
// defines a new function with a name, a parameter list, return type,
// and a body. For example:
//
//	func greet
//	    print "howdy!"
//	end
//
// FuncDefStmt implements the [Node] interface.
type FuncDefStmt struct {
	token             *lexer.Token // The "func" token
	Name              string
	Params            []*Var
	VariadicParam     *Var
	ReturnType        *Type
	VariadicParamType *Type
	Body              *BlockStatement

	isCalled bool
}

// Token returns the token of the Evy source program associated with the
// IfStmt node.
func (i *IfStmt) Token() *lexer.Token {
	return i.token
}

// IfStmt is an AST node that represents a conditional statement. It
// specifies a condition that must be met for a block of statements to
// be executed. It can optionally have else-if and else blocks. For
// example:
//
//	if 2 * 5 == 10
//	    print "✔"
//	end
//
// IfStmt implements the [Node] interface.
type IfStmt struct {
	token        *lexer.Token
	IfBlock      *ConditionalBlock
	ElseIfBlocks []*ConditionalBlock
	Else         *BlockStatement
}

// Token returns the token of the Evy source program associated with the
// WhileStmt node.
func (w *WhileStmt) Token() *lexer.Token {
	return w.token
}

// WhileStmt is an AST node that represents a while statement, such as
//
//	while true
//	    print "🌞"
//	end
//
// WhileStmt implements the [Node] interface.
type WhileStmt struct {
	ConditionalBlock
}

// Token returns the token of the Evy source program associated with the
// ForStmt node.
func (f *ForStmt) Token() *lexer.Token {
	return f.token
}

// ForStmt is an AST node that represents a for loop. A for loop is a
// statement that repeats a block of code a certain number of times.
// The following code snippet is an example of a for loop:
//
//	for n := range 1 10 2
//	    print n // 1 3 5 7 9
//	end
//
// ForStmt implements the [Node] interface.
type ForStmt struct {
	token *lexer.Token

	LoopVar *Var
	Range   Node // StepRange or array/map/string expression

	Block *BlockStatement
}

// Token returns the token of the Evy source program associated with the
// StepRange node.
func (s *StepRange) Token() *lexer.Token {
	return s.token
}

// StepRange is an AST node that represents a step range in a for loop.
// A step range is used to iterate over a sequence of numbers, starting
// from the first number and ending with the last number, incrementing
// by the step size. For example:
//
//	for n := range 1 10 2
//	    print n // 1 3 5 7 9
//	end
//
// StepRange implements the [Node] interface.
type StepRange struct {
	token *lexer.Token

	Start Node // num expression or nil
	Stop  Node // num expression
	Step  Node // num expression or nil
}

// Token returns the token of the Evy source program associated with the
// ConditionalBlock node.
func (c *ConditionalBlock) Token() *lexer.Token {
	return c.token
}

// ConditionalBlock is an AST node that represents a conditional block.
// A conditional block is a block of statements that is executed only
// if a certain condition is met. Conditional blocks are used in
// [IfStmt] and [WhileStmt] statements.
//
// ConditionalBlock implements the [Node] interface.
type ConditionalBlock struct {
	token     *lexer.Token
	Condition Node // must be of type bool
	Block     *BlockStatement
}

// Token returns the token of the Evy source program associated with the
// EventHandlerStmt node.
func (e *EventHandlerStmt) Token() *lexer.Token {
	return e.token
}

// EventHandlerStmt is an AST node that represents an event handler
// definition. It includes the handler body, such as:
//
//	on key k:string
//	    print "key pressed:" k
//	end
//
// EventHandlerStmt implements the [Node] interface.
type EventHandlerStmt struct {
	token  *lexer.Token // The "on" token
	Name   string
	Params []*Var
	Body   *BlockStatement
}

// Token returns the token of the Evy source program associated with the
// Var node.
func (v *Var) Token() *lexer.Token {
	return v.token
}

// Var is an AST node that represents a variable, its name and type but
// not its value.
//
// Var implements the [Node] interface.
type Var struct {
	token  *lexer.Token
	Name   string
	T      *Type
	isUsed bool
}

// Token returns the token of the Evy source program associated with the
// BlockStatement node.
func (b *BlockStatement) Token() *lexer.Token {
	return b.token
}

// BlockStatement is an AST node that represents a block of statements.
// A block of statements is a sequence of statements that are executed
// together, such as those used in [FuncDefStmt] and [IfStmt].
//
// BlockStatement implements the [Node] interface.
type BlockStatement struct {
	token       *lexer.Token // the NL before the first statement
	Statements  []Node
	alwaysTerms bool
}

// Token returns the token of the Evy source program associated with the
// BoolLiteral node.
func (b *BoolLiteral) Token() *lexer.Token {
	return b.token
}

// BoolLiteral is an AST node that represents a boolean literal. A
// boolean literal is a value that can be either true or false.
//
// BoolLiteral implements the [Node] interface.
type BoolLiteral struct {
	token *lexer.Token
	Value bool
}

// Token returns the token of the Evy source program associated with the
// NumLiteral node.
func (n *NumLiteral) Token() *lexer.Token {
	return n.token
}

// NumLiteral is an AST node that represents a numeric literal. A
// numeric literal is a number, such as 12 or 34.567.
//
// NumLiteral implements the [Node] interface.
type NumLiteral struct {
	token *lexer.Token
	Value float64
}

// Token returns the token of the Evy source program associated with the
// StringLiteral node.
func (s *StringLiteral) Token() *lexer.Token {
	return s.token
}

// StringLiteral is an AST node that represents a string literal. A
// string literal is a sequence of characters enclosed in double
// quotes, such as "abc".
//
// StringLiteral implements the [Node] interface.
type StringLiteral struct {
	token *lexer.Token
	Value string
}

// Token returns the token of the Evy source program associated with the
// ArrayLiteral node.
func (a *ArrayLiteral) Token() *lexer.Token {
	return a.token
}

// ArrayLiteral is an AST node that represents an array literal, such
// as: [1 2 3].
//
// ArrayLiteral implements the [Node] interface.
type ArrayLiteral struct {
	token    *lexer.Token
	Elements []Node
	T        *Type
}

// Token returns the token of the Evy source program associated with the
// MapLiteral node.
func (m *MapLiteral) Token() *lexer.Token {
	return m.token
}

// MapLiteral is an AST node that represents a map literal. A map
// literal is a collection of key-value pairs, such as {a: 1, b: 2}.
//
// MapLiteral implements the [Node] interface.
type MapLiteral struct {
	token *lexer.Token
	Pairs map[string]Node
	Order []string // Track insertion order of keys for deterministic output.
	T     *Type
}

// Token returns the token of the Evy source program associated with the
// Program node.
func (p *Program) Token() *lexer.Token {
	return p.token
}

// String returns a string representation of the Program node.
func (p *Program) String() string {
	return newlineList(p.Statements)
}

// Format returns a string of the formatted program with consistent
// indentation and vertical whitespace.
func (p *Program) Format() string {
	var sb strings.Builder
	p.formatting.w = &sb
	p.formatting.format(p)
	return sb.String()
}

// Type returns [NONE_TYPE] for Program because a program does not have
// a type.
func (*Program) Type() *Type {
	return NONE_TYPE
}

func (p *Program) alwaysTerminates() bool {
	return p.alwaysTerms
}

// Token returns the token of the Evy source program associated with the
// EmptyStmt node.
func (e *EmptyStmt) Token() *lexer.Token {
	return e.token
}

// String returns a string representation of the EmptyStmt node.
func (e *EmptyStmt) String() string {
	return ""
}

// Type returns [NONE_TYPE] for EmptyStmt because the empty statement
// does not have a type.
func (*EmptyStmt) Type() *Type { return NONE_TYPE }

// Token returns the token of the Evy source program associated with the
// FuncCall node.
func (f *FuncCall) Token() *lexer.Token {
	return f.token
}

// String returns a string representation of the FuncCall node.
func (f *FuncCall) String() string {
	s := make([]string, len(f.Arguments))
	for i, arg := range f.Arguments {
		s[i] = arg.String()
	}
	args := strings.Join(s, ", ")
	return f.Name + "(" + args + ")"
}

// Type returns the return type of the called function.
func (f *FuncCall) Type() *Type {
	return f.FuncDef.ReturnType
}

// Token returns the token of the Evy source program associated with the
// FuncCallStmt node.
func (f *FuncCallStmt) Token() *lexer.Token {
	return f.token
}

// String returns a string representation of the FuncCallStmt node.
func (f *FuncCallStmt) String() string {
	return f.FuncCall.String()
}

// Type returns the return type of the called function.
func (f *FuncCallStmt) Type() *Type {
	return f.FuncCall.FuncDef.ReturnType
}

// Token returns the token of the Evy source program associated with the
// UnaryExpression node.
func (u *UnaryExpression) Token() *lexer.Token {
	return u.token
}

// String returns a string representation of the UnaryExpression node.
func (u *UnaryExpression) String() string {
	return "(" + u.Op.String() + u.Right.String() + ")"
}

// Type returns the type of the UnaryExpression, such as bool or num.
func (u *UnaryExpression) Type() *Type {
	return u.Right.Type()
}

// Token returns the token of the Evy source program associated with the
// BinaryExpression node.
func (b *BinaryExpression) Token() *lexer.Token {
	return b.token
}

// String returns a string representation of the BinaryExpression node.
func (b *BinaryExpression) String() string {
	if b.Op == OP_AND || b.Op == OP_OR {
		return "(" + b.Left.String() + " " + b.Op.String() + " " + b.Right.String() + ")"
	}
	return "(" + b.Left.String() + b.Op.String() + b.Right.String() + ")"
}

// Type returns the type of the BinaryExpression, such as bool, num or string.
func (b *BinaryExpression) Type() *Type {
	return b.T
}

// Token returns the token of the Evy source program associated with the
// IndexExpression node.
func (i *IndexExpression) Token() *lexer.Token {
	return i.token
}

// String returns a string representation of the IndexExpression node.
func (i *IndexExpression) String() string {
	return "(" + i.Left.String() + "[" + i.Index.String() + "])"
}

// Type returns the type of the IndexExpression, for example num for an
// array of numbers with type []num.
func (i *IndexExpression) Type() *Type {
	return i.T
}

// Token returns the token of the Evy source program associated with the
// SliceExpression node.
func (s *SliceExpression) Token() *lexer.Token {
	return s.token
}

// String returns a string representation of the SliceExpression node.
func (s *SliceExpression) String() string {
	start := ""
	if s.Start != nil {
		start = s.Start.String()
	}
	end := ""
	if s.End != nil {
		end = s.End.String()
	}
	return "(" + s.Left.String() + "[" + start + ":" + end + "])"
}

// Type returns the type of the SliceExpression, which is the same type
// as the array that is sliced or string if a string is sliced.
func (s *SliceExpression) Type() *Type {
	return s.T
}

// Token returns the token of the Evy source program associated with the
// DotExpression node.
func (d *DotExpression) Token() *lexer.Token {
	return d.token
}

// String returns a string representation of the DotExpression node.
func (d *DotExpression) String() string {
	return "(" + d.Left.String() + "." + d.Key + ")"
}

// Type returns the type of the DotExpression, which is the type of the
// map's values. For map := {a: true}, the type of map.a is bool.
func (d *DotExpression) Type() *Type {
	return d.T
}

// Token returns the token of the Evy source program associated with the
// TypeAssertion node.
func (t *TypeAssertion) Token() *lexer.Token {
	return t.token
}

// String returns a string representation of the TypeAssertion node.
func (t *TypeAssertion) String() string {
	return "(" + t.Left.String() + "." + "(" + t.T.String() + ")" + ")"
}

// Type returns the type of the TypeAssertion, which is the type that is
// asserted.
func (t *TypeAssertion) Type() *Type {
	return t.T
}

// Token returns the token of the Evy source program associated with the
// GroupExpression node.
func (d *GroupExpression) Token() *lexer.Token {
	return d.token
}

// String returns a string representation of the GroupExpression node.
func (d *GroupExpression) String() string {
	return d.Expr.String()
}

// Type returns the type of the GroupExpression, for example num for
// 2*(3+4).
func (d *GroupExpression) Type() *Type {
	return d.Expr.Type()
}

// Token returns the token of the Evy source program associated with the
// Decl node.
func (d *Decl) Token() *lexer.Token {
	return d.token
}

// String returns a string representation of the Decl node.
func (d *Decl) String() string {
	if d.Value == nil {
		return d.Var.String()
	}
	return d.Var.String() + "=" + d.Value.String()
}

// Type returns the type of the variable that is declared.
func (d *Decl) Type() *Type {
	return d.Var.T
}

// Token returns the token of the Evy source program associated with the
// TypedDeclStmt node.
func (d *TypedDeclStmt) Token() *lexer.Token {
	return d.token
}

// String returns a string representation of the TypedDeclStmt node.
func (d *TypedDeclStmt) String() string {
	return d.Decl.String()
}

// Type returns the type of the variable that is declared.
func (d *TypedDeclStmt) Type() *Type {
	return d.Decl.Var.T
}

// Token returns the token of the Evy source program associated with the
// InferredDeclStmt node.
func (d *InferredDeclStmt) Token() *lexer.Token {
	return d.token
}

// String returns a string representation of the InferredDeclStmt node.
func (d *InferredDeclStmt) String() string {
	return d.Decl.String()
}

// Type returns the type of the variable that is declared.
func (d *InferredDeclStmt) Type() *Type {
	return d.Decl.Var.T
}

// Token returns the token of the Evy source program associated with the
// ReturnStmt node.
func (r *ReturnStmt) Token() *lexer.Token {
	return r.token
}

// String returns a string representation of the ReturnStmt node.
func (r *ReturnStmt) String() string {
	if r.Value == nil {
		return "return"
	}
	return "return " + r.Value.String()
}

// Type returns the type of the value returned by the return statement.
func (r *ReturnStmt) Type() *Type {
	return r.T
}

func (*ReturnStmt) alwaysTerminates() bool {
	return true
}

// Token returns the token of the Evy source program associated with the
// BreakStmt node.
func (b *BreakStmt) Token() *lexer.Token {
	return b.token
}

// String returns a string representation of the eakStmt node.
func (*BreakStmt) String() string {
	return "break"
}

// Type returns [NONE_TYPE] for BreakStmt because the empty statement
// does not have a type.
func (*BreakStmt) Type() *Type {
	return NONE_TYPE
}

func (b *BreakStmt) alwaysTerminates() bool {
	return true
}

// Token returns the token of the Evy source program associated with the
// AssignmentStmt node.
func (a *AssignmentStmt) Token() *lexer.Token {
	return a.token
}

// String returns a string representation of the AssignmentStmt node.
func (a *AssignmentStmt) String() string {
	return a.Target.String() + " = " + a.Value.String()
}

// Type returns the type of the variable that is assigned.
func (a *AssignmentStmt) Type() *Type {
	return a.Target.Type()
}

// String returns a string representation of the FuncDefStmt node.
func (f *FuncDefStmt) String() string {
	s := make([]string, len(f.Params))
	for i, param := range f.Params {
		s[i] = param.String()
	}
	params := strings.Join(s, ", ")
	if f.VariadicParam != nil {
		params += f.VariadicParam.String() + "..."
	}
	signature := f.Name + "(" + params + ")"
	body := ""
	if f.Body != nil {
		body = f.Body.String()
	}
	return signature + "{\n" + body + "}\n"
}

// Type returns the return type of the function.
func (f *FuncDefStmt) Type() *Type {
	return f.ReturnType
}

// String returns a string representation of the IfStmt node.
func (i *IfStmt) String() string {
	result := "if " + i.IfBlock.String()
	for _, elseif := range i.ElseIfBlocks {
		result += "else if" + elseif.String()
	}
	if i.Else != nil {
		result += "else {\n" + i.Else.String() + "}\n"
	}
	return result
}

// Type returns [NONE_TYPE] for IfStmt because an if statement doest not
// have a type.
func (i *IfStmt) Type() *Type {
	return NONE_TYPE
}

func (i *IfStmt) alwaysTerminates() bool {
	if i.Else == nil || !i.Else.alwaysTerminates() {
		return false
	}
	if !i.IfBlock.alwaysTerminates() {
		return false
	}
	for _, b := range i.ElseIfBlocks {
		if !b.alwaysTerminates() {
			return false
		}
	}
	return true
}

// String returns a string representation of the EventHandlerStmt node.
func (e *EventHandlerStmt) String() string {
	body := e.Body.String()
	return "on " + e.Name + " {\n" + body + "}\n"
}

// Type returns [NONE_TYPE] for EventHandlerStmt because an event
// handler definition does not have a type.
func (e *EventHandlerStmt) Type() *Type {
	return NONE_TYPE
}

// String returns a string representation of the Var node.
func (v *Var) String() string {
	return v.Name
}

// Type returns the type of the variable.
func (v *Var) Type() *Type {
	return v.T
}

// String returns a string representation of the BlockStatement node.
func (b *BlockStatement) String() string {
	return newlineList(b.Statements)
}

// Type returns [NONE_TYPE] for BlockStatement because a block statement
// does not have a type.
func (b *BlockStatement) Type() *Type {
	return NONE_TYPE
}

func (b *BlockStatement) alwaysTerminates() bool {
	return b.alwaysTerms
}

func alwaysTerms(n Node) bool {
	r, ok := n.(interface{ alwaysTerminates() bool })
	return ok && r.alwaysTerminates()
}

// String returns a string representation of the WhileStmt node.
func (w *WhileStmt) String() string {
	return "while " + w.ConditionalBlock.String()
}

// Type returns [NONE_TYPE] for WhileStmt because a while statement does
// not have a type.
func (w *WhileStmt) Type() *Type {
	return NONE_TYPE
}

func (*WhileStmt) alwaysTerminates() bool {
	return false
}

// String returns a string representation of the ForStmt node.
func (f *ForStmt) String() string {
	header := "for "
	if f.LoopVar != nil {
		header += f.LoopVar.Name + " := "
	}
	header += f.Range.String()
	return header + " {\n" + f.Block.String() + "}"
}

// Type returns [NONE_TYPE] for ForStmt because a while statement does
// not have a type.
func (f *ForStmt) Type() *Type {
	return NONE_TYPE
}

// String returns a string representation of the StepRange node.
func (s *StepRange) String() string {
	start := "0"
	if s.Start != nil {
		start = s.Start.String()
	}
	stop := s.Stop.String()
	step := "1"
	if s.Step != nil {
		step = s.Step.String()
	}
	return start + " " + stop + " " + step
}

// Type returns [NUM_TYPE] for StepRange as a step range always
// represents a set of number value.
func (s *StepRange) Type() *Type {
	return NUM_TYPE
}

func (*ForStmt) alwaysTerminates() bool {
	return false
}

// String returns a string representation of the ConditionalBlock node.
func (c *ConditionalBlock) String() string {
	condition := "(" + c.Condition.String() + ")"
	return condition + " {\n" + c.Block.String() + "}"
}

// Type returns [NONE_TYPE] for ConditionalBlock because a conditional
// block statement does not have a type.
func (c *ConditionalBlock) Type() *Type {
	return NONE_TYPE
}

func (c *ConditionalBlock) alwaysTerminates() bool {
	return c.Block.alwaysTerminates()
}

// String returns a string representation of the BoolLiteral node.
func (b *BoolLiteral) String() string {
	return strconv.FormatBool(b.Value)
}

// Type returns [BOOL_TYPE] for BoolLiteral as a bool literal always has
// the bool type.
func (b *BoolLiteral) Type() *Type {
	return BOOL_TYPE
}

// String returns a string representation of the NumLiteral node.
func (n *NumLiteral) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

// Type returns [NUM_TYPE] for NumLiteral as a number literal always has
// the num type.
func (n *NumLiteral) Type() *Type {
	return NUM_TYPE
}

// String returns a string representation of the StringLiteral node.
func (s *StringLiteral) String() string {
	return fmt.Sprintf("%q", s.Value)
}

// Type returns [STRING_TYPE] for StringLiteral as a string literal
// always has the string type.
func (s *StringLiteral) Type() *Type {
	return STRING_TYPE
}

// String returns a string representation of the ArrayLiteral node.
func (a *ArrayLiteral) String() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// Type returns the type of the array literal, such as []num for [1 2 3].
func (a *ArrayLiteral) Type() *Type {
	return a.T
}

// String returns a string representation of the MapLiteral node.
func (m *MapLiteral) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for _, key := range m.Order {
		val := m.Pairs[key]
		pairs = append(pairs, key+":"+val.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// Type returns the type of the map literal such as {}num for {a:1 b:2}.
func (m *MapLiteral) Type() *Type {
	return m.T
}

// IsCompositeLiteral returns true if the node is an array or map
// literal.
func IsCompositeLiteral(n Node) bool {
	_, aok := n.(*ArrayLiteral)
	_, mok := n.(*MapLiteral)
	return aok || mok
}

func newlineList(nodes []Node) string {
	lines := make([]string, len(nodes))
	for i, n := range nodes {
		lines[i] = n.String()
	}
	return strings.Join(lines, "\n") + "\n"
}

func zeroValue(t *Type, tt *lexer.Token) Node {
	switch t.Name {
	case NUM:
		return &NumLiteral{Value: 0, token: tt}
	case STRING:
		return &StringLiteral{Value: "", token: tt}
	case BOOL:
		return &BoolLiteral{Value: false, token: tt}
	case ANY:
		return &BoolLiteral{Value: false, token: tt}
	case ARRAY:
		return &ArrayLiteral{T: t, token: tt}
	case MAP:
		return &MapLiteral{T: t, token: tt}
	}
	return nil
}
