package bytecode

// Platform must be implemented by an environment in order
// to execute Evy source code and Evy builtins. See evaluator.Platform
// for more.
type Platform interface {
	Print(string)
}
