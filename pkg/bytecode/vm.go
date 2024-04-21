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
	// ErrBadRepetition is returned when the right-hand side of the array
	// repetition operator is invalid; i.e. negative or not an integer.
	ErrBadRepetition = fmt.Errorf("%w: bad repetition count", ErrPanic)
)

// VM is responsible for executing evy programs from bytecode.
type VM struct {
	constants    []value
	globals      []value
	instructions Instructions
	iterStack    []iterator
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
		iterStack:    make([]iterator, 0),
		stack:        make([]value, StackSize),
		sp:           0,
	}
}

// Run executes the provided bytecode instructions in order, any error
// will stop the execution.
//
//nolint:maintidx,gocognit // Run is special in that it is written to be optimal
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
		case OpNot:
			val := vm.popBoolVal()
			err = vm.push(!val)
		case OpMinus:
			val := vm.popNumVal()
			err = vm.push(-val)
		case OpEqual:
			right := vm.pop()
			left := vm.pop()
			err = vm.push(boolVal(left.Equals(right)))
		case OpNotEqual:
			right := vm.pop()
			left := vm.pop()
			err = vm.push(boolVal(!left.Equals(right)))
		case OpNumLessThan:
			right, left := vm.popBinaryNums()
			err = vm.push(boolVal(left < right))
		case OpNumLessThanEqual:
			right, left := vm.popBinaryNums()
			err = vm.push(boolVal(left <= right))
		case OpNumGreaterThan:
			right, left := vm.popBinaryNums()
			err = vm.push(boolVal(left > right))
		case OpNumGreaterThanEqual:
			right, left := vm.popBinaryNums()
			err = vm.push(boolVal(left >= right))
		case OpStringLessThan:
			right, left := vm.popBinaryStrings()
			err = vm.push(boolVal(left < right))
		case OpStringLessThanEqual:
			right, left := vm.popBinaryStrings()
			err = vm.push(boolVal(left <= right))
		case OpStringGreaterThan:
			right, left := vm.popBinaryStrings()
			err = vm.push(boolVal(left > right))
		case OpStringGreaterThanEqual:
			right, left := vm.popBinaryStrings()
			err = vm.push(boolVal(left >= right))
		case OpStringConcatenate:
			right, left := vm.popBinaryStrings()
			err = vm.push(stringVal(left + right))
		case OpArray:
			arrLen := int(ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			arr := arrayVal{Elements: make([]value, arrLen)}
			// fill the array in reverse because elements were
			// added onto the stack LIFO
			for i := arrLen - 1; i >= 0; i-- {
				arr.Elements[i] = vm.pop()
			}
			err = vm.push(arr)
		case OpArrayConcatenate:
			right := vm.popArrayVal()
			left := vm.popArrayVal()
			concatenated := arrayVal{Elements: []value{}}
			concatenated.Elements = append(concatenated.Elements, left.Elements...)
			concatenated.Elements = append(concatenated.Elements, right.Elements...)
			err = vm.push(concatenated)
		case OpArrayRepeat:
			right := vm.popNumVal()
			left := vm.popArrayVal()
			repetitions := int(right)
			if float64(repetitions) != float64(right) {
				return fmt.Errorf("%w: not an integer: %s", ErrBadRepetition, right)
			}
			if repetitions < 0 {
				return fmt.Errorf("%w: negative count: %s", ErrBadRepetition, right)
			}
			elements := make([]value, 0, len(left.Elements)*repetitions)
			for range repetitions {
				elements = append(elements, left.Elements...)
			}
			err = vm.push(arrayVal{Elements: elements})
		case OpMap:
			mapLen := int(ReadUint16(vm.instructions[ip+1:])) / 2
			ip += 2
			m := mapVal{
				order: make([]stringVal, mapLen),
				m:     make(map[stringVal]value, mapLen),
			}
			for i := mapLen - 1; 0 <= i; i-- {
				val := vm.pop()
				key := vm.popStringVal()
				m.m[key] = val
				m.order[i] = key
			}
			err = vm.push(m)
		case OpIndex:
			index := vm.pop()
			left := vm.pop()
			indexed := left.(indexable)
			var val value
			val, err = indexed.Index(index)
			if err != nil {
				return err
			}
			err = vm.push(val)
		case OpSetIndex:
			index := vm.pop()
			left := vm.pop()
			val := vm.pop()
			switch left := left.(type) {
			case mapVal:
				err = left.Set(index, val)
			case arrayVal:
				err = left.Set(index, val)
			}
		case OpSlice:
			end := vm.pop()
			start := vm.pop()
			left := vm.pop()
			sliced := left.(sliceable)
			var val value
			val, err = sliced.Slice(start, end)
			if err != nil {
				return err
			}
			err = vm.push(val)
		case OpNone:
			err = vm.push(noneVal{})
		case OpJump:
			pos := int(ReadUint16(vm.instructions[ip+1:]))
			// jump the instruction pointer to pos - 1 because the run
			// loop will increment it on the next iteration
			ip = pos - 1
		case OpJumpOnFalse:
			condition := vm.popBoolVal()
			if !condition {
				pos := int(ReadUint16(vm.instructions[ip+1:]))
				// jump the instruction pointer to pos - 1 because
				// the run loop will increment it on the next iteration
				ip = pos - 1
			} else {
				// no jump, so advance the instruction pointer over the
				// unread operand
				ip += 2
			}
		case OpStepRange:
			stop := vm.popNumVal()
			step := vm.popNumVal()
			curr := vm.popNumVal()
			if step > 0 {
				err = vm.push(boolVal(curr < stop))
			} else {
				err = vm.push(boolVal(curr > stop))
			}
			if err != nil {
				return err
			}
		case OpArrayRange:
			arr := vm.popArrayVal()
			curr := vm.popNumVal()
			err = vm.push(boolVal(int(curr) < len(arr.Elements)))
		case OpMapRange:
			m := vm.popMapVal()
			curr := vm.popNumVal()
			err = vm.push(boolVal(int(curr) < len(m.m)))
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

// popBinaryStrings pops the top two elements of the stack (the left
// and right sides of the binary expressions) as strings and returns both.
func (vm *VM) popBinaryStrings() (string, string) {
	// the right was compiled last, so is higher on the stack
	// than the left
	right := vm.popStringVal()
	left := vm.popStringVal()
	return string(right), string(left)
}

// popNumVal pops an element from the stack and casts it to a num
// before returning the value. If elem is not a num then it will error.
func (vm *VM) popNumVal() numVal {
	elem := vm.pop()
	val, ok := elem.(numVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to pop numVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
}

// popBoolVal pops an element from the stack and casts it to a bool
// before returning the value. If elem is not a bool then it will error.
func (vm *VM) popBoolVal() boolVal {
	elem := vm.pop()
	val, ok := elem.(boolVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to pop boolVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
}

// popStringVal pops an element from the stack and casts it to a string
// before returning the value. If elem is not a string then it will error.
func (vm *VM) popStringVal() stringVal {
	elem := vm.pop()
	val, ok := elem.(stringVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to pop stringVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
}

// popArrayVal pops an element from the stack and casts it to an arrayVal
// before returning the value. If elem is not an arrayVal then it will error.
func (vm *VM) popArrayVal() arrayVal {
	elem := vm.pop()
	val, ok := elem.(arrayVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to pop arrayVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
}

// popMapVal pops an element from the stack and casts it to a mapVal
// before returning the value. If elem is not an mapVal then it will error.
func (vm *VM) popMapVal() mapVal {
	elem := vm.pop()
	val, ok := elem.(mapVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to pop arrayVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
}
