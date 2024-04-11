package bytecode

import (
	"fmt"

	"evylang.dev/evy/pkg/parser"
)

const (
	// JumpPlaceholder is used as a placeholder operand value in OpJump
	// and OpJumpOnFalse.
	JumpPlaceholder = 9999
)

var (
	// ErrUndefinedVar is returned when a variable name cannot
	// be resolved in the symbol table.
	ErrUndefinedVar = fmt.Errorf("%w: undefined variable", ErrPanic)
	// ErrUnknownOperator is returned when an operator cannot
	// be resolved.
	ErrUnknownOperator = fmt.Errorf("%w: unknown operator", ErrInternal)
	// ErrUnsupportedExpression is returned when an expression is not
	// supported by the compiler, this indicates an error in the compiler
	// itself, as all parseable evy expressions should be supported.
	ErrUnsupportedExpression = fmt.Errorf("%w: unsupported expression", ErrInternal)
)

// Compiler is responsible for turning a parsed evy program into
// bytecode.
type Compiler struct {
	constants    []value
	instructions Instructions
	globals      *SymbolTable
	// breaks tracks the positions of break statements in the inner-most loop.
	breaks []int
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
	case *parser.IndexExpression:
		return c.compileIndexExpression(node)
	case *parser.InferredDeclStmt:
		return c.compileDecl(node.Decl)
	case *parser.AssignmentStmt:
		return c.compileAssignment(node)
	case *parser.BinaryExpression:
		return c.compileBinaryExpression(node)
	case *parser.BreakStmt:
		return c.compileBreakStatement(node)
	case *parser.BlockStatement:
		return c.compileBlockStatement(node)
	case *parser.IfStmt:
		return c.compileIfStatement(node)
	case *parser.WhileStmt:
		return c.compileWhileStatement(node)
	case *parser.SliceExpression:
		return c.compileSliceExpression(node)
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
	case *parser.StringLiteral:
		num := stringVal(node.Value)
		if err := c.emit(OpConstant, c.addConstant(num)); err != nil {
			return err
		}
	case *parser.ArrayLiteral:
		for _, elem := range node.Elements {
			if err := c.Compile(elem); err != nil {
				return err
			}
		}
		if err := c.emit(OpArray, len(node.Elements)); err != nil {
			return err
		}
	case *parser.MapLiteral:
		for _, k := range node.Order {
			str := stringVal(k)
			if err := c.emit(OpConstant, c.addConstant(str)); err != nil {
				return err
			}
			if err := c.Compile(node.Pairs[k]); err != nil {
				return err
			}
		}
		if err := c.emit(OpMap, len(node.Pairs)*2); err != nil {
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
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

// emit makes and writes an instruction to the bytecode.
func (c *Compiler) emit(op Opcode, operands ...int) error {
	_, err := c.emitPos(op, operands...)
	return err
}

// emitPos makes and writes an instruction to the bytecode and returns the
// position of the instruction.
func (c *Compiler) emitPos(op Opcode, operands ...int) (int, error) {
	ins, err := Make(op, operands...)
	if err != nil {
		return 0, err
	}
	newPos := c.addInstruction(ins)
	return newPos, nil
}

func (c *Compiler) compileBinaryExpression(expr *parser.BinaryExpression) error {
	if err := c.Compile(expr.Left); err != nil {
		return err
	}
	if err := c.Compile(expr.Right); err != nil {
		return err
	}
	// equality and inequality are type agnostic in the vm, so no type checking
	// is required to decide which opcode to output.
	switch expr.Op {
	case parser.OP_EQ:
		return c.emit(OpEqual)
	case parser.OP_NOT_EQ:
		return c.emit(OpNotEqual)
	}
	if expr.Left.Type() == parser.NUM_TYPE && expr.Right.Type() == parser.NUM_TYPE {
		return c.compileNumBinaryExpression(expr)
	}
	if expr.Left.Type() == parser.STRING_TYPE && expr.Right.Type() == parser.STRING_TYPE {
		return c.compileStringBinaryExpression(expr)
	}
	if expr.Left.Type().Name == parser.ARRAY && expr.Right.Type().Name == parser.ARRAY && expr.Op == parser.OP_PLUS {
		return c.emit(OpArrayConcatenate)
	}
	if expr.Left.Type().Name == parser.ARRAY && expr.Right.Type() == parser.NUM_TYPE && expr.Op == parser.OP_ASTERISK {
		return c.emit(OpArrayRepeat)
	}
	return fmt.Errorf("%w: %s with types %s %s", ErrUnsupportedExpression,
		expr, expr.Left.Type(), expr.Right.Type())
}

func (c *Compiler) compileNumBinaryExpression(expr *parser.BinaryExpression) error {
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
	case parser.OP_LT:
		return c.emit(OpNumLessThan)
	case parser.OP_LTEQ:
		return c.emit(OpNumLessThanEqual)
	case parser.OP_GT:
		return c.emit(OpNumGreaterThan)
	case parser.OP_GTEQ:
		return c.emit(OpNumGreaterThanEqual)
	default:
		return fmt.Errorf("%w %s", ErrUnknownOperator, expr.Op)
	}
}

func (c *Compiler) compileStringBinaryExpression(expr *parser.BinaryExpression) error {
	switch expr.Op {
	case parser.OP_PLUS:
		return c.emit(OpStringConcatenate)
	case parser.OP_LT:
		return c.emit(OpStringLessThan)
	case parser.OP_LTEQ:
		return c.emit(OpStringLessThanEqual)
	case parser.OP_GT:
		return c.emit(OpStringGreaterThan)
	case parser.OP_GTEQ:
		return c.emit(OpStringGreaterThanEqual)
	default:
		return fmt.Errorf("%w %s", ErrUnknownOperator, expr.Op)
	}
}

func (c *Compiler) compileBlockStatement(block *parser.BlockStatement) error {
	for _, stmt := range block.Statements {
		if err := c.Compile(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileIfStatement(stmt *parser.IfStmt) error {
	firstJumpPos, err := c.compileConditionalBlock(stmt.IfBlock)
	if err != nil {
		return err
	}
	jumpPositions := []int{firstJumpPos}
	for _, elseif := range stmt.ElseIfBlocks {
		opJumpPos, err := c.compileConditionalBlock(elseif)
		if err != nil {
			return err
		}
		jumpPositions = append(jumpPositions, opJumpPos)
	}
	if stmt.Else != nil {
		if err := c.Compile(stmt.Else); err != nil {
			return err
		}
	}
	// rewrite all OpJump to jump to the end of the entire if statement,
	// optimisation: if the else block is empty then the last jump will
	// "jump" to the next instruction
	stmtEndPos := len(c.instructions)
	for _, jumpPos := range jumpPositions {
		c.instructions.changeOperand(jumpPos, stmtEndPos)
	}
	return nil
}

// compileConditionalBlock will compile the condition and block of a ConditionalBlock, emitting
// an OpJumpOnFalse after the condition and an OpJump after the block. The position of the
// OpJump is returned so that it can be rewritten in the event that this statement is part
// of a larger IfStmt.
func (c *Compiler) compileConditionalBlock(block *parser.ConditionalBlock) (int, error) {
	if err := c.Compile(block.Condition); err != nil {
		return 0, err
	}
	jumpOnFalsePos, err := c.emitPos(OpJumpOnFalse, JumpPlaceholder)
	if err != nil {
		return 0, err
	}
	if err := c.Compile(block.Block); err != nil {
		return 0, err
	}
	jumpPos, err := c.emitPos(OpJump, JumpPlaceholder)
	if err != nil {
		return 0, err
	}
	// rewrite the JumpPlaceholder in the OpJumpOnFalse so that it will jump to the end
	// of the statement when the condition is not truthy anymore
	afterBlockPos := len(c.instructions)
	c.instructions.changeOperand(jumpOnFalsePos, afterBlockPos)
	return jumpPos, nil
}

func (c *Compiler) compileWhileStatement(stmt *parser.WhileStmt) error {
	startPos := len(c.instructions)
	if err := c.Compile(stmt.Condition); err != nil {
		return err
	}
	// Prepare end position of while block, jump to end if condition is false
	jumpOnFalsePos, err := c.emitPos(OpJumpOnFalse, JumpPlaceholder)
	if err != nil {
		return err
	}
	// take a snapshot of the break list before compiling the body of the loop
	outOfScopeBreaks := c.breaks
	c.breaks = []int{}
	if err := c.Compile(stmt.Block); err != nil {
		return err
	}
	// Jump back to start of while condition
	if err := c.emit(OpJump, startPos); err != nil {
		return err
	}
	// rewrite the JumpPlaceholder in the OpJumpOnFalse so that it will
	// jump to the end of the statement when the condition is false
	afterBlockPos := len(c.instructions)
	c.instructions.changeOperand(jumpOnFalsePos, afterBlockPos)
	// rewrite the JumpPlaceholder in the break statements to jump
	// to the end of the loop
	for _, breakPos := range c.breaks {
		c.instructions.changeOperand(breakPos, afterBlockPos)
	}
	// reset the break list
	c.breaks = outOfScopeBreaks
	return nil
}

func (c *Compiler) compileBreakStatement(_ *parser.BreakStmt) error {
	// JumpPlaceholder will be rewritten by the parent loop
	pos, err := c.emitPos(OpJump, JumpPlaceholder)
	if err != nil {
		return err
	}
	c.breaks = append(c.breaks, pos)
	return nil
}

func (c *Compiler) compileSliceExpression(expr *parser.SliceExpression) error {
	var err error
	if err = c.Compile(expr.Left); err != nil {
		return err
	}
	if err = c.compileOrEmitNone(expr.Start); err != nil {
		return err
	}
	if err = c.compileOrEmitNone(expr.End); err != nil {
		return err
	}
	return c.emit(OpSlice)
}

// compilerOrEmitNone will emit OpNone if the provided parser node is
// nil. If the node is not nil then it will be compiled as normal.
func (c *Compiler) compileOrEmitNone(node parser.Node) error {
	if node != nil {
		return c.Compile(node)
	}
	return c.emit(OpNone)
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
	if err := c.Compile(stmt.Value); err != nil {
		return err
	}
	switch target := stmt.Target.(type) {
	case *parser.Var:
		symbol, ok := c.globals.Resolve(target.Name)
		if !ok {
			return fmt.Errorf("%w %s", ErrUndefinedVar, target.Name)
		}
		return c.emit(OpSetGlobal, symbol.Index)
	case *parser.IndexExpression:
		if err := c.Compile(target.Left); err != nil {
			return err
		}
		if err := c.Compile(target.Index); err != nil {
			return err
		}
		return c.emit(OpSetIndex)
	}
	return c.Compile(stmt.Target)
}

func (c *Compiler) compileVar(variable *parser.Var) error {
	symbol, ok := c.globals.Resolve(variable.Name)
	if !ok {
		return fmt.Errorf("%w %s", ErrUndefinedVar, variable.Name)
	}
	return c.emit(OpGetGlobal, symbol.Index)
}

func (c *Compiler) compileIndexExpression(expr *parser.IndexExpression) error {
	if err := c.Compile(expr.Left); err != nil {
		return err
	}
	if err := c.Compile(expr.Index); err != nil {
		return err
	}
	if err := c.emit(OpIndex); err != nil {
		return err
	}
	return nil
}
