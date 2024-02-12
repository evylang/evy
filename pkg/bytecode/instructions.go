package bytecode

import (
	"bytes"
	"fmt"
)

// Instructions represents raw bytecode, which is composed of opcodes
// and an optional number of operands after each opcode.
type Instructions []byte

// String prints opcodes and their associated operands.
func (ins Instructions) String() string {
	var out bytes.Buffer
	for i := 0; i < len(ins); {
		def, err := Lookup(Opcode(ins[i]))
		if err != nil {
			panic(err.Error())
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func fmtInstruction(def *OpDefinition, operands []int) string {
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
	default:
		return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
	}
}
