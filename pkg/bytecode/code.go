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
	// OpGetGlobal retrieves a symbol from the symbol table at the
	// specified index.
	OpGetGlobal
	// OpSetGlobal adds a symbol to the specified index in the symbol
	// table.
	OpSetGlobal
)

// ErrUnknownOp is returned when an unknown opcode is encountered.
var ErrUnknownOp = errors.New("unknown opcode")

// definitions is a mapping of OpCode to OpDefinition.
var definitions = map[Opcode]*OpDefinition{
	// The definition for OpConstant says that its only operand is two
	// bytes wide, which makes it a uint16. This means that an evy
	// bytecode program can only have 65535 constants defined.
	OpConstant:  {"OpConstant", []int{2}},
	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetGlobal: {"OpSetGlobal", []int{2}},
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
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
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
		return nil, fmt.Errorf("%w: %d", ErrUnknownOp, op)
	}
	return def, nil
}
