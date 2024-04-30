package answer

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

var testQuestions = map[string]Answer{
	"question1":             Answer{Single: "c"},
	"question2":             Answer{Single: "a"},
	"question-img1":         Answer{Single: "b"},
	"question-img2":         Answer{Single: "d"},
	"question-source-link1": Answer{Single: "c"},
	"question-source-link2": Answer{Single: "b"},
}

func TestNewQuestionMarkdown(t *testing.T) {
	for name := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			got, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)

			assert.Equal(t, fname, got.Filename)
			want := FrontmatterType("question")
			assert.Equal(t, want, got.Frontmatter.Type)
		})
	}
}

func TestValidateAnswer(t *testing.T) {
	for name := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)
			err = md.Verify("")
			assert.NoError(t, err)
		})
	}
}

func TestExportAnswer(t *testing.T) {
	for name, want := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)
			privateKey := ""
			gotAnswerkey, err := md.ExportAnswerkey(privateKey)
			assert.NoError(t, err)
			got := gotAnswerkey["course1"]["unit1"]["exercise1"][name]
			assert.Equal(t, true, want.Equals(got))
		})
	}
}

func TestSealAnswer(t *testing.T) {
	for name, answer := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)

			err = md.Seal(testKeyPublic)
			assert.NoError(t, err)
			assert.Equal(t, "", md.Frontmatter.Answer)
			unsealedAnswer, err := Decrypt(testKeyPrivate, md.Frontmatter.SealedAnswer)
			assert.NoError(t, err)
			want := answer.correctAnswers()
			assert.Equal(t, want, unsealedAnswer)
		})
	}
}

func TestUnsealAnswer(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/question1-sealed.md"
	md, err := NewQuestionMarkdown(fname)
	assert.NoError(t, err)
	assert.Equal(t, "", md.Frontmatter.Answer)

	err = md.Unseal(testKeyPrivate)
	assert.NoError(t, err)
	assert.Equal(t, "c", md.Frontmatter.Answer)
	assert.Equal(t, "", md.Frontmatter.SealedAnswer)
}

func TestExportAnswerkeyFromSeal(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/question1-sealed.md"
	md, err := NewQuestionMarkdown(fname)
	assert.NoError(t, err)
	gotAnswerkey, err := md.ExportAnswerkey(testKeyPrivate)
	assert.NoError(t, err)
	want := Answer{Single: "c"}
	got := gotAnswerkey["course1"]["unit1"]["exercise1"]["question1-sealed"]
	assert.Equal(t, true, want.Equals(got))
}

func TestErrInvalidAnswer(t *testing.T) {
	errQuestions := []string{
		"err-false-positive",
		"err-false-negative",
		"err-img1",
		"err-img2",
	}
	for _, name := range errQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)
			err = md.Verify("")
			assert.Error(t, ErrInvalidAnswer, err)
		})
	}
}
