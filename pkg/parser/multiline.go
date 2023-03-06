package parser

import "strings"

// multilineItem is used to represent multiline array and map literals
// as used in formatting. It does not explicitly store array or map
// values, just placeholders - "el" for next element in arrays, keys for
// maps.
//
// A multilineItem can be:
// 1. Empty `""` representing a newline.
// 2. Starting with `"//"` represents a full comment.
// 3a. For arrays the multiline item `"el"` represents the next array
// literal element.
// 3b. For maps all the multiline items that don't represent
// newline (`""`) or comments (`// …`) are the key of the next map pair.
//
// The follow example shows a multiline array literal and its
// multilineItem slice representation:
//
//	arr := [ 1 // commment1
//	    2
//
//	    // comment2
//	    3
//	]
//
// has the representation:
//
//	[ "el", "// comment1", "el", "", "", "// comment2", "el", ""]
//
// The next example shows a multiline map literal and its
// multilineItem slice representation:
//
//	map := {
//	    a: 1 // commment1
//	    b: 2
//
//	    // comment2
//	    c: 3
//	}
//
// has the representation:
//
//	[ "", "a","// comment1", "b", "", "", "// comment2", "c", ""]
type multilineItem string

func (m multilineItem) isComment() bool {
	return strings.HasPrefix(string(m), "//")
}

func (m multilineItem) isNL() bool {
	return m == multilineNL
}

func (m multilineItem) isKey() bool {
	return !m.isNL() && !m.isComment()
}

const (
	multilineEl = multilineItem("el")
	multilineNL = multilineItem("\n")
)

func multilineComment(s string) multilineItem {
	s = strings.TrimSpace(s)
	return multilineItem(s + "\n")
}

func formatMultiline(multilineItems []multilineItem) []multilineItem {
	formatted := make([]multilineItem, 0, len(multilineItems))
	nlCount := 0
	for _, item := range multilineItems {
		switch {
		case item.isNL():
			nlCount++
		case item.isComment():
			nlCount = 1
		default: // array element or map key
			nlCount = 0
		}
		if nlCount <= 2 {
			formatted = append(formatted, item)
		}
	}
	return formatted
}
