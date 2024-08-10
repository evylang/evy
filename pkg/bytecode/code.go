package bytecode

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// OpConstant defines a constant that will be referred to by index
	// in the bytecode.
	OpConstant Opcode = iota
	// OpGetGlobal retrieves a symbol from the global symbol table at the
	// specified index.
	OpGetGlobal
	// OpSetGlobal adds a symbol to the specified index in the global
	// symbol table.
	OpSetGlobal
	// OpDrop pops and discards the top N elements of the stack.
	OpDrop
	// OpGetLocal retrieves a symbol from the current local-scoped
	// symbol table at the specified index. This index also corresponds
	// to its location on the VM stack.
	OpGetLocal
	// OpSetLocal adds a symbol to the specified index in the current
	// local-scoped symbol table. This index also corresponds to its
	// location on the VM stack.
	OpSetLocal
	// OpAdd instructs the virtual machine to perform an addition.
	OpAdd
	// OpSubtract instructs the virtual machine to perform a subtraction.
	OpSubtract
	// OpMultiply instructs the virtual machine to perform a multiplication.
	OpMultiply
	// OpDivide instructs the virtual machine to perform a division.
	OpDivide
	// OpModulo instructs the virtual machine to perform a modulo operation.
	// Modulo returns the remainder of dividing the left side of an expression
	// by the right.
	OpModulo
	// OpTrue represents the boolean literal true.
	OpTrue
	// OpFalse represents the boolean literal false.
	OpFalse
	// OpNot represents the not unary operator, which performs logical
	// negation on a boolean operand.
	OpNot
	// OpMinus represents the minus unary operator, which negates the
	// value of a numeric operand.
	OpMinus
	// OpEqual represents the equality operator.
	OpEqual
	// OpNotEqual represents the inequality operator.
	OpNotEqual
	// OpNumLessThan represents a less than operator for numeric
	// operands.
	OpNumLessThan
	// OpNumLessThanEqual represents a less than or equal operator for
	// numeric operands.
	OpNumLessThanEqual
	// OpNumGreaterThan represents a greater than operator for numeric
	// operands.
	OpNumGreaterThan
	// OpNumGreaterThanEqual represents a greater than or equal operator
	// for numeric operands.
	OpNumGreaterThanEqual
	// OpStringLessThan represents a less than operator for string
	// operands. Strings are compared using lexicographic order.
	OpStringLessThan
	// OpStringLessThanEqual represents a less than or equal operator for
	// string operands. Strings are compared using lexicographic order.
	OpStringLessThanEqual
	// OpStringGreaterThan represents a greater than operator for string
	// operands. Strings are compared using lexicographic order.
	OpStringGreaterThan
	// OpStringGreaterThanEqual represents a greater than or equal operator
	// for string operands. Strings are compared using lexicographic order.
	OpStringGreaterThanEqual
	// OpStringConcatenate represents a + operator used to concatenate two
	// strings.
	OpStringConcatenate
	// OpArray represents an array literal, the operand N is the length of
	// the array, and instructs the vm to pop N elements from the stack.
	OpArray
	// OpArrayConcatenate represents a + operator used to concatenate two
	// arrays.
	OpArrayConcatenate
	// OpArrayRepeat represents a * operator used to repeat an array a
	// number of times.
	OpArrayRepeat
	// OpMap represents a map literal, the operand N is the length of
	// the map. The keys and values are read sequentially from the
	// stack (e.g. k1, v1, k2, v2...).
	OpMap
	// OpIndex represents an index operator used on an array, map or
	// string variable.
	OpIndex
	// OpSetIndex represents an index operator on the left hand side of
	// an assignment.
	OpSetIndex
	// OpSlice represents a slice operator used on an array or string. The top
	// three elements are read from the stack and are the end index, start index
	// and the data structure being operated on. OpNone is used as a default
	// when the start or end are not specified by the user program.
	OpSlice
	// OpNone represents an unspecified value used where values are
	// optional and unspecified such as the start and end value of a
	// slice index, or the values in a step range.
	OpNone
	// OpJump will force the virtual machine to jump to the instruction
	// address within its operand.
	OpJump
	// OpJumpOnFalse will pop the top element of the stack as a
	// boolean and evaluate it. It will jump to the instruction address
	// in its operand if the condition evaluates to false.
	OpJumpOnFalse
	// OpStepRange represents a range over a numeric start, stop and step.
	// It has one operand that specifies if the loop range assigns to a
	// loop variable.
	OpStepRange
	// OpIterRange represents a range over an iterable structure (a string,
	// array or map). It has one operand that specifies if the loop range
	// assigns to a loop variable.
	OpIterRange
)

