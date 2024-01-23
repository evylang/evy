package bytecode

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}
	for _, tt := range tests {
		instruction, err := Make(tt.op, tt.operands...)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, instruction)
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		mustMake(t, OpConstant, 2),
		mustMake(t, OpConstant, 65535),
	}
	expected := "0000 OpConstant 2\n0003 OpConstant 65535\n"
	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	assert.Equal(t, expected, concatted.String(), "instructions wrongly formatted.")
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}
	for _, tt := range tests {
		instruction, err := Make(tt.op, tt.operands...)
		assert.NoError(t, err)
		def, err := Lookup(tt.op)
		assert.NoError(t, err)
		operandsRead, read := ReadOperands(def, instruction[1:])
		assert.Equal(t, read, tt.bytesRead, "wrong num bytes read")
		assert.Equal(t, tt.operands, operandsRead, "wrong operands")
	}
}

func mustMake(t *testing.T, op Opcode, operands ...int) []byte {
	t.Helper()
	ins, err := Make(op, operands...)
	assert.NoError(t, err)
	return ins
}
