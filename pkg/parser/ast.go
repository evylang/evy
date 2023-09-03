package parser

import (
	"fmt"
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

type Node interface {
	Token() *lexer.Token
	String() string
	Type() *Type
}

type Program struct {
	token              *lexer.Token
	Statements         []Node
	EventHandlers      map[string]*EventHandlerStmt
	CalledBuiltinFuncs []string

	alwaysTerms bool
	formatting  *formatting
}

type EmptyStmt struct {
	token *lexer.Token // The NL token
}

type FuncCallStmt struct {
	token    *lexer.Token // The IDENT of the function
	FuncCall *FuncCall
}

type FuncCall struct {
	token     *lexer.Token // The IDENT of the function
	Name      string
	Arguments []Node
	FuncDef   *FuncDefStmt
}

type UnaryExpression struct {
	token *lexer.Token // The unary operation token, e.g. !
	Op    Operator
	Right Node
}

type BinaryExpression struct {
	T     *Type
	token *lexer.Token // The binary operation token, e.g. +
	Op    Operator
	Left  Node
	Right Node
}

type IndexExpression struct {
	T     *Type
	token *lexer.Token // The [ token
	Left  Node
	Index Node
}

type SliceExpression struct {
	T     *Type
	token *lexer.Token // The [ token
	Left  Node
	Start Node
	End   Node
}

type DotExpression struct {
	T     *Type
	token *lexer.Token // The . token
	Left  Node
	Key   string // m := { age: 42}; m.age => key: "age"
}

type TypeAssertion struct {
	T     *Type
	token *lexer.Token
	Left  Node
}

type GroupExpression struct {
	token *lexer.Token
	Expr  Node
}

type Decl struct {
	token *lexer.Token
	Var   *Var
	Value Node // literal, expression, assignable, ...
}

type TypedDeclStmt struct {
	token *lexer.Token
	Decl  *Decl
}

type InferredDeclStmt struct {
	token *lexer.Token
	Decl  *Decl
}

type AssignmentStmt struct {
	token  *lexer.Token
	Target Node // Variable, index or field expression
	Value  Node // literal, expression, assignable, ...
}

type ReturnStmt struct {
	token *lexer.Token
	Value Node // literal, expression, assignable, ...
	T     *Type
}

type BreakStmt struct {
	token *lexer.Token
}

func (f *FuncDefStmt) Token() *lexer.Token {
	return f.token
}

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

func (i *IfStmt) Token() *lexer.Token {
	return i.token
}

type IfStmt struct {
	token        *lexer.Token
	IfBlock      *ConditionalBlock
	ElseIfBlocks []*ConditionalBlock
	Else         *BlockStatement
}

func (w *WhileStmt) Token() *lexer.Token {
	return w.token
}

type WhileStmt struct {
	ConditionalBlock
}

func (f *ForStmt) Token() *lexer.Token {
	return f.token
}

type ForStmt struct {
	token *lexer.Token

	LoopVar *Var
	Range   Node // StepRange or array/map/string expression

	Block *BlockStatement
}

func (s *StepRange) Token() *lexer.Token {
	return s.token
}

type StepRange struct {
	token *lexer.Token

	Start Node // num expression or nil
	Stop  Node // num expression
	Step  Node // num expression or nil
}

func (c *ConditionalBlock) Token() *lexer.Token {
	return c.token
}

type ConditionalBlock struct {
	token     *lexer.Token
	Condition Node // must be of type bool
	Block     *BlockStatement
}

func (e *EventHandlerStmt) Token() *lexer.Token {
	return e.token
}

type EventHandlerStmt struct {
	token  *lexer.Token // The "on" token
	Name   string
	Params []*Var
	Body   *BlockStatement
}

func (v *Var) Token() *lexer.Token {
	return v.token
}

type Var struct {
	token  *lexer.Token
	Name   string
	T      *Type
	isUsed bool
}

func (b *BlockStatement) Token() *lexer.Token {
	return b.token
}

type BlockStatement struct {
	token       *lexer.Token // the NL before the first statement
	Statements  []Node
	alwaysTerms bool
}

func (b *BoolLiteral) Token() *lexer.Token {
	return b.token
}

type BoolLiteral struct {
	token *lexer.Token
	Value bool
}

func (n *NumLiteral) Token() *lexer.Token {
	return n.token
}

type NumLiteral struct {
	token *lexer.Token
	Value float64
}

func (s *StringLiteral) Token() *lexer.Token {
	return s.token
}

type StringLiteral struct {
	token *lexer.Token
	Value string
}

func (a *ArrayLiteral) Token() *lexer.Token {
	return a.token
}

type ArrayLiteral struct {
	token    *lexer.Token
	Elements []Node
	T        *Type
}

func (m *MapLiteral) Token() *lexer.Token {
	return m.token
}

type MapLiteral struct {
	token *lexer.Token
	Pairs map[string]Node
	Order []string // Track insertion order of keys for deterministic output.
	T     *Type
}

func (p *Program) Token() *lexer.Token {
	return p.token
}

func (p *Program) String() string {
	return newlineList(p.Statements)
}

func (p *Program) Format() string {
	var sb strings.Builder
	p.formatting.w = &sb
	p.formatting.format(p) // todo: maybe formatting.formatProgram(prog, w)
	return sb.String()
}

func (*Program) Type() *Type {
	return NONE_TYPE
}

func (p *Program) alwaysTerminates() bool {
	return p.alwaysTerms
}

func (e *EmptyStmt) Token() *lexer.Token {
	return e.token
}

func (e *EmptyStmt) String() string {
	return ""
}

func (*EmptyStmt) Type() *Type { return NONE_TYPE }

func (f *FuncCall) Token() *lexer.Token {
	return f.token
}

func (f *FuncCall) String() string {
	s := make([]string, len(f.Arguments))
	for i, arg := range f.Arguments {
		s[i] = arg.String()
	}
	args := strings.Join(s, ", ")
	return f.Name + "(" + args + ")"
}

func (f *FuncCall) Type() *Type {
	return f.FuncDef.ReturnType
}

func (f *FuncCallStmt) Token() *lexer.Token {
	return f.token
}

func (f *FuncCallStmt) String() string {
	return f.FuncCall.String()
}

func (f *FuncCallStmt) Type() *Type {
	return f.FuncCall.FuncDef.ReturnType
}

func (u *UnaryExpression) Token() *lexer.Token {
	return u.token
}

func (u *UnaryExpression) String() string {
	return "(" + u.Op.String() + u.Right.String() + ")"
}

func (u *UnaryExpression) Type() *Type {
	return u.Right.Type()
}

func (b *BinaryExpression) Token() *lexer.Token {
	return b.token
}

func (b *BinaryExpression) String() string {
	if b.Op == OP_AND || b.Op == OP_OR {
		return "(" + b.Left.String() + " " + b.Op.String() + " " + b.Right.String() + ")"
	}
	return "(" + b.Left.String() + b.Op.String() + b.Right.String() + ")"
}

func (b *BinaryExpression) Type() *Type {
	return b.T
}

func (i *IndexExpression) Token() *lexer.Token {
	return i.token
}

func (i *IndexExpression) String() string {
	return "(" + i.Left.String() + "[" + i.Index.String() + "])"
}

func (i *IndexExpression) Type() *Type {
	return i.T
}

func (s *SliceExpression) Token() *lexer.Token {
	return s.token
}

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

func (s *SliceExpression) Type() *Type {
	return s.T
}

func (d *DotExpression) Token() *lexer.Token {
	return d.token
}

func (d *DotExpression) String() string {
	return "(" + d.Left.String() + "." + d.Key + ")"
}

func (d *DotExpression) Type() *Type {
	return d.T
}

func (t *TypeAssertion) Token() *lexer.Token {
	return t.token
}

func (t *TypeAssertion) String() string {
	return "(" + t.Left.String() + "." + "(" + t.T.String() + ")" + ")"
}

func (t *TypeAssertion) Type() *Type {
	return t.T
}

func (d *GroupExpression) Token() *lexer.Token {
	return d.token
}

func (d *GroupExpression) String() string {
	return d.Expr.String()
}

func (d *GroupExpression) Type() *Type {
	return d.Expr.Type()
}

func (d *Decl) Token() *lexer.Token {
	return d.token
}

func (d *Decl) String() string {
	if d.Value == nil {
		return d.Var.String()
	}
	return d.Var.String() + "=" + d.Value.String()
}

func (d *Decl) Type() *Type {
	return d.Var.T
}

func (d *TypedDeclStmt) Token() *lexer.Token {
	return d.token
}

func (d *TypedDeclStmt) String() string {
	return d.Decl.String()
}

func (d *TypedDeclStmt) Type() *Type {
	return d.Decl.Var.T
}

func (d *InferredDeclStmt) Token() *lexer.Token {
	return d.token
}

func (d *InferredDeclStmt) String() string {
	return d.Decl.String()
}

func (d *InferredDeclStmt) Type() *Type {
	return d.Decl.Var.T
}

func (r *ReturnStmt) Token() *lexer.Token {
	return r.token
}

func (r *ReturnStmt) String() string {
	if r.Value == nil {
		return "return"
	}
	return "return " + r.Value.String()
}

func (r *ReturnStmt) Type() *Type {
	return r.T
}

func (*ReturnStmt) alwaysTerminates() bool {
	return true
}

func (b *BreakStmt) Token() *lexer.Token {
	return b.token
}

func (*BreakStmt) String() string {
	return "break"
}

func (*BreakStmt) Type() *Type {
	return NONE_TYPE
}

func (b *BreakStmt) alwaysTerminates() bool {
	return true
}

func (a *AssignmentStmt) Token() *lexer.Token {
	return a.token
}

func (a *AssignmentStmt) String() string {
	return a.Target.String() + " = " + a.Value.String()
}

func (a *AssignmentStmt) Type() *Type {
	return a.Target.Type()
}

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

func (f *FuncDefStmt) Type() *Type {
	return f.ReturnType
}

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

func (e *EventHandlerStmt) String() string {
	body := e.Body.String()
	return "on " + e.Name + " {\n" + body + "}\n"
}

func (e *EventHandlerStmt) Type() *Type {
	return NONE_TYPE
}

func (v *Var) String() string {
	return v.Name
}

func (v *Var) Type() *Type {
	return v.T
}

func (b *BlockStatement) String() string {
	return newlineList(b.Statements)
}

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

func (w *WhileStmt) String() string {
	return "while " + w.ConditionalBlock.String()
}

func (w *WhileStmt) Type() *Type {
	return w.ConditionalBlock.Type()
}

func (*WhileStmt) alwaysTerminates() bool {
	return false
}

func (f *ForStmt) String() string {
	header := "for "
	if f.LoopVar != nil {
		header += f.LoopVar.Name + " := "
	}
	header += f.Range.String()
	return header + " {\n" + f.Block.String() + "}"
}

func (f *ForStmt) Type() *Type {
	return f.Block.Type()
}

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

func (s *StepRange) Type() *Type {
	return NUM_TYPE
}

func (*ForStmt) alwaysTerminates() bool {
	return false
}

func (c *ConditionalBlock) String() string {
	condition := "(" + c.Condition.String() + ")"
	return condition + " {\n" + c.Block.String() + "}"
}

func (c *ConditionalBlock) Type() *Type {
	return NONE_TYPE
}

func (c *ConditionalBlock) alwaysTerminates() bool {
	return c.Block.alwaysTerminates()
}

func (b *BoolLiteral) String() string {
	return strconv.FormatBool(b.Value)
}

func (b *BoolLiteral) Type() *Type {
	return BOOL_TYPE
}

func (n *NumLiteral) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

func (n *NumLiteral) Type() *Type {
	return NUM_TYPE
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("%q", s.Value)
}

func (s *StringLiteral) Type() *Type {
	return STRING_TYPE
}

func (a *ArrayLiteral) String() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

func (a *ArrayLiteral) Type() *Type {
	return a.T
}

func (m *MapLiteral) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for _, key := range m.Order {
		val := m.Pairs[key]
		pairs = append(pairs, key+":"+val.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (m *MapLiteral) Type() *Type {
	return m.T
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
