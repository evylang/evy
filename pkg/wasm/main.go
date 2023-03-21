//go:build tinygo

package main

import (
	"foxygo.at/evy/pkg/evaluator"
	"foxygo.at/evy/pkg/parser"
)

var (
	version string
	eval    *evaluator.Evaluator
	events  []evaluator.Event
)

func main() {
	yielder := newSleepingYielder()
	builtins := evaluator.DefaultBuiltins(newJSRuntime(yielder))

	defer afterStop()
	source, err := format(builtins)
	if err != nil {
		builtins.Print(err.Error())
		return
	}
	evaluate(source, builtins, yielder)
}

func format(evalBuiltins evaluator.Builtins) (string, error) {
	input := getEvySource()

	builtins := evalBuiltins.ParserBuiltins()
	prog, err := parser.Parse(input, builtins)
	if err != nil {
		return "", parser.TruncateError(err, 8)
	}
	formattedInput := prog.Format()
	if formattedInput != input {
		setEvySource(formattedInput)
	}
	return formattedInput, nil
}

func evaluate(input string, builtins evaluator.Builtins, yielder *sleepingYielder) {
	eval = evaluator.NewEvaluator(builtins)
	eval.Yielder = yielder

	eval.Run(input)
	handleEvents(yielder)
}

func getEvySource() string {
	addr := evySource()
	return getStringFromAddr(addr)
}

func handleEvents(yielder *sleepingYielder) {
	if eval == nil || len(eval.EventHandlerNames()) == 0 {
		return
	}
	for _, name := range eval.EventHandlerNames() {
		registerEventHandler(name)
	}
	for {
		if eval.Stopped {
			return
		}
		// unsynchronized access to events - ok in WASM as single threaded.
		if len(events) > 0 {
			event := events[0]
			events = events[1:]
			yielder.Reset()
			eval.HandleEvent(event)
		} else {
			yielder.Sleep(minSleepDur)
		}
	}
}
