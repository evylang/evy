package bytecode

import (
	"fmt"
	"math"
)

const (
	// StackSize defines an upper limit for the size of the stack.
	StackSize = 2048
	// GlobalsSize is the total number of globals that can be specified
	// in an evy program.
	GlobalsSize = 65536
)

var (
	// ErrStackOverflow is returned when the stack exceeds its size limit.
	ErrStackOverflow = fmt.Errorf("%w: stack overflow", ErrPanic)
	// ErrDivideByZero is returned when a division by zero would
	// produce an invalid result. In Golang, floating point division
	// by zero produces +Inf, and modulo by zero produces NaN.
	ErrDivideByZero = fmt.Errorf("%w: division by zero", ErrPanic)
)

// VM is responsible for executing evy programs from bytecode.
type VM struct {
	constants    []value
	globals      []value
	instructions Instructions
	stack        []value
	// sp is the stack pointer and always points to
	// the next value in the stack. The top of the stack is stack[sp-1].
	sp int
}

// NewVM returns a new VM.
func NewVM(bytecode *Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		globals:      make([]value, GlobalsSize),
		instructions: bytecode.Instructions,
		stack:        make([]value, StackSize),
		sp:           0,
	}
}

// Run executes the provided bytecode instructions in order, any error
// will stop the execution.
func (vm *VM) Run() error {
	var err error
	for ip := 0; ip < len(vm.instructions); ip++ {
		// This loop is the hot path of the vm, avoid unnecessary
		// lookups or memory movement.
		op := Opcode(vm.instructions[ip])
		switch op {
		case OpConstant:
			constIndex := ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err = vm.push(vm.constants[constIndex])
		case OpGetGlobal:
			globalIndex := ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err = vm.push(vm.globals[globalIndex])
		case OpSetGlobal:
			globalIndex := ReadUint16(vm.instructions[ip+1:])
			ip += 2
			vm.globals[globalIndex] = vm.pop()
		case OpAdd:
			right, left := vm.popBinaryNums()
			err = vm.push(numVal(left + right))
		case OpSubtract:
			right, left := vm.popBinaryNums()
			err = vm.push(numVal(left - right))
		case OpMultiply:
			right, left := vm.popBinaryNums()
			err = vm.push(numVal(left * right))
		case OpDivide:
			right, left := vm.popBinaryNums()
			if right == 0 {
				return ErrDivideByZero
			}
			err = vm.push(numVal(left / right))
		case OpModulo:
			right, left := vm.popBinaryNums()
			if right == 0 {
				return ErrDivideByZero
			}
			// floating point modulo has to be handled using this math function
			err = vm.push(numVal(math.Mod(left, right)))
		case OpTrue:
			err = vm.push(boolVal(true))
		case OpFalse:
			err = vm.push(boolVal(false))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// lastPoppedStackElem returns the last element that was
// popped from the stack. It is used in testing to
// check that the state of the vm is correct.
func (vm *VM) lastPoppedStackElem() value {
	return vm.stack[vm.sp]
}

func (vm *VM) push(o value) error {
	if vm.sp >= StackSize {
		return ErrStackOverflow
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() value {
	// Ignore stack underflow errors as that indicates an error in the
	// vm and the out-of-bounds slice panic is sufficient for that,
	// as opposed to the stack overflow above which can occur due to
	// a user program that the vm is running.
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

// popBinaryNums pops the top two elements of the stack (the left
// and right sides of the binary expressions) as nums and returns both.
func (vm *VM) popBinaryNums() (float64, float64) {
	// the right was compiled last, so is higher on the stack
	// than the left
	right := vm.popNumVal()
	left := vm.popNumVal()
	return float64(right), float64(left)
}

// popNumVal pops an element from the stack and casts it to a num
// before returning the value. If elem is not a num it will error.
func (vm *VM) popNumVal() numVal {
	elem := vm.pop()
	val, ok := elem.(numVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to pop numVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
}
