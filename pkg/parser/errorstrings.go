package parser

import "fmt"

func pluralize(n int, unit string) string {
	if n == 1 {
		return unit
	}
	return unit + "s"
}

func quantify(n int, unit string) string {
	return fmt.Sprintf("%d %s", n, pluralize(n, unit))
}

// ordinalize returns ordinal number as string. Invalid for negative
// numbers. E.g. `ordinalize(1) == "1st"`.
func ordinalize(n int) string {
	if 10 < n%100 && n%100 < 14 {
		return fmt.Sprintf("%dth", n)
	}
	m := map[int]string{0: "th", 1: "st", 2: "nd", 3: "rd", 4: "th", 5: "th", 6: "th", 7: "th", 8: "th", 9: "th"}
	return fmt.Sprintf("%d%s", n, m[n%10])
}
