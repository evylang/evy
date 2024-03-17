package bytecode

import (
	"fmt"

	"evylang.dev/evy/pkg/parser"
)

var (
	// ErrUndefinedVar is returned when a variable name cannot
	// be resolved in the symbol table.
	ErrUndefinedVar = fmt.Errorf("%w: undefined variable", ErrPanic)
	// ErrUnknownOperator is returned when an operator cannot
	// be resolved.
	ErrUnknownOperator = fmt.Errorf("%w: unknown operator", ErrInternal)
)

// Compiler is responsible for turning a parsed evy program into
// bytecode.
type Compiler struct {
	constants    []value
	instructions Instructions
	globals      *SymbolTable
}

// Bytecode represents raw evy bytecode.
type Bytecode struct {
	Constants    []value
	Instructions Instructions
}

// NewCompiler returns a new compiler.
func NewCompiler() *Compiler {
	return &Compiler{globals: NewSymbolTable()}
}

// Compile accepts an AST node and renders it to bytecode internally.
func (c *Compiler) Compile(node parser.Node) error {
	switch node := node.(type) {
	case *parser.Program:
		return c.compileProgram(node)
	case *parser.InferredDeclStmt:
		return c.compileDecl(node.Decl)
	case *parser.AssignmentStmt:
		return c.compileAssignment(node)
	case *parser.BinaryExpression:
		return c.compileBinaryExpression(node)
	case *parser.UnaryExpression:
		return c.compileUnaryExpression(node)
	case *parser.GroupExpression:
		return c.Compile(node.Expr)
	case *parser.Var:
		return c.compileVar(node)
	case *parser.NumLiteral:
		num := numVal(node.Value)
		if err := c.emit(OpConstant, c.addConstant(num)); err != nil {
			return err
		}
	case *parser.BoolLiteral:
		opcode := OpFalse
		if node.Value {
			opcode = OpTrue
		}
		if err := c.emit(opcode); err != nil {
			return err
		}
	}
	return nil
}

// Bytecode renders the compiler instructions into Bytecode.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

// addConstant appends the provided value to the constants
// and returns the index of that constant.
func (c *Compiler) addConstant(obj value) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// addInstruction appends bytes to the instruction set and returns the
// position of the instruction.
func (c *Compiler) addInstruction(ins []byte) {
	c.instructions = append(c.instructions, ins...)
}

// emit makes and writes an instruction to the bytecode and returns the
// position of the instruction.
func (c *Compiler) emit(op Opcode, operands ...int) error {
	ins, err := Make(op, operands...)
	if err != nil {
		return err
	}
	c.addInstruction(ins)
	return nil
}

func (c *Compiler) compileBinaryExpression(expr *parser.BinaryExpression) error {
	if err := c.Compile(expr.Left); err != nil {
		return err
	}
	if err := c.Compile(expr.Right); err != nil {
		return err
	}
	switch expr.Op {
	case parser.OP_PLUS:
		return c.emit(OpAdd)
	case parser.OP_MINUS:
		return c.emit(OpSubtract)
	case parser.OP_ASTERISK:
		return c.emit(OpMultiply)
	case parser.OP_SLASH:
		return c.emit(OpDivide)
	case parser.OP_PERCENT:
		return c.emit(OpModulo)
	case parser.OP_EQ:
		return c.emit(OpEqual)
	case parser.OP_NOT_EQ:
		return c.emit(OpNotEqual)
	default:
		return fmt.Errorf("%w %s", ErrUnknownOperator, expr.Op)
	}
}

func (c *Compiler) compileUnaryExpression(expr *parser.UnaryExpression) error {
	if err := c.Compile(expr.Right); err != nil {
		return err
	}
	switch expr.Op {
	case parser.OP_MINUS:
		return c.emit(OpMinus)
	case parser.OP_BANG:
		return c.emit(OpNot)
	}
	return nil
}

func (c *Compiler) compileProgram(prog *parser.Program) error {
	for _, s := range prog.Statements {
		if err := c.Compile(s); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileDecl(decl *parser.Decl) error {
	if err := c.Compile(decl.Value); err != nil {
		return err
	}
	symbol := c.globals.Define(decl.Var.Name)
	return c.emit(OpSetGlobal, symbol.Index)
}

func (c *Compiler) compileAssignment(stmt *parser.AssignmentStmt) error {
	if err := c.Compile(stmt.Target); err != nil {
		return err
	}
	symbol := c.globals.Define(stmt.Value.String())
	return c.emit(OpSetGlobal, symbol.Index)
}

func (c *Compiler) compileVar(variable *parser.Var) error {
	symbol, ok := c.globals.Resolve(variable.Name)
	if !ok {
		return fmt.Errorf("%w %s", ErrUndefinedVar, variable.Name)
	}
	return c.emit(OpGetGlobal, symbol.Index)
}
