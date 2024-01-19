package compiler

import (
	"fmt"

	"evylang.dev/evy/pkg/code"
	"evylang.dev/evy/pkg/object"
	"evylang.dev/evy/pkg/parser"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
	symbolTable  *SymbolTable
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		symbolTable:  NewSymbolTable(),
	}
}

func (c *Compiler) Compile(node parser.Node) error {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	// TODO: TypedDecStmt
	case *parser.TypedDeclStmt:

	case *parser.InferredDeclStmt:
		err := c.Compile(node.Decl.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Decl.Var.Name)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *parser.AssignmentStmt:
		if err := c.Compile(node.Target); err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Value.String())
		c.emit(code.OpSetGlobal, symbol.Index)
	case *parser.Var:
		symbol, ok := c.symbolTable.Resolve(node.Name)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Name)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
	case *parser.BinaryExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Op {
		case parser.OP_PLUS:
			c.emit(code.OpAdd)
		case parser.OP_MINUS:
			c.emit(code.OpSubtract)
		case parser.OP_ASTERISK:
			c.emit(code.OpMultiply)
		case parser.OP_SLASH:
			c.emit(code.OpDivide)
		case parser.OP_PERCENT:
			c.emit(code.OpModulo)
		}
	case *parser.NumLiteral:
		integer := &object.Integer{Value: int64(node.Value)}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *parser.GroupExpression:
		return c.Compile(node.Expr)
	default:
		return fmt.Errorf("unknown node type %s", node.Type())
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}
