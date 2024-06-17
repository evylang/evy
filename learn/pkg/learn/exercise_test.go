package learn

import (
	"os"
	"slices"
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestNewExerciseModel(t *testing.T) {
	m, err := NewExerciseModel("testdata/course1/unit1/exercise1/README.md")
	assert.NoError(t, err)
	want := []DifficultyCount{
		{Difficulty: "easy", Count: 2},
		{Difficulty: "medium", Count: 2},
	}
	assert.Equal(t, want, m.Frontmatter.Composition)
}

func TestNewExerciseModelErrInvalidFrontmatter(t *testing.T) {
	_, err := NewExerciseModel("testdata/err-course/unit1/err-exercise1/err-exercise.md")
	assert.Error(t, ErrInvalidFrontmatter, err)
}

func TestNewExerciseModelErrExercise(t *testing.T) {
	_, err := NewExerciseModel("testdata/err-course/unit1/err-exercise2/err-exercise.md")
	assert.Error(t, ErrExercise, err)
}

func TestExerciseToHTML(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/README.md"
	m, err := NewExerciseModel(fname, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	got, err := m.ToHTML(false) // without marked
	assert.NoError(t, err)
	goldenFile := "testdata/golden/exercise/exercise1.html"
	b, err := os.ReadFile(goldenFile)
	assert.NoError(t, err)
	want := string(b)
	assert.Equal(t, want, got)
}

func TestExerciseToHTMLWithMarked(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/README.md"
	m, err := NewExerciseModel(fname, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	got, err := m.ToHTML(true) // with marked
	assert.NoError(t, err)

	goldenFile := "testdata/golden/exercise-with-marked/exercise1.html"
	b, err := os.ReadFile(goldenFile)
	assert.NoError(t, err)
	want := string(b)
	assert.Equal(t, want, got)
}

func TestQuestionsByDifficultySetForExercise(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/README.md"
	tests := []struct {
		name          string
		options       []Option
		wantEasyCount int
	}{
		{
			name:          "with-private",
			options:       []Option{WithPrivateKey(testKeyPrivate)},
			wantEasyCount: 5,
		},
		{
			name:          "ignore-sealed",
			options:       []Option{WithIgnoreSealed()},
			wantEasyCount: 4,
		},
		{
			name:          "empty-opts",
			options:       nil,
			wantEasyCount: 5,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m, err := NewExerciseModel(fname, test.options...)
			assert.NoError(t, err)
			questions := m.QuestionsByDifficulty
			assert.Equal(t, 2, len(questions))
			assert.Equal(t, test.wantEasyCount, len(questions["easy"]))
			assert.Equal(t, 5, len(questions["medium"]))
		})
	}
}

func TestGenerateQuestionSetForExercise(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/README.md"
	m, err := NewExerciseModel(fname, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	questions := GenerateQuestionSet(m.QuestionsByDifficulty, m.Frontmatter.Composition)
	assert.Equal(t, 4, len(questions))
	easy := m.QuestionsByDifficulty["easy"]
	for _, qid := range questions[:2] {
		assert.Equal(t, true, slices.Contains(easy, qid))
	}
	medium := m.QuestionsByDifficulty["medium"]
	for _, qid := range questions[2:] {
		assert.Equal(t, true, slices.Contains(medium, qid))
	}
}

func TestGenerateQuestionSetForTxtarExercise(t *testing.T) {
	fname := "testdata/course1/unit1/exercise-txtar/README.md"
	m, err := NewExerciseModel(fname)
	assert.NoError(t, err)

	questions := GenerateQuestionSet(m.QuestionsByDifficulty, m.Frontmatter.Composition)
	assert.Equal(t, 4, len(questions))
	// Ensure no accidental duplication of generated question variants
	qFilesNoHash := map[string]bool{}
	for _, q := range questions {
		fname := baseNoExt(q.Filename())
		// ignore generation hash
		idx := strings.LastIndex(fname, "-")
		fname = fname[:idx]
		qFilesNoHash[fname] = true
	}
	assert.Equal(t, 4, len(qFilesNoHash))
}
