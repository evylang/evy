package parser

import (
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

type Node interface {
	String() string
}

type Program struct {
	Statements []Node
}

type FunctionCall struct {
	Token     *lexer.Token // The IDENT of the function
	Name      string
	Arguments []*Term
}

type Term struct {
	Token *lexer.Token
	Type  *TypeNode
	Value Node
}

type Declaration struct {
	Token *lexer.Token
	Var   *Var
	Value Node // literal, expression, assignable, ...
}

type Var struct {
	Token *lexer.Token
	Name  string
	Type  *TypeNode
}

func zeroValue(t Type, token *lexer.Token) Node {
	switch t {
	case NUM:
		return &NumLiteral{Token: token, Value: 0}
	case STRING:
		return &StringLiteral{Token: token, Value: ""}
	case BOOL:
		return &Bool{Token: token, Value: false}
	case ANY:
		return &Bool{Token: token, Value: false}
	case ARRAY:
		return &ArrayLiteral{Token: token}
	case MAP:
		return &MapLiteral{Token: token}
	}
	return ILLEGAL_TYPE
}

type TypeNode struct {
	Name Type      // string, num, bool, composite types array, map
	Sub  *TypeNode // e.g.: `[]int` : Type{Name: "array", Sub: &Type{Name: "int"} }
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
	Elements []*Term
}

type MapLiteral struct {
	Token *lexer.Token
	Pairs map[string]*Term
}

func (v *Var) String() string           { return v.Name + ":" + v.Type.String() }
func (n *NumLiteral) String() string    { return strconv.FormatFloat(n.Value, 'f', -1, 64) }
func (s *StringLiteral) String() string { return "'" + s.Value + "'" }
func (b *Bool) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}
func (a *ArrayLiteral) String() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

func (m *MapLiteral) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for key, val := range m.Pairs {
		pairs = append(pairs, key+":"+val.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (f *FunctionCall) String() string {
	args := make([]string, len(f.Arguments))
	for i, arg := range f.Arguments {
		args[i] = arg.String()
	}
	argList := strings.Join(args, ", ")
	return f.Name + "(" + argList + ")"
}

func (t *Term) String() string {
	return t.Value.String()
}

func (s *Declaration) String() string {
	if s.Value == nil {
		return s.Var.String()
	}
	return s.Var.String() + "=" + s.Value.String()

}

func (t *TypeNode) String() string {
	if t.Sub == nil {
		return t.Name.String()
	}
	return t.Name.String() + " " + t.Sub.String()
}

func (p *Program) String() string {
	stmts := make([]string, len(p.Statements))

	for i, s := range p.Statements {
		stmts[i] = s.String()
	}
	return strings.Join(stmts, "\n") + "\n"
}
