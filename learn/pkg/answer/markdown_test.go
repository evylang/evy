package answer

import (
	"os"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

var testQuestions = map[string]Answer{
	"question1":      {Single: "c", Type: "single-choice"},
	"question2":      {Single: "a", Type: "single-choice"},
	"question-img1":  {Single: "b", Type: "single-choice"},
	"question-img2":  {Single: "d", Type: "single-choice"},
	"question-link1": {Single: "c", Type: "single-choice"},
	"question-link2": {Single: "c", Type: "single-choice"},
	"question-link3": {Single: "d", Type: "single-choice"},
	"question-link4": {Multi: []string{"b", "c"}, Type: "multiple-choice"},
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
	want := Answer{Single: "c", Type: "single-choice"}
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

func TestErrNoExistMD(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/MISSING-FILE.md"
	_, err := NewQuestionMarkdown(fname)
	assert.Error(t, os.ErrNotExist, err)
}

func TestErrNoExistSVG(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/err-img3.md"
	md, err := NewQuestionMarkdown(fname)
	assert.NoError(t, err)
	err = md.Verify("")
	assert.Error(t, os.ErrNotExist, err)
}

func TestErrBadMDImg(t *testing.T) {
	errQuestions := []string{
		"err-img4",
		"err-img5",
		"err-img6",
		"err-img7",
	}
	for _, name := range errQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)
			err = md.Verify("")
			assert.Error(t, ErrBadMarkdownStructure, err)
		})
	}
}
func TestErrBadMDLink(t *testing.T) {
	errQuestions := []string{
		"err-link1",
		"err-link2",
		"err-link3",
		"err-link4",
		"err-link5",
		"err-link6",
	}
	for _, name := range errQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)
			err = md.Verify("")
			assert.Error(t, ErrBadMarkdownStructure, err)
		})
	}
}

func TestErrInconsistency(t *testing.T) {
	errQuestions := []string{
		"err-inconsistent1",
		"err-inconsistent2",
	}
	for _, name := range errQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/questions/" + name + ".md"
			md, err := NewQuestionMarkdown(fname)
			assert.NoError(t, err)
			err = md.Verify("")
			assert.Error(t, ErrInconsistentMdoel, err)
		})
	}
}
