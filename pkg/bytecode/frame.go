package bytecode

type frame struct {
	fn funcVal
	ip int
}

func newFrame(fn funcVal) *frame {
	return &frame{fn: fn, ip: -1}
}

func (f *frame) Instructions() Instructions {
	return f.fn.Instructions
}
