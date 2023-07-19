package parser

import (
	"strings"
)

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

// accumulation classifies statements into "empty", "comment", "stmt"
// and "func". It is used for newline insertion before function and
// eventHandler declarations taking leading comments into account.
type accumulation struct {
	stmtType string // "empty", "comment", "stmt, "func"
	idx      int
}

// nlAfter returns a map (set) of statements that need to be
// followed by a newline.
func nlAfter(stmts []Node, comments map[Node]string) map[int]bool {
	accums := newAccumulations(stmts, comments)
	indices := map[int]bool{}
	length := len(accums)
	if length == 0 {
		return indices
	}
	for i, accum := range accums[:length-1] {
		switch {
		case accum.stmtType == "empty" || accum.stmtType == "comment":
			// do nothing for empty lines and comments
		case accum.stmtType == "func" && accums[i+1].stmtType == "stmt":
			// add NL after func decl directly followed by stmt
			indices[accum.idx] = true
		case accums[i+1].stmtType == "func":
			// add NL before func decl (before stmt or other func decl)
			beforeFuncIdx := accums[i+1].idx - 1
			indices[beforeFuncIdx] = true
		case i+2 < length && accums[i+1].stmtType == "comment" && accums[i+2].stmtType == "func":
			// add NL before comments of func decl (after stmt or other func decl)
			indices[accum.idx] = true
		}
	}
	return indices
}

func newAccumulations(stmts []Node, comments map[Node]string) []accumulation {
	lastStmtType := ""
	var accums []accumulation
	for i, stmt := range stmts {
		stmtType := "stmt"
		switch s := stmt.(type) {
		case *EmptyStmt:
			stmtType = "empty"
			if comments[s] != "" {
				stmtType = "comment"
			}
		case *FuncDeclStmt, *EventHandlerStmt:
			stmtType = "func"
		}
		if stmtType != lastStmtType || stmtType == "func" {
			accum := accumulation{stmtType: stmtType, idx: i}
			accums = append(accums, accum)
			lastStmtType = stmtType
		}
	}
	return accums
}
