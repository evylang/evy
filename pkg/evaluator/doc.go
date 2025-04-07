// Package evaluator executes a given Evy program by walking its AST.
//
// Initially, an Evy program is turned into a sequence of tokens by the
// [lexer]. Then, the [parser] generates an Abstract Syntax Tree(AST) from
// the tokens. Finally, the [Evaluator] of this package walks the AST and
// executes the program.
//
// The [Evaluator] is a tree-walking interpreter. It directly interprets
// the AST, without any preprocessing, intermediate representation, or
// compilation step. This is a straightforward way to implement an
// interpreter, but it trades off execution performance for simplicity.
//
// The [Evaluator] uses different [Platform] implementations to target
// different environments. For example, there is a JS/[WASM] platform for
// the browser, which has full support for all graphics built-in
// functions. There is also a command line environment, which does not
// have graphics functions support.
//
// The [NewEvaluator] function creates a new [Evaluator] for a given
// [Platform]. The evaluator can then be used by either:
//   - Passing an Evy program directly to the [Evaluator.Run] function.
//   - Passing the pre-generated AST to the [Evaluator.Eval] function.
//
// After the Evaluator has finished evaluating all top-level code and
// returned, the environment is responsible for starting an event loop if
// external events such as key press events or pointer/mouse move events
// are supported.
//
// In the event loop, new events are queued and handled sequentially by
// calling the [Evaluator.HandleEvent] function. Each event is passed
// to its event handler once. The event handler body runs to
// completion, and then the control is returned to the event loop. The
// event loop is terminated when the Evaluator is stopped. For a sample
// implementation of an event loop, see the main package in the
// pkg/wasm directory used in the browser environment.
//
// [WASM]: https://webassembly.org/
package evaluator
