package vm

import (
	"fmt"
	"math"

	"evylang.dev/evy/pkg/code"
	"evylang.dev/evy/pkg/compiler"
	"evylang.dev/evy/pkg/object"
)

const (
	StackSize = 2048

	GlobalsSize = 65536 // FIXME: limited to 65536 globals, re-eval // TODO: change fixme to todo
)

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
)

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpPop:
			vm.pop()
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpGreaterThanEqual, code.OpLessThan, code.OpLessThanEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSubtract, code.OpMultiply, code.OpDivide, code.OpModulo:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			operand := vm.pop()
			switch operand {
			case True:
				err := vm.push(False)
				if err != nil {
					return err
				}
			case False:
				err := vm.push(True)
				if err != nil {
					return err
				}
			}
		case code.OpMinus:
			obj := vm.pop()
			num, ok := obj.(*object.Integer)
			if !ok {
				return fmt.Errorf("unsupported type for negation: %s", obj.Type())
			}
			vm.push(&object.Integer{Value: -num.Value})
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
		// TODO: binary string
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	result := False
	switch op {
	case code.OpEqual:
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
			if left.(*object.Integer).Value == right.(*object.Integer).Value {
				result = True
			}
			return vm.push(result)
		}
		if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
			if left.(*object.Boolean).Value == right.(*object.Boolean).Value {
				result = True
			}
			return vm.push(result)
		}
	case code.OpNotEqual:
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
			if left.(*object.Integer).Value != right.(*object.Integer).Value {
				result = True
			}
			return vm.push(result)
		}
		if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
			if left.(*object.Boolean).Value != right.(*object.Boolean).Value {
				result = True
			}
			return vm.push(result)
		}
	case code.OpGreaterThan:
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
			if left.(*object.Integer).Value > right.(*object.Integer).Value {
				result = True
			}
			return vm.push(result)
		}
	case code.OpGreaterThanEqual:
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
			if left.(*object.Integer).Value >= right.(*object.Integer).Value {
				result = True
			}
			return vm.push(result)
		}
	case code.OpLessThan:
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
			if left.(*object.Integer).Value < right.(*object.Integer).Value {
				result = True
			}
			return vm.push(result)
		}
	case code.OpLessThanEqual:
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
			if left.(*object.Integer).Value <= right.(*object.Integer).Value {
				result = True
			}
			return vm.push(result)
		}
	default:
		return fmt.Errorf("unknown comparison operator: %d", op)
	}
	panic(fmt.Sprintf("invalid comparison between %s and %s", left.Inspect(), right.Inspect()))
}

func (vm *VM) executeBinaryIntegerOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result float64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSubtract:
		result = leftValue - rightValue
	case code.OpMultiply:
		result = leftValue * rightValue
	case code.OpDivide:
		result = leftValue / rightValue
	case code.OpModulo:
		result = math.Mod(leftValue, rightValue)
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}