package evaluator

import "strings"

func Run(input string, print func(string)) {
	print(strings.ToUpper(input))
}
