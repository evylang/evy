package parser

import (
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

type Node interface {
	String() string
	Type() *Type
}

type Program struct {
	Statements       []Node
	alwaysTerminates bool
}

type FunctionCall struct {
	Token     *lexer.Token // The IDENT of the function
	Name      string
	Arguments []Node
	FuncDecl  *FuncDecl
}

type UnaryExpression struct {
	Token *lexer.Token // The unary operation token, e.g. !
	Op    Operator
	Right Node
}

type BinaryExpression struct {
	T     *Type
	Token *lexer.Token // The binary operation token, e.g. +
	Op    Operator
	Left  Node
	Right Node
}

type IndexExpression struct {
	T     *Type
	Token *lexer.Token // The [ token
	Left  Node
	Index Node
}

type SliceExpression struct {
	T     *Type
	Token *lexer.Token // The [ token
	Left  Node
	Start Node
	End   Node
}

type DotExpression struct {
	T     *Type
	Token *lexer.Token // The . token
	Left  Node
	Key   string // m := { age: 42}; m.age => key: "age"
}

type Declaration struct {
	Token *lexer.Token
	Var   *Var
	Value Node // literal, expression, assignable, ...
}

type Assignment struct {
	Token  *lexer.Token
	Target Node // Variable, index or field expression
	Value  Node // literal, expression, assignable, ...
}

type Return struct {
	Token *lexer.Token
	Value Node // literal, expression, assignable, ...
	T     *Type
}

type Break struct {
	Token *lexer.Token
}

type FuncDecl struct {
	Token         *lexer.Token // The 'func' token
	Name          string
	Params        []*Var
	VariadicParam *Var
	ReturnType    *Type
	Body          *BlockStatement
}

type If struct {
	Token        *lexer.Token
	IfBlock      *ConditionalBlock
	ElseIfBlocks []*ConditionalBlock
	Else         *BlockStatement
}

type While struct {
	ConditionalBlock
}

type For struct {
	Token *lexer.Token

	LoopVar *Var
	Range   Node // StepRange or array/map/string expression

	Block *BlockStatement
}

type StepRange struct {
	Token *lexer.Token

	Start Node // num expression or nil
	Stop  Node // num expression
	Step  Node // num expression or nil
}

type ConditionalBlock struct {
	Token     *lexer.Token
	Condition Node // must be of type bool
	Block     *BlockStatement
}

type EventHandler struct {
	Token  *lexer.Token // The 'on' token
	Name   string
	Params []*Var
	Body   *BlockStatement
}

type Var struct {
	Token  *lexer.Token
	Name   string
	T      *Type
	isUsed bool
}

type BlockStatement struct {
	Token            *lexer.Token // the NL before the first statement
	Statements       []Node
	alwaysTerminates bool
}

type Bool struct {
	Token *lexer.Token
	Value bool
}

type NumLiteral struct {
	Token *lexer.Token
	Value float64
}

type StringLiteral struct {
	Token *lexer.Token
	Value string
}

type ArrayLiteral struct {
	Token    *lexer.Token
	Elements []Node
	T        *Type
}

type MapLiteral struct {
	Token *lexer.Token
	Pairs map[string]Node
	Order []string // Track insertion order of keys for deterministic output.
	T     *Type
}

func (p *Program) String() string {
	return newlineList(p.Statements)
}

func (*Program) Type() *Type {
	return NONE_TYPE
}

func (p *Program) AlwaysTerminates() bool {
	return p.alwaysTerminates
}

func (f *FunctionCall) String() string {
	s := make([]string, len(f.Arguments))
	for i, arg := range f.Arguments {
		s[i] = arg.String()
	}
	args := strings.Join(s, ", ")
	return f.Name + "(" + args + ")"
}

func (f *FunctionCall) Type() *Type {
	return f.FuncDecl.ReturnType
}

func (u *UnaryExpression) String() string {
	return "(" + u.Op.String() + u.Right.String() + ")"
}

func (u *UnaryExpression) Type() *Type {
	return u.Right.Type()
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

func (i *IndexExpression) String() string {
	return "(" + i.Left.String() + "[" + i.Index.String() + "])"
}

func (i *IndexExpression) Type() *Type {
	return i.T
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

func (d *DotExpression) String() string {
	return "(" + d.Left.String() + "." + d.Key + ")"
}

func (d *DotExpression) Type() *Type {
	return d.T
}

func (d *Declaration) String() string {
	if d.Value == nil {
		return d.Var.String()
	}
	return d.Var.String() + "=" + d.Value.String()
}

func (d *Declaration) Type() *Type {
	return d.Var.T
}

func (r *Return) String() string {
	if r.Value == nil {
		return "return"
	}
	return "return " + r.Value.String()
}

func (r *Return) Type() *Type {
	return r.T
}

func (*Return) AlwaysTerminates() bool {
	return true
}

func (*Break) String() string {
	return "break"
}

func (*Break) Type() *Type {
	return NONE_TYPE
}

func (b *Break) AlwaysTerminates() bool {
	return true
}

func (a *Assignment) String() string {
	return a.Target.String() + " = " + a.Value.String()
}

func (a *Assignment) Type() *Type {
	return a.Target.Type()
}

func (f *FuncDecl) String() string {
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

func (f *FuncDecl) Type() *Type {
	return f.ReturnType
}

func (i *If) String() string {
	result := "if " + i.IfBlock.String()
	for _, elseif := range i.ElseIfBlocks {
		result += "else if" + elseif.String()
	}
	if i.Else != nil {
		result += "else {\n" + i.Else.String() + "}\n"
	}
	return result
}

func (i *If) Type() *Type {
	return NONE_TYPE
}

func (i *If) AlwaysTerminates() bool {
	if i.Else == nil || !i.Else.AlwaysTerminates() {
		return false
	}
	if !i.IfBlock.AlwaysTerminates() {
		return false
	}
	for _, b := range i.ElseIfBlocks {
		if !b.AlwaysTerminates() {
			return false
		}
	}
	return true
}

func (e *EventHandler) String() string {
	body := e.Body.String()
	return "on " + e.Name + " {\n" + body + "}\n"
}

func (e *EventHandler) Type() *Type {
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

func (b *BlockStatement) AlwaysTerminates() bool {
	return b.alwaysTerminates
}

func alwaysTerminates(n Node) bool {
	r, ok := n.(interface{ AlwaysTerminates() bool })
	return ok && r.AlwaysTerminates()
}

func (w *While) String() string {
	return "while " + w.ConditionalBlock.String()
}

func (w *While) Type() *Type {
	return w.ConditionalBlock.Type()
}

func (*While) AlwaysTerminates() bool {
	return false
}

func (f *For) String() string {
	header := "for " + f.LoopVar.Name + " := " + f.Range.String()
	return header + " {\n" + f.Block.String() + "}"
}

func (f *For) Type() *Type {
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

func (*For) AlwaysTerminates() bool {
	return false
}

func (c *ConditionalBlock) String() string {
	condition := "(" + c.Condition.String() + ")"
	return condition + " {\n" + c.Block.String() + "}"
}

func (c *ConditionalBlock) Type() *Type {
	return NONE_TYPE
}

func (c *ConditionalBlock) AlwaysTerminates() bool {
	return c.Block.AlwaysTerminates()
}

func (b *Bool) String() string {
	return strconv.FormatBool(b.Value)
}

func (b *Bool) Type() *Type {
	return BOOL_TYPE
}

func (n *NumLiteral) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

func (n *NumLiteral) Type() *Type {
	return NUM_TYPE
}

func (s *StringLiteral) String() string {
	return "'" + s.Value + "'"
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

func zeroValue(t TypeName) Node {
	switch t {
	case NUM:
		return &NumLiteral{Value: 0}
	case STRING:
		return &StringLiteral{Value: ""}
	case BOOL:
		return &Bool{Value: false}
	case ANY:
		return &Bool{Value: false}
	case ARRAY:
		return &ArrayLiteral{}
	case MAP:
		return &MapLiteral{}
	}
	return nil
}
