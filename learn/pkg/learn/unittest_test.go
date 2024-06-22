package learn

import (
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestNewUnittestModel(t *testing.T) {
	unittest, err := NewUnittestModel("testdata/course1/unit1/unittest.md")
	assert.NoError(t, err)
	questions := GenerateQuestionSet(unittest.QuestionsByDifficulty, unittest.Frontmatter.Composition)
	assert.Equal(t, 13, len(questions))
	questionSet := map[string]bool{}
	for _, q := range questions {
		questionSet[q.Filename()] = true
		assert.Equal(t, false, strings.Contains(q.Filename(), "exercise1"), "unittest should ignore exercise1")
	}
	assert.Equal(t, 13, len(questionSet))
}
