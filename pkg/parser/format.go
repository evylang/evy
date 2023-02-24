package parser

import (
	"io"
	"strings"
)

func newFormatting() *formatting {
	return &formatting{
		wss:         map[*BinaryExpression]bool{},
		comments:    map[Node]string{},
		indentLevel: -1,
	}
}

type formatting struct {
	w io.StringWriter

	wss         map[*BinaryExpression]bool
	comments    map[Node]string
	indentLevel int
}

func (f *formatting) recordComment(n Node, comment string) {
	f.comments[n] = comment
}

func (f *formatting) recordWSS(n *BinaryExpression) {
	f.wss[n] = true
}

func (f *formatting) format(n Node) {
	switch n := n.(type) {
	case *Program:
		f.writeStmts(n.Statements)
		f.writeComment(n)
	case *EmptyStmt:
		f.writeComment(n)
	case *TypedDeclStmt:
		f.format(n.Decl)
		f.writeComment(n)
	case *InferredDeclStmt:
		f.format(n.Decl.Var)
		f.write(" := ")
		f.format(n.Decl.Value)
		f.writeComment(n)
	case *AssignmentStmt:
		f.format(n.Target)
		f.write(" = ")
		f.format(n.Value)
		f.writeComment(n)
	case *IfStmt:
		f.formatIfStmt(n)
	case *WhileStmt:
		f.write("while ")
		f.format(&n.ConditionalBlock)
	case *BreakStmt:
		f.write("break")
		f.writeComment(n)
	case *ForStmt:
		f.formatForStmt(n)
	case *ReturnStmt:
		f.formatReturnStmt(n)
	case *FuncDeclStmt:
		f.formatFuncDeclStmt(n)
	case *FuncCallStmt:
		f.format(n.FuncCall)
		f.writeComment(n)
	case *EventHandlerStmt:
		f.formatEventHandlerStmt(n)
	case *Decl:
		f.writeDecl(n.Var)
	case *Var:
		f.write(n.Name)
	case *ConditionalBlock:
		f.format(n.Condition)
		f.writeComment(n)
		f.writeLn()
		f.format(n.Block)
	case *BlockStatement:
		f.writeStmts(n.Statements)
		f.indent()
		f.write("end")
		f.writeComment(n)
	case *StepRange:
		f.formatStepRange(n)
	case *FuncCall:
		f.formatFuncCall(n)
	case *UnaryExpression:
		f.writes(n.Op.String(), n.Right.String())
	case *BinaryExpression:
		f.format(n.Left)
		f.writeWSS(n)
		f.write(n.Op.String())
		f.writeWSS(n)
		f.format(n.Right)
	case *IndexExpression:
		f.format(n.Left)
		f.write("[")
		f.format(n.Index)
		f.write("]")
	case *SliceExpression:
		f.format(n.Left)
		f.write("[")
		f.formatIfNotNil(n.Start)
		f.write(":")
		f.formatIfNotNil(n.End)
		f.write("]")
	case *DotExpression:
		f.format(n.Left)
		f.writes(".", n.Key)
	case *GroupExpression:
		f.write("(")
		f.format(n.Expr)
		f.write(")")
	case *Bool:
		f.write(n.String())
	case *NumLiteral:
		f.write(n.String())
	case *StringLiteral:
		f.writes(`"`, n.Value, `"`)
	case *ArrayLiteral:
		f.formatArrayLiteral(n)
	case *MapLiteral:
		f.formatMapLiteral(n)
	default:
		f.write("format unimplemented for " + n.String())
	}
}

func (f *formatting) formatIfStmt(s *IfStmt) {
	f.write("if ")
	f.format(s.IfBlock.Condition)
	f.writeComment(s.IfBlock) // if comment
	f.write("\n")
	f.writeStmts(s.IfBlock.Block.Statements)
	for _, elseif := range s.ElseIfBlocks {
		f.indent()
		f.write("else if ")
		f.format(elseif.Condition)
		f.writeComment(elseif) // else if comment
		f.write("\n")
		f.writeStmts(elseif.Block.Statements)
	}
	if s.Else != nil {
		f.indent()
		f.write("else")
		f.writeComment(s.Else) // else comment
		f.write("\n")
		f.writeStmts(s.Else.Statements)
	}
	f.indent()
	f.write("end")
	f.writeComment(s) // end comment
}

