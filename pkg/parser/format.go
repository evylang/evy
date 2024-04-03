package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func newFormatting() *formatting {
	return &formatting{
		wss:       map[*BinaryExpression]bool{},
		comments:  map[Node]string{},
		multiline: map[Node][]multilineItem{},
	}
}

type formatting struct {
	w io.StringWriter

	wss       map[*BinaryExpression]bool
	comments  map[Node]string
	multiline map[Node][]multilineItem

	indentLevel int
}

func (f *formatting) recordComment(n Node, comment string) {
	f.comments[n] = comment
}

func (f *formatting) recordWSS(n *BinaryExpression) {
	f.wss[n] = true
}

func (f *formatting) recordMultiline(n Node, multiline []multilineItem) {
	f.multiline[n] = multiline
}

func (f *formatting) format(n Node) {
	switch n := n.(type) {
	case *Program:
		f.formatProgram(n)
	case *EmptyStmt:
		f.writeComment(n)
	case *Any:
		f.format(n.Value)
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
	case *FuncDefStmt:
		f.formatFuncDefStmt(n)
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
		f.write(n.Op.String())
		f.format(n.Right)
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
	case *TypeAssertion:
		f.format(n.Left)
		f.write(".(")
		f.formatType(n.T)
		f.write(")")
	case *GroupExpression:
		f.write("(")
		f.format(n.Expr)
		f.write(")")
	case *BoolLiteral:
		f.write(n.String())
	case *NumLiteral:
		f.write(n.String())
	case *StringLiteral:
		f.write(strconv.Quote(n.Value))
	case *ArrayLiteral:
		f.formatArrayLiteral(n)
	case *MapLiteral:
		f.formatMapLiteral(n)
	default:
		f.write(fmt.Sprintf("format unimplemented for %v", n))
	}
}

func (f *formatting) formatProgram(p *Program) {
	if len(p.Statements) == 0 {
		f.write("\n")
		return
	}
	nl := nlAfter(p.Statements, f.comments)

	empty := false
	for i, stmt := range p.Statements {
		if empty = f.writeBlankLine(stmt, empty); empty {
			continue
		}
		f.indent()
		f.format(stmt)
		f.writeLn()
		if nl[i] {
			f.writeLn() // write newline after func / on decl
		}
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

func (f *formatting) formatFuncDefStmt(s *FuncDefStmt) {
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
	multi := formatMultiline(f.multiline[n])
	if len(multi) == 0 {
		f.write("[]")
		return
	}
	f.write("[")
	f.indentLevel++
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

		if i+1 < length && !multi[i+1].isNL() { // next is element, comment or `]`
			f.indent()
		}
	}
	f.indentLevel--
	if multi[length-1].isNL() {
		f.indent()
	}
	f.write("]")
}

func (f *formatting) formatMapLiteral(n *MapLiteral) {
	multi := formatMultiline(f.multiline[n])
	if len(multi) == 0 {
		f.write("{}")
		return
	}
	f.write("{")
	f.indentLevel++
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
		if i+1 < length && !multi[i+1].isNL() { // next is pair, comment or `}`
			f.indent()
		}
	}
	f.indentLevel--
	if multi[length-1].isNL() {
		f.indent()
	}
	f.write("}")
}

func (f *formatting) formatType(t *Type) {
	f.write(t.Name.String())
	if t.Sub != nil && t != EMPTY_ARRAY && t != EMPTY_MAP {
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

const indentStr = "    "

func (f *formatting) indent() {
	for range f.indentLevel {
		f.write(indentStr)
	}
}

func (f *formatting) writeStmts(stmts []Node) {
	f.indentLevel++
	empty := false
	for _, stmt := range stmts {
		if empty = f.writeBlankLine(stmt, empty); empty {
			continue
		}
		f.indent()
		f.format(stmt)
		f.writeLn()
	}

	f.indentLevel--
}

func (f *formatting) writeBlankLine(n Node, lastEmpty bool) bool {
	_, ok := n.(*EmptyStmt)
	if !ok || f.comments[n] != "" {
		return false
	}
	if !lastEmpty {
		f.writeLn()
	}
	return true
}

func (f *formatting) writeWSS(n *BinaryExpression) {
	if !f.wss[n] {
		f.write(" ")
	}
}
