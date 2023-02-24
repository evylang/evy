package parser

import (
	"fmt"
	"io"
	"strings"
)

func newFormatting() *formatting {
	return &formatting{
		indentLevel: -1,
	}
}

type formatting struct {
	w io.StringWriter

	indentLevel int
}

func (f *formatting) format(n Node) {
	switch n := n.(type) {
	case *Program:
		f.writeStmts(n.Statements)
	case *EmptyStmt:
		f.writes(strings.TrimSpace(n.comment))
	case *TypedDeclStmt:
		f.format(n.Decl)
		f.writeComment(n.comment)
	case *InferredDeclStmt:
		f.format(n.Decl.Var)
		f.write(" := ")
		f.format(n.Decl.Value)
		f.writeComment(n.comment)
	case *AssignmentStmt:
		f.format(n.Target)
		f.write(" = ")
		f.format(n.Value)
		f.writeComment(n.comment)
	case *IfStmt:
		f.formatIfStmt(n)
	case *WhileStmt:
		f.write("while ")
		f.format(&n.ConditionalBlock)
	case *BreakStmt:
		f.write("break")
		f.writeComment(n.comment)
	case *ForStmt:
		f.formatForStmt(n)
	case *ReturnStmt:
		f.formatReturnStmt(n)
	case *FuncDeclStmt:
		f.formatFuncDeclStmt(n)
	case *FuncCallStmt:
		f.format(n.FuncCall)
		f.writeComment(n.comment)
	case *EventHandlerStmt:
		f.formatEventHandlerStmt(n)
	case *Decl:
		f.writeDecl(n.Var)
	case *Var:
		f.write(n.Name)
	case *ConditionalBlock:
		f.format(n.Condition)
		f.writeComment(n.comment)
		f.writeLn()
		f.format(n.Block)
	case *BlockStatement:
		f.writeStmts(n.Statements)
		f.indent()
		f.write("end")
		f.writeComment(n.comment)
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
		f.write(fmt.Sprintf("format unimplemented for %v", n))
	}
}

func (f *formatting) formatIfStmt(s *IfStmt) {
	f.write("if ")
	f.format(s.IfBlock.Condition)
	f.writeComment(s.IfBlock.comment) // if comment
	f.write("\n")
	f.writeStmts(s.IfBlock.Block.Statements)
	for _, elseif := range s.ElseIfBlocks {
		f.indent()
		f.write("else if ")
		f.format(elseif.Condition)
		f.writeComment(elseif.comment) // else if comment
		f.write("\n")
		f.writeStmts(elseif.Block.Statements)
	}
	if s.Else != nil {
		f.indent()
		f.write("else")
		f.writeComment(s.Else.comment) // else comment
		f.write("\n")
		f.writeStmts(s.Else.Statements)
	}
	f.indent()
	f.write("end")
	f.writeComment(s.comment) // end comment
}

func (f *formatting) formatForStmt(s *ForStmt) {
	f.write("for ")
	if s.LoopVar != nil {
		f.writes(s.LoopVar.Name, " := ")
	}
	f.write("range ")
	f.format(s.Range)
	f.writeComment(s.comment)
	f.write("\n")
	f.format(s.Block)
}

func (f *formatting) formatReturnStmt(s *ReturnStmt) {
	f.write("return")
	if s.Value != nil {
		f.write(" ")
		f.format(s.Value)
	}
	f.writeComment(s.comment)
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
	f.writeComment(s.comment)
	f.write("\n")
	f.format(s.Body)
}

func (f *formatting) formatEventHandlerStmt(s *EventHandlerStmt) {
	f.writes("on ", s.Name)
	for _, param := range s.Params {
		f.write(" ")
		f.writeDecl(param)
	}
	f.writeComment(s.comment)
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
	multi := formatMultiline(n.multilines)
	if len(multi) == 0 {
		f.write("[]")
		return
	}
	f.write("[")
	if multi[0].isComment() {
		f.write(" ")
	}

	length := len(multi)
	idx := 0
	for i, m := range multi {
		if m == multilineEl {
			f.format(n.Elements[idx])
			idx++
			if i+1 < length && !multi[i+1].isNL() {
				f.write(" ") // add space before next element or comment
			}
			continue
		}
		// newline or comment
		f.write(string(m))

		if i+1 == length || !multi[i+1].isNL() { // next is element, comment or `]`
			f.indent()
			if i+1 < length {
				f.write(indentStr) // one extra indent for element or comment
			}
		}
	}
	f.write("]")
}

func (f *formatting) formatMapLiteral(n *MapLiteral) {
	multi := formatMultiline(n.multilines)
	if len(multi) == 0 {
		f.write("{}")
		return
	}
	f.write("{")
	if multi[0].isComment() {
		f.write(" ")
	}

	length := len(multi)
	for i, m := range multi {
		if m.isKey() { // key
			key := string(m)
			f.writes(key, ":")
			f.format(n.Pairs[key])
			if i+1 < length && !multi[i+1].isNL() {
				f.write(" ") // add space before next pair or comment
			}
			continue
		}
		// newline or comment
		f.writes(string(m))
		if i+1 == length || !multi[i+1].isNL() { // next is pair, comment or `}`
			f.indent()
			if i+1 < length {
				f.write(indentStr) // one extra indent for pair or comment
			}
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

func (f *formatting) writeComment(c string) {
	if c == "" {
		return
	}
	f.writes(" ", strings.TrimSpace(c))
}

func (f *formatting) writeDecl(n *Var) {
	f.format(n)
	f.write(":")
	f.formatType(n.Type())
}

const indentStr = "    "

func (f *formatting) indent() {
	for i := 0; i < f.indentLevel; i++ {
		f.write(indentStr)
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
	empty, ok := n.(*EmptyStmt)
	if ok && empty.comment == "" {
		if !lastEmpty {
			f.writeLn()
		}
		return true
	}
	return false
}

func (f *formatting) writeWSS(n *BinaryExpression) {
	if !n.wss {
		f.write(" ")
	}
}
