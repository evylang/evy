package evaluator

import "strconv"

// AssertInfo contains flags for test runs, e.g. FailFast and testResult
// information, e.g. total count.
type AssertInfo struct {
	FailFast           bool
	NoAssertionSummary bool

	errors []error
	total  int
}

// FailCount returns the number of failed assertions.
func (a *AssertInfo) FailCount() int {
	return len(a.errors)
}

// SuccessCount returns the number of successful assertions.
func (a *AssertInfo) SuccessCount() int {
	return a.total - len(a.errors)
}

// TotalCount returns the total number of assertions executed.
func (a *AssertInfo) TotalCount() int {
	return a.total
}

// Report prints a summary of the test results.
func (a *AssertInfo) Report(printFn func(string)) {
	if a.NoAssertionSummary || a.TotalCount() == 0 {
		return
	}
	succs := a.SuccessCount()
	fails := a.FailCount()
	if fails > 0 {
		printFn("❌ " + strconv.Itoa(fails) + " failed assertion" + suffix(fails) + "\n" +
			"✔️ " + strconv.Itoa(succs) + " passed assertion" + suffix(succs) + "\n")
	} else {
		printFn("✅ " + strconv.Itoa(succs) + " passed assertion" + suffix(succs) + "\n")
	}
}

func suffix(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
