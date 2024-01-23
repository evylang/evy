package bytecode

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestString(t *testing.T) {
	expected := "0000 OpConstant 0\n"
	var ins Instructions = mustMake(t, OpConstant, 0)
	assert.Equal(t, expected, ins.String())
}

func TestStringUnknownOpcode(t *testing.T) {
	_, exp := Lookup(0xff)
	defer func() {
		assert.Equal(t, exp.Error(), recover())
	}()
	var ins Instructions = []byte{0xff, 0x00, 0x00}
	_ = ins.String()
}
