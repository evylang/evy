package bytecode

import (
	"errors"
	"fmt"

	"evylang.dev/evy/pkg/parser"
)

// ErrUndefinedVar is returned when a variable name cannot
// be resolved in the symbol table.
var ErrUndefinedVar = errors.New("undefined variable")

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
	return &Compiler{
		constants:    []value{},
		instructions: Instructions{},
		globals:      NewSymbolTable(),
	}
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
	case *parser.Var:
		return c.compileVar(node)
	case *parser.NumLiteral:
		num := &numVal{V: node.Value}
		if err := c.emit(OpConstant, c.addConstant(num)); err != nil {
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
