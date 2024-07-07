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
	// MaxFrames is the total call depth that an evy program can go to.
	MaxFrames = 1024
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
	constants   []value
	frames      []*frame
	framesIndex int
	globals     []value
	stack       []value
	// sp is the stack pointer and always points to
	// the next value in the stack. The top of the stack is stack[sp-1].
	sp int
}

// NewVM returns a new VM.
func NewVM(bytecode *Bytecode) *VM {
	mainFn := funcVal{Instructions: bytecode.Instructions}
	mainFrame := newFrame(mainFn, 0)
	frames := make([]*frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants:   bytecode.Constants,
		globals:     make([]value, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
		stack:       make([]value, StackSize),
		sp:          0,
	}
}

// Run executes the provided bytecode instructions in order, any error
// will stop the execution.
//
//nolint:maintidx,gocognit // Run is special in that it is written to be optimal
func (vm *VM) Run() error {
	var err error
	for vm.currentFrame().ip < len(vm.currentFrame().instructions())-1 {
		vm.currentFrame().ip++
		ip := vm.currentFrame().ip
		ins := vm.currentFrame().instructions()
		// This loop is the hot path of the vm, avoid unnecessary
		// lookups or memory movement.
		switch Opcode(ins[ip]) {
		case OpConstant:
			constIndex := ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.constants[constIndex])
		case OpGetGlobal:
			globalIndex := ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.globals[globalIndex])
		case OpGetLocal:
			idx := ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			frame := vm.currentFrame()

			err = vm.push(vm.stack[frame.base+int(idx)])
		case OpSetGlobal:
			globalIndex := ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()
		case OpDrop:
			n := ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.drop(int(n))
		case OpSetLocal:
			idx := ReadUint16(ins[ip+1:])
			frame := vm.currentFrame()
			vm.currentFrame().ip += 2
			vm.stack[frame.base+int(idx)] = vm.pop()
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
			arrLen := int(ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

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
			mapLen := int(ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			m := mapVal{
				order: make([]stringVal, mapLen),
				m:     make(map[stringVal]value, mapLen),
			}
			for i := mapLen - 1; i >= 0; i-- {
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
				key := index.(stringVal)
				left.m[key] = val
				left.order = append(left.order, key)
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
			pos := int(ReadUint16(ins[ip+1:]))
			// jump the instruction pointer to pos - 1 because the run
			// loop will increment it on the next iteration
			vm.currentFrame().ip = pos - 1
		case OpJumpOnFalse:
			condition := vm.popBoolVal()
			if !condition {
				pos := int(ReadUint16(ins[ip+1:]))
				// jump the instruction pointer to pos - 1 because
				// the run loop will increment it on the next iteration
				vm.currentFrame().ip = pos - 1
			} else {
				// no jump, so advance the instruction pointer over the
				// unread operand
				vm.currentFrame().ip += 2
			}
		case OpStepRange:
			// OpStepRange works by popping the step range state from the stack,
			// pushing it back with an updated index, and pushing a bool that
			// says whether the range has completed or not. The operand says
			// whether there is a loop var or not, and if so, before pushing
			// the bool, we push the index so the bytecode can assign the value
			// to the loop var.
			hasLoopVar := int(ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			index := vm.popNumVal()
			step := vm.popNumVal()
			stop := vm.popNumVal()
			// stack overflow wont happen because we just popped these values
			_ = vm.push(stop)
			_ = vm.push(step)
			_ = vm.push(index + step)
			stillGoing := (step > 0 && index < stop) || (step < 0 && index > stop)
			if stillGoing && hasLoopVar != 0 {
				// Ignore stack overflow. Next push will error in that case
				_ = vm.push(index)
			}
			err = vm.push(boolVal(stillGoing))
		case OpIterRange:
			// OpIterRange works by popping the range state from the stack,
			// pushing it back with an updated index, and pushing a bool that
			// says whether the range has completed or not. The operand says
			// whether there is a loop var or not, and if so, before pushing
			// the bool, we push the value of the iterable at the index so
			// the bytecode can assign the value to the loop var.
			hasLoopVar := int(ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			index := int(vm.popNumVal())
			iter := vm.pop()
			// stack overflow wont happen because we just popped these values
			_ = vm.push(iter)
			_ = vm.push(numVal(index + 1))

			var val value
			if a, ok := iter.(arrayVal); ok && index < len(a.Elements) {
				val = a.Elements[index]
			} else if m, ok := iter.(mapVal); ok && index < len(m.order) {
				val = m.order[index]
			} else if s, ok := iter.(stringVal); ok && index < len(s) {
				val = s[index : index+1]
			}
			// val != nil means we're still going
			if val != nil && hasLoopVar != 0 {
				// Ignore stack overflow. Next push will error in that case
				_ = vm.push(val)
			}
			err = vm.push(boolVal(val != nil))
		case OpCall:
			numArgs := int(ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			fn := vm.peekFuncVal(numArgs)
			frame := newFrame(fn, vm.sp-numArgs)
			vm.pushFrame(frame)
			// create a hole on the stack for the local args of the fn
			// to be stored in
			vm.sp = frame.base + fn.NumLocals
		case OpReturn:
			retVal := vm.pop()
			frame := vm.popFrame()
			vm.sp = frame.base - 1
			err = vm.push(retVal)
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

func (vm *VM) peek(offset int) value {
	return vm.stack[vm.sp-1-offset]
}

// peekFuncVal peeks an element from the stack and casts it to an funcVal
// before returning the value. If elem is not a funcVal then it will error.
func (vm *VM) peekFuncVal(offset int) funcVal {
	elem := vm.peek(offset)
	val, ok := elem.(funcVal)
	if !ok {
		panic(fmt.Errorf("%w: expected to peek funcVal but got %s",
			ErrInternal, elem.Type()))
	}
	return val
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

func (vm *VM) drop(n int) {
	vm.sp -= n
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

func (vm *VM) currentFrame() *frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}
