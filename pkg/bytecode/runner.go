// copied all of the bits from Evaluator that didn't fit into VM
// and added them here as an abstraction layer - yielding doesn't work atm
// and no events are handled
package bytecode

import (
	"errors"

	"evylang.dev/evy/pkg/evaluator"
	"evylang.dev/evy/pkg/parser"
)

var (
	ErrStopped = errors.New("stopped")
)

type Runner struct {
	Stopped           bool
	EventHandlerNames []string

	yielder evaluator.Yielder // Yield to give JavaScript/browser events a chance to run.
}

func NewRunner(rt evaluator.Runtime) *Runner {
	return &Runner{
		yielder: rt.Yielder(),
	}
}

func (rt *Runner) Run(prog *parser.Program) error {
	compiler := NewCompiler()
	if err := compiler.Compile(prog); err != nil {
		return err
	}
	rt.yield()
	return NewVM(compiler.Bytecode()).Run()
}

func (r *Runner) HandleEvent(ev evaluator.Event) error {
	return nil
}

func (rt *Runner) yield() {
	if rt.yielder != nil {
		rt.yielder.Yield()
	}
}
