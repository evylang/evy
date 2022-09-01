package evaluator

import "strings"

type BuiltinFunction func(args ...Object) Object

// Builtin implements the Object interface
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN }
func (b *Builtin) Inspect() string  { return "builtin function" }

func builtinFuncs(print func(string)) map[string]*Builtin {
	return map[string]*Builtin{
		"print": &Builtin{
			Fn: func(args ...Object) Object {
				argList := make([]string, len(args))
				for i, arg := range args {
					argList[i] = arg.Inspect()
				}
				print(strings.Join(argList, " "))
				return nil
			},
		},
	}
}
