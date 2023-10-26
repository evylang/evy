package parser

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestInfer(t *testing.T) {
	// arr := [[]]
	arr := &Type{Name: ARRAY, Sub: EMPTY_ARRAY}
	got := arr.infer()
	want := &Type{
		Name: ARRAY,
		Sub:  &Type{Name: ARRAY, Sub: ANY_TYPE},
	}
	assert.Equal(t, want, got)
}

func TestTypeEquals(t *testing.T) {
	t1 := &Type{
		Name: ARRAY,
		Sub:  ANY_TYPE,
	}
	t2 := &Type{
		Name: ARRAY,
		Sub: &Type{
			Name: ANY,
			Sub:  nil,
		},
	}
	assert.Equal(t, true, t1.Equals(t2))
	assert.Equal(t, true, t2.Equals(t1))
}
