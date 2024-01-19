package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var definitions = map[Opcode]*Definition{
	OpConstant:         {"OpConstant", []int{2}},
	OpAdd:              {"OpAdd", []int{}},
	OpSubtract:         {"OpSubtract", []int{}},
	OpMultiply:         {"OpMultiply", []int{}},
	OpModulo:           {"OpModulo", []int{}},
	OpDivide:           {"OpDivide", []int{}},
	OpPop:              {"OpPop", []int{}},
	OpGetGlobal:        {"OpGetGlobal", []int{2}},
	OpSetGlobal:        {"OpSetGlobal", []int{2}},
	OpTrue:             {"OpTrue", []int{}},
	OpFalse:            {"OpFalse", []int{}},
	OpEqual:            {"OpEqual", []int{}},
	OpNotEqual:         {"OpNotEqual", []int{}},
	OpLessThan:         {"OpLessThan", []int{}},
	OpLessThanEqual:    {"OpLessThanEqual", []int{}},
	OpGreaterThan:      {"OpGreaterThan", []int{}},
	OpGreaterThanEqual: {"OpGreaterThanEqual", []int{}},
	OpMinus:            {"OpMinus", []int{}},
	OpBang:             {"OpBang", []int{}},
}

const (
	OpConstant Opcode = iota

	OpAdd
	OpSubtract
	OpMultiply
	OpModulo
	OpDivide

	OpPop

	OpGetGlobal
	OpSetGlobal

	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanEqual
	OpLessThan
	OpLessThanEqual
	OpMinus
	OpBang
)

type Instructions []byte

type Opcode byte

type Definition struct {
	Name          string
	OperandWidths []int
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// TODO: Make panics when you specify a number of operands that
// is different from what the definiton expect
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
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
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
