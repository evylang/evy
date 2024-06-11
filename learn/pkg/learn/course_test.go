package learn

import (
	"bytes"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestNewCourseModel(t *testing.T) {
	course, err := NewCourseModel("testdata/course1/README.md")
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	err = course.printUnitBadgesHTML(buf)
	assert.NoError(t, err)
	got := buf.String()
	want := `
<h2>Unit 1: Demo Unit</h2>
<a href="exercise1/index.html">🔲</a>
<a href="shape/index.html">🔲</a>
<a href="quiz1.html">✨</a>
<a href="text/index.html">🔲</a>
<a href="cls/index.html">🔲</a>
<a href="quiz2.html">✨</a>
<a href="unittest.html">⭐️</a>
`[1:]
	assert.Equal(t, want, got)
}
