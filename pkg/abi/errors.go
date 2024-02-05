package abi

import (
	"errors"
	"fmt"
)

// The Evaluator can return the following sentinel errors:
//   - ErrStopped is returned when the program has been stopped externally.
//   - ErrPanic and errors wrapping ErrPanic report runtime errors, such as an index out of bounds error.
//   - ErrInternal and errors wrapping ErrInternal report internal errors of the evaluator or AST. These errors should not occur.
var (
	ErrStopped = errors.New("stopped")

	ErrPanic         = errors.New("panic")
	ErrBounds        = fmt.Errorf("%w: index out of bounds", ErrPanic)
	ErrRangevalue    = fmt.Errorf("%w: bad range value", ErrPanic)
	ErrMapKey        = fmt.Errorf("%w: no value for map key", ErrPanic)
	ErrSlice         = fmt.Errorf("%w: bad slice", ErrPanic)
	ErrBadArguments  = fmt.Errorf("%w: bad arguments", ErrPanic)
	ErrAnyConversion = fmt.Errorf("%w: error converting any to type", ErrPanic)
	ErrVarNotSet     = fmt.Errorf("%w: variable has not been set yet", ErrPanic)

	ErrInternal         = errors.New("internal error")
	ErrUnknownNode      = fmt.Errorf("%w: unknown AST node", ErrInternal)
	ErrType             = fmt.Errorf("%w: type error", ErrInternal)
	ErrRangeType        = fmt.Errorf("%w: bad range type", ErrInternal)
	ErrOperation        = fmt.Errorf("%w: unknown operation", ErrInternal)
	ErrAssignmentTarget = fmt.Errorf("%w: bad assignment target", ErrInternal)
)

// ExitError is returned by [Evaluator.Eval] if Evy's [Builtin exit]
// function is called.
//
// [Builtin exit]: https://github.com/evylang/evy/blob/main/docs/Builtins.md#exit
type ExitError int

// Error implements the error interface and returns message containing the exit status.
func (e ExitError) Error() string {
	return fmt.Sprintf("exit %d", int(e))
}

// PanicError is returned by [Evaluator.Eval] if Evy's [Builtin panic]
// function is called or a runtime error occurs.
//
// [Builtin panic]: https://github.com/evylang/evy/blob/main/docs/Builtins.md#panic
type PanicError string

// Error implements the error interface and returns the panic message.
func (e PanicError) Error() string {
	return string(e)
}

// Unwrap returns the ErrPanic sentinel error so that it can be used in
//
//	errors.Is(err, evaluator.ErrPanic)
func (e *PanicError) Unwrap() error {
	return ErrPanic
}