var (
	// ErrInternal and errors wrapping ErrInternal report internal
	// errors of the VM that should not occur during normal
	// program execution.
	ErrInternal = errors.New("internal error")
	// ErrPanic and errors wrapping ErrPanic report runtime errors, such
	// as an index out of bounds or a stack overflow.
	ErrPanic = errors.New("user error")
	// ErrUnknownOpcode is returned when an unknown opcode is encountered.
	ErrUnknownOpcode = fmt.Errorf("%w: unknown opcode", ErrInternal)
)

// definitions is a mapping of OpCode to OpDefinition.
var definitions = map[Opcode]*OpDefinition{
	// The definition for OpConstant says that its only operand is two
	// bytes wide, which makes it a uint16. This means that an evy
	// bytecode program can only have 65535 constants defined.
	OpConstant:  {"OpConstant", []int{2}},
	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetGlobal: {"OpSetGlobal", []int{2}},
	OpDrop:      {"OpDrop", []int{2}},
	OpGetLocal:  {"OpGetLocal", []int{2}},
	OpSetLocal:  {"OpSetLocal", []int{2}},
	// Operations like OpAdd have no operand width because the virtual
	// machine is expected to pop the values from the stack when reading
	// this instruction.
	OpAdd:                    {"OpAdd", nil},
	OpSubtract:               {"OpSubtract", nil},
	OpMultiply:               {"OpMultiply", nil},
	OpDivide:                 {"OpDivide", nil},
	OpModulo:                 {"OpModulo", nil},
	OpTrue:                   {"OpTrue", nil},
	OpFalse:                  {"OpFalse", nil},
	OpNot:                    {"OpNot", nil},
	OpMinus:                  {"OpMinus", nil},
	OpEqual:                  {"OpEqual", nil},
	OpNotEqual:               {"OpNotEqual", nil},
	OpNumLessThan:            {"OpNumLessThan", nil},
	OpNumLessThanEqual:       {"OpNumLessThanEqual", nil},
	OpNumGreaterThan:         {"OpNumGreaterThan", nil},
	OpNumGreaterThanEqual:    {"OpNumGreaterThanEqual", nil},
	OpStringLessThan:         {"OpStringLessThan", nil},
	OpStringLessThanEqual:    {"OpStringLessThanEqual", nil},
	OpStringGreaterThan:      {"OpStringGreaterThan", nil},
	OpStringGreaterThanEqual: {"OpStringGreaterThanEqual", nil},
	OpStringConcatenate:      {"OpStringConcatenate", nil},
	// This operand width only allows arrays up to 65535 elements in length.
	OpArray:            {"OpArray", []int{2}},
	OpArrayConcatenate: {"OpArrayConcatenate", nil},
	OpArrayRepeat:      {"OpArrayRepeat", nil},
	// This operand width only allows maps up to 65535 elements in length.
	OpMap:         {"OpMap", []int{2}},
	OpIndex:       {"OpIndex", nil},
	OpSetIndex:    {"OpSetIndex", nil},
	OpSlice:       {"OpSlice", nil},
	OpNone:        {"OpNone", nil},
	OpJump:        {"OpJump", []int{2}},
	OpJumpOnFalse: {"OpJumpOnFalse", []int{2}},
	OpStepRange:   {"OpStepRange", []int{2}}, // operand: hasLoopVar
	OpIterRange:   {"OpIterRange", []int{2}}, // operand: hasLoopVar
}

// OpDefinition defines a name and expected operand width for each OpCode.
type OpDefinition struct {
	Name          string
	OperandWidths []int
}

// Opcode defines the type of operation to be performed when reading
// the bytecode.
type Opcode byte

// Make assembles and returns a single bytecode instruction out
// of an Opcode and an optional list of operands. An error will
// be returned if there is no definition for the provided Opcode.
func Make(op Opcode, operands ...int) ([]byte, error) {
	def, err := Lookup(op)
	if err != nil {
		return nil, err
	}
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}
	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)
	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		if width == 2 {
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o)) //nolint:gosec // we are just going to be lax about overflow errors at the moment
		}
		offset += width
	}
	return instruction, nil
}

// ReadOperands will read from the provided Instructions based on the
// width of the provided OpDefinition. It returns the operands and the
// read offset inside the Instructions.
func ReadOperands(def *OpDefinition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		if width == 2 {
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

// ReadUint16 is exposed to allow reading without performing a Lookup to
// get an OpDefinition to pass to ReadOperands.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// Lookup returns an OpDefinition for the value of an Opcode, any
// unknown value will result in an error.
func Lookup(op Opcode) (*OpDefinition, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrUnknownOpcode, op)
	}
	return def, nil
}
