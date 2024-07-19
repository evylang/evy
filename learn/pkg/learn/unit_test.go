package learn

import (
	"bytes"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestNewUnitModel(t *testing.T) {
	unit, err := NewUnitModel("testdata/course1/unit1/README.md")
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	unitDir := "testdata/course1/unit1"
	err = unit.printBadgesHTML(buf, unitDir)
	assert.NoError(t, err)
	got := buf.String()
	want := `
<a href="exercise1/index.html">🔲</a>
<a href="exercise-txtar/index.html">🔲</a>
<a href="exercise-parse-error/index.html">🔲</a>
<a href="exercise-text/index.html">🔲</a>
<a href="quiz1.html">✨</a>
<a href="shape/index.html">🔲</a>
<a href="text/index.html">🔲</a>
<a href="cls/index.html">🔲</a>
<a href="quiz2.html">✨</a>
<a href="unittest.html">⭐️</a>
`[1:]
	assert.Equal(t, want, got)
}
