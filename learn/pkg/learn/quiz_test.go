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
	quiz1Exercises := []string{"exercise1", "exercise-txtar", "exercise-parse-error"}
	for _, q := range questions {
		questionSet[q.Filename()] = true
		got := containsAny(q.Filename(), quiz1Exercises)
		assert.Equal(t, true, got, "unexpected exercises in quiz1 %q", q.Filename())
	}
	assert.Equal(t, 5, len(questionSet))
}

func TestQuiz2(t *testing.T) {
	quiz, err := NewQuizModel("testdata/course1/unit1/quiz2.md")
	assert.NoError(t, err)
	questions := GenerateQuestionSet(quiz.QuestionsByDifficulty, quiz.Frontmatter.Composition)
	assert.Equal(t, 6, len(questions))
	questionSet := map[string]bool{}
	quiz2Exercises := []string{"shape", "text", "cls"}
	for _, q := range questions {
		questionSet[q.Filename()] = true
		got := containsAny(q.Filename(), quiz2Exercises)
		assert.Equal(t, true, got, "unexpected exercises in quiz2 %q", q.Filename())
	}
	assert.Equal(t, 6, len(questionSet))
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
