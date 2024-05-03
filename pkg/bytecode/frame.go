package bytecode

// frame is a window over the instructions set.
type frame struct {
	// fn is the funcVal that contains the instructions to run
	// in this frame.
	fn funcVal
	// ip is the current location of the instruction pointer in
	// this frame.
	ip int
	// base is the stack pointer before this frame was executed, it
	// provides a relative location to create locals for this fn, and
	// a marker to return to after the frame is finished with an
	// OpReturn.
	base int
}

func newFrame(fn funcVal, base int) *frame {
	return &frame{fn: fn, ip: -1, base: base}
}

func (f *frame) instructions() Instructions {
	return f.fn.Instructions
}
