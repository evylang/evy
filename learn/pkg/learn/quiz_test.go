package learn

import (
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestQuiz1(t *testing.T) {
	quiz, err := NewQuizModel("testdata/course1/unit1/quiz1.md")
	assert.NoError(t, err)
	questions := GenerateQuestionSet(quiz.QuestionsByDifficulty, quiz.Frontmatter.Composition)
	assert.Equal(t, 5, len(questions))
	questionSet := map[string]bool{}
	for _, q := range questions {
		questionSet[q.Filename] = true
		got := strings.Contains(q.Filename, "exercise1") || strings.Contains(q.Filename, "shape")
		assert.Equal(t, true, got, "quiz1 should only contain exercise1 or shape questions", q.Filename)
	}
	assert.Equal(t, 5, len(questionSet))
}

func TestQuiz2(t *testing.T) {
	quiz, err := NewQuizModel("testdata/course1/unit1/quiz2.md")
	assert.NoError(t, err)
	questions := GenerateQuestionSet(quiz.QuestionsByDifficulty, quiz.Frontmatter.Composition)
	assert.Equal(t, 6, len(questions))
	questionSet := map[string]bool{}
	for _, q := range questions {
		questionSet[q.Filename] = true
		got := strings.Contains(q.Filename, "text") || strings.Contains(q.Filename, "cls")
		assert.Equal(t, true, got, "quiz2 should only contain text or cls questions", q.Filename)
	}
	assert.Equal(t, 6, len(questionSet))
}
