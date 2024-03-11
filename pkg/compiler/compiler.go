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
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Target.String())
		c.emit(code.OpSetGlobal, symbol.Index)
	case *parser.Var:
		symbol, ok := c.symbolTable.Resolve(node.Name)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Name)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
	case *parser.IndexExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		c.emit(code.OpIndex)
	case *parser.SliceExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if node.Start != nil {
			if err := c.Compile(node.Start); err != nil {
				return err
			}
		} else {
			c.emit(code.OpNull)
		}
		if node.End != nil {
			if err := c.Compile(node.End); err != nil {
				return err
			}
		} else {
			c.emit(code.OpNull)
		}
		c.emit(code.OpSlice)
	case *parser.BinaryExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Right); err != nil {
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
		case parser.OP_LT:
			c.emit(code.OpLessThan)
		case parser.OP_LTEQ:
			c.emit(code.OpLessThanEqual)
		case parser.OP_GT:
			c.emit(code.OpGreaterThan)
		case parser.OP_GTEQ:
			c.emit(code.OpGreaterThanEqual)
		case parser.OP_EQ:
			c.emit(code.OpEqual)
		case parser.OP_NOT_EQ:
			c.emit(code.OpNotEqual)
		}
	case *parser.UnaryExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Op {
		case parser.OP_MINUS:
			c.emit(code.OpMinus)
		case parser.OP_BANG:
			c.emit(code.OpBang)
		}
	case *parser.Any:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
	case *parser.NumLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *parser.BoolLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *parser.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *parser.GroupExpression:
		return c.Compile(node.Expr)
	case *parser.IfStmt:
		if err := c.Compile(node.IfBlock.Condition); err != nil {
			return err
		}

		// emit 9999 as a placeholder value, this will be backfilled
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		// compile node consequence
		if err := c.Compile(node.IfBlock.Block); err != nil {
			return err
		}

		jumpPositions := []int{c.emit(code.OpJump, 9999)}

		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		// compile all conditional alternatives
		for _, elseif := range node.ElseIfBlocks {
			if err := c.Compile(elseif.Condition); err != nil {
				return err
			}

			// emit 9999 as a placeholder value, this will be backfilled
			jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

			if err := c.Compile(elseif.Block); err != nil {
				return err
			}

			afterConsequencePos := len(c.instructions)
			c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

			jumpPositions = append(jumpPositions, c.emit(code.OpJump, 9999))
		}

		// compile node alternative
		if node.Else != nil {
			if err := c.Compile(node.Else); err != nil {
				return err
			}
		}

		afterAlternativePos := len(c.instructions)
		for _, jumpPos := range jumpPositions {
			c.changeOperand(jumpPos, afterAlternativePos)
		}
	case *parser.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *parser.ArrayLiteral:
		for _, elem := range node.Elements {
			if err := c.Compile(elem); err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *parser.MapLiteral:
		for _, k := range node.Order {
			str := &object.String{Value: k}
			c.emit(code.OpConstant, c.addConstant(str))

			if err := c.Compile(node.Pairs[k]); err != nil {
				return err
			}
		}
		c.emit(code.OpMap, len(node.Pairs)*2)
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

func (c *Compiler) changeOperand(opPosition int, operand int) {
	op := code.Opcode(c.instructions[opPosition])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPosition, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}
