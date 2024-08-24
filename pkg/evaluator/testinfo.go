package evaluator

import "strconv"

// TestInfo contains flags for test runs, e.g. FailFast and test result
// information, e.g. total count.
type TestInfo struct {
	FailFast      bool
	NoTestSummary bool

	errors []error
	total  int
}

// FailCount returns the number of failed tests.
func (t *TestInfo) FailCount() int {
	return len(t.errors)
}

// SuccessCount returns the number of successful tests.
func (t *TestInfo) SuccessCount() int {
	return t.total - len(t.errors)
}

// TotalCount returns the total number of tests executed.
func (t *TestInfo) TotalCount() int {
	return t.total
}

// Report prints a summary of the test results.
func (t *TestInfo) Report(printFn func(string)) {
	if t.NoTestSummary || t.TotalCount() == 0 {
		return
	}
	succs := t.SuccessCount()
	fails := t.FailCount()
	if fails > 0 {
		printFn("❌ " + strconv.Itoa(fails) + " failed test" + suffix(fails) + "\n" +
			"✔️ " + strconv.Itoa(succs) + " passed test" + suffix(succs) + "\n")
	} else {
		printFn("✅ " + strconv.Itoa(succs) + " passed test" + suffix(succs) + "\n")
	}
}

func suffix(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
