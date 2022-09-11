package parser

import (
	"strconv"
)

// itoa is a shorthand for strconv.Itoa as it is used a lot when
// concatenating strings as tinygo does string formatting with
// fmt.Sprintf.
func itoa(n int) string {
	return strconv.Itoa(n)
}

func pluralize(n int, unit string) string {
	if n == 1 {
		return unit
	}
	return unit + "s"
}

func quantify(n int, unit string) string {
	return itoa(n) + " " + pluralize(n, unit)
}

// ordinalize returns ordinal number as string. Invalid for negative
// numbers. E.g. `ordinalize(1) == "1st"`.
func ordinalize(n int) string {
	if 10 < n%100 && n%100 < 14 {
		return itoa(n) + "th"
	}
	m := map[int]string{0: "th", 1: "st", 2: "nd", 3: "rd", 4: "th", 5: "th", 6: "th", 7: "th", 8: "th", 9: "th"}
	return itoa(n) + m[n%10]
}
