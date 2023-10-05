//go:build tinygo

package main

import (
	"errors"
	"strings"

	"evylang.dev/evy/pkg/evaluator"
	"evylang.dev/evy/pkg/parser"
)

var (
	version string
	eval    *evaluator.Evaluator
	events  []evaluator.Event
)

func main() {
	defer afterStop()
	actions := getActions()

	input := getEvySource()
	ast, err := parse(input)
	if err != nil {
		jsError(err.Error())
		return
	}
	if actions["fmt"] {
		formattedInput := ast.Format()
		if formattedInput != input {
			setEvySource(formattedInput)
			ast, err = parse(formattedInput)
			if err != nil {
				jsError(err.Error())
				return
			}
		}
	}
	if actions["ui"] {
		prepareUI(ast)
	}
	if actions["eval"] {
		rt := newJSRuntime()
		err := evaluate(ast, rt)
		if err == nil || errors.Is(err, evaluator.ErrStopped) {
			return
		}
		var exitErr evaluator.ExitError
		if errors.As(err, &exitErr) && exitErr == 0 {
			return
		}
		jsError(err.Error())
	}
}

func getActions() map[string]bool {
	m := map[string]bool{}
	addr := jsActions()
	s := getStringFromAddr(addr)
	actions := strings.Split(s, ",")
	for _, action := range actions {
		if action != "" {
			m[action] = true
		}
	}
	return m
}

func getEvySource() string {
	addr := evySource()
	return getStringFromAddr(addr)
}

func parse(input string) (*parser.Program, error) {
	builtins := evaluator.BuiltinDecls()
	prog, err := parser.Parse(input, builtins)
	if err != nil {
		var parseErrors parser.Errors
		if errors.As(err, &parseErrors) {
			err = parseErrors.Truncate(8)
		}
		return nil, err
	}
	return prog, nil
}

func prepareUI(prog *parser.Program) {
	names := make([]string, 0, len(prog.CalledBuiltinFuncs)+len(prog.EventHandlers))
	names = append(names, prog.CalledBuiltinFuncs...)
	for name := range prog.EventHandlers {
		names = append(names, name)
	}
	jsPrepareUI(strings.Join(names, ","))
}

func evaluate(prog *parser.Program, rt *jsRuntime) error {
	eval = evaluator.NewEvaluator(rt)
	if err := eval.Eval(prog); err != nil {
		return err
	}
	return handleEvents(rt.yielder)
}

func handleEvents(yielder *sleepingYielder) error {
	if eval == nil || len(eval.EventHandlerNames) == 0 {
		return nil
	}
	for _, name := range eval.EventHandlerNames {
		registerEventHandler(name)
	}
	for {
		if eval.Stopped {
			return nil
		}
		// unsynchronized access to events - ok in WASM as single threaded.
		if len(events) > 0 {
			event := events[0]
			events = events[1:]
			yielder.Reset()
			if err := eval.HandleEvent(event); err != nil {
				return err
			}
		} else {
			yielder.ForceYield()
		}
	}
	return nil
}