func (f *formatting) formatForStmt(s *ForStmt) {
	f.write("for ")
	if s.LoopVar != nil {
		f.writes(s.LoopVar.Name, " := ")
	}
	f.write("range ")
	f.format(s.Range)
	f.writeComment(s)
	f.write("\n")
	f.format(s.Block)
}

func (f *formatting) formatReturnStmt(s *ReturnStmt) {
	f.write("return")
	if s.Value != nil {
		f.write(" ")
		f.format(s.Value)
	}
	f.writeComment(s)
}

func (f *formatting) formatFuncDeclStmt(s *FuncDeclStmt) {
	f.writes("func ", s.Name)
	if s.ReturnType != NONE_TYPE {
		f.write(":")
		f.formatType(s.ReturnType)
	}
	for _, param := range s.Params {
		f.write(" ")
		f.writeDecl(param)
	}
	if s.VariadicParam != nil {
		f.write(" ")
		f.writeDecl(s.VariadicParam)
		f.write("...")
	}
	f.writeComment(s)
	f.write("\n")
	f.format(s.Body)
}

func (f *formatting) formatEventHandlerStmt(s *EventHandlerStmt) {
	f.writes("on ", s.Name)
	for _, param := range s.Params {
		f.write(" ")
		f.writeDecl(param)
	}
	f.writeComment(s)
	f.write("\n")
	f.format(s.Body)
}

func (f *formatting) formatStepRange(n *StepRange) {
	if n.Start != nil {
		f.format(n.Start)
		f.write(" ")
	}
	f.format(n.Stop)
	if n.Step != nil {
		f.write(" ")
		f.format(n.Step)
	}
}

func (f *formatting) formatFuncCall(n *FuncCall) {
	f.write(n.Name)
	for _, arg := range n.Arguments {
		f.write(" ")
		f.format(arg)
	}
}

func (f *formatting) formatArrayLiteral(n *ArrayLiteral) {
	// TODO: handle multilines
	f.write("[")
	length := len(n.Elements)
	for i, el := range n.Elements {
		f.format(el)
		if i+1 < length {
			f.write(" ")
		}
	}
	f.write("]")
}

func (f *formatting) formatMapLiteral(n *MapLiteral) {
	// TODO: handle multilines
	f.write("{")
	length := len(n.Pairs)
	for i, key := range n.Order {
		f.writes(key, ":")
		f.format(n.Pairs[key])
		if i+1 < length {
			f.write(" ")
		}
	}
	f.write("}")
}

func (f *formatting) formatType(t *Type) {
	f.write(t.Name.String())
	if t.Sub != nil && t != GENERIC_ARRAY && t != GENERIC_MAP {
		f.formatType(t.Sub)
	}
}

func (f *formatting) formatIfNotNil(n Node) {
	if n != nil {
		f.format(n)
	}
}

func (f *formatting) write(s string) {
	if _, err := f.w.WriteString(s); err != nil {
		panic("formatting.write: " + err.Error())
	}
}

func (f *formatting) writes(strs ...string) {
	for _, str := range strs {
		f.write(str)
	}
}

func (f *formatting) writeLn() {
	f.write("\n")
}

func (f *formatting) writeComment(n Node) {
	c := f.comments[n]
	if c == "" {
		return
	}
	if _, ok := n.(*EmptyStmt); !ok {
		f.write(" ")
	}
	f.write(strings.TrimSpace(c))
}

func (f *formatting) writeDecl(n *Var) {
	f.format(n)
	f.write(":")
	f.formatType(n.Type())
}

func (f *formatting) indent() {
	for i := 0; i < f.indentLevel; i++ {
		f.write("    ")
	}
}

func (f *formatting) writeStmts(stmts []Node) {
	f.indentLevel++

	if len(stmts) == 0 {
		stmts = []Node{&EmptyStmt{}} // write at least a single new line
	}

	empty := false
	for _, stmt := range stmts {
		if empty = f.writeEmptyStmt(stmt, empty); empty {
			continue
		}
		f.indent()
		f.format(stmt)
		f.writeLn()
	}

	f.indentLevel--
}

func (f *formatting) writeEmptyStmt(n Node, lastEmpty bool) bool {
	_, ok := n.(*EmptyStmt)
	if ok && f.comments[n] == "" {
		if !lastEmpty {
			f.writeLn()
		}
		return true
	}
	return false
}

func (f *formatting) writeWSS(n *BinaryExpression) {
	if !f.wss[n] {
		f.write(" ")
	}
}
