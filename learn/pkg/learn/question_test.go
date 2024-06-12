package learn

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

var testQuestions = map[string]Answer{
	"question1":      {Single: "c"},
	"question2":      {Single: "a"},
	"question-img1":  {Single: "b"},
	"question-img2":  {Single: "d"},
	"question-link1": {Single: "c"},
	"question-link2": {Single: "c"},
	"question-link3": {Single: "d"},
	"question-link4": {Multi: []string{"b", "c"}},
}

var allTestQuestions = []string{
	"question1",
	"question1-sealed",
	"question2",
	"question-img1",
	"question-img2",
	"question-link1",
	"question-link2",
	"question-link3",
	"question-link4",
}

func TestNewQuestionModel(t *testing.T) {
	for name := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			got, err := NewQuestionModel(fname)
			assert.NoError(t, err)

			assert.Equal(t, fname, got.Filename)
			want := frontmatterType("question")
			assert.Equal(t, want, got.Frontmatter.Type)
		})
	}
}

func TestValidateAnswer(t *testing.T) {
	for name := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			err = model.Verify()
			assert.NoError(t, err)
		})
	}
}

func TestExportAnswer(t *testing.T) {
	for name, want := range testQuestions {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile("testdata/golden/answerkey/" + name + ".json")
			assert.NoError(t, err)
			wantAnswerKey := AnswerKey{}
			err = json.Unmarshal(b, &wantAnswerKey)
			assert.NoError(t, err)

			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			gotAnswerKey, err := model.ExportAnswerKey()
			assert.NoError(t, err)

			model, err = NewQuestionModel(fname, WithPrivateKey(testKeyPrivate))
			assert.NoError(t, err)
			gotAnswerKeyWithSealed, err := model.ExportAnswerKey()
			assert.NoError(t, err)

			assert.Equal(t, wantAnswerKey, gotAnswerKey)
			assert.Equal(t, wantAnswerKey, gotAnswerKeyWithSealed)

			got := gotAnswerKey["course1"]["unit1"]["exercise1"][name]
			assert.Equal(t, true, want.Equals(got))
		})
	}
}

func TestSealAnswer(t *testing.T) {
	for name, answer := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)

			err = model.Seal(testKeyPublic)
			assert.NoError(t, err)
			assert.Equal(t, "", model.Frontmatter.Answer)
			unsealedAnswer, err := Decrypt(testKeyPrivate, model.Frontmatter.SealedAnswer)
			assert.NoError(t, err)
			want := answer.correctAnswers()
			assert.Equal(t, want, unsealedAnswer)
		})
	}
}

func TestUnsealAnswer(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/question1-sealed.md"
	model, err := NewQuestionModel(fname, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	assert.Equal(t, "", model.Frontmatter.Answer)

	err = model.Unseal()
	assert.NoError(t, err)
	assert.Equal(t, "c", model.Frontmatter.Answer)
	assert.Equal(t, "", model.Frontmatter.SealedAnswer)
}

func TestExportAnswerKeyFromSeal(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/question1-sealed.md"
	model, err := NewQuestionModel(fname, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	gotAnswerKey, err := model.ExportAnswerKey()
	assert.NoError(t, err)

	b, err := os.ReadFile("testdata/golden/answerkey/question1-sealed.json")
	assert.NoError(t, err)
	wantAnswerKey := AnswerKey{}
	err = json.Unmarshal(b, &wantAnswerKey)
	assert.NoError(t, err)
	assert.Equal(t, wantAnswerKey, gotAnswerKey)

	want := Answer{Single: "c"}
	got := gotAnswerKey["course1"]["unit1"]["exercise1"]["question1-sealed"]
	assert.Equal(t, true, want.Equals(got))

	model, err = NewQuestionModel(fname, WithIgnoreSealed())
	assert.NoError(t, err)
	gotAnswerKey, err = model.ExportAnswerKey()
	assert.NoError(t, err)
	assert.Equal(t, AnswerKey{}, gotAnswerKey)

	model, err = NewQuestionModel(fname) // no key for sealed question, expect error
	assert.NoError(t, err)
	_, err = model.ExportAnswerKey()
	assert.Error(t, ErrSealedAnswerNoKey, err)
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
			fname := "testdata/err-course/unit1/err-exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			err = model.Verify()
			assert.Error(t, ErrWrongAnswer, err)
		})
	}
}

func TestErrNoExistMD(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/MISSING-FILE.md"
	_, err := NewQuestionModel(fname)
	assert.Error(t, os.ErrNotExist, err)
}

func TestErrNoExistSVG(t *testing.T) {
	fname := "testdata/course1/unit1/err-exercise1/err-img3.md"
	_, err := NewQuestionModel(fname)
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
			fname := "testdata/err-course/unit1/err-exercise1/" + name + ".md"
			_, err := NewQuestionModel(fname)
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
			fname := "testdata/err-course/unit1/err-exercise1/" + name + ".md"
			_, err := NewQuestionModel(fname)
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
			fname := "testdata/err-course/unit1/err-exercise1/" + name + ".md"
			_, err := NewQuestionModel(fname)
			assert.Error(t, ErrInconsistentMdoel, err)
		})
	}
}

func TestRendererTracking(t *testing.T) {
	embeds := map[string][]string{
		"question1":      nil,
		"question2":      nil,
		"question-img1":  {"dot-dot-a-evy-svg", "dot-dot-b-evy-svg", "dot-dot-c-evy-svg", "dot-dot-d-evy-svg"},
		"question-img2":  {"dot-dot-a-evy-svg"},
		"question-link1": {"dot-dot-c-evy", "dot-dot-a-evy", "dot-dot-b-evy", "dot-dot-c-evy", "dot-dot-d-evy"},
		"question-link2": {"dot-dot-c-evy", "dot-dot-a-evy", "dot-dot-b-evy", "dot-dot-c-evy", "dot-dot-d-evy"},
		"question-link3": {"print-print-b-evy", "print-print-a-evy", "print-print-b-evy", "print-print-c-evy", "print-print-d-evy"},
		"question-link4": {"print-print-d-evy", "print-print-a-evy", "print-print-b-evy", "print-print-c-evy", "print-print-d-evy"},
	}
	for name, want := range embeds {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			got := model.embeds
			assert.Equal(t, len(want), len(got))
			m := map[string]bool{}
			for _, w := range want {
				m[w] = true
			}
			for _, g := range got {
				assert.Equal(t, true, m[g.id], "cannot find "+g.id)
			}
		})
	}
}

func TestPrintHTML(t *testing.T) {
	for _, name := range allTestQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			buf := &bytes.Buffer{}
			err = model.PrintHTML(buf, false) // without marked answers
			assert.NoError(t, err)
			got := buf.String()
			goldenFile := "testdata/golden/question/" + name + ".html"
			b, err := os.ReadFile(goldenFile)
			assert.NoError(t, err)
			want := string(b)
			assert.Equal(t, want, got)
		})
	}
}

func TestToHTML(t *testing.T) {
	for _, name := range allTestQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			got, err := model.ToHTML(false) // without marked
			assert.NoError(t, err)

			goldenFile := "testdata/golden/question/" + name + ".html"
			b, err := os.ReadFile(goldenFile)
			assert.NoError(t, err)
			want := string(b)
			assert.Equal(t, want, got)
		})
	}
}

func TestToHTMLWithAnswersMarked(t *testing.T) {
	for name := range testQuestions {
		t.Run(name, func(t *testing.T) {
			fname := "testdata/course1/unit1/exercise1/" + name + ".md"
			model, err := NewQuestionModel(fname)
			assert.NoError(t, err)
			got, err := model.ToHTML(true) // with marked
			assert.NoError(t, err)

			goldenFile := "testdata/golden/question-with-marked/" + name + ".html"
			b, err := os.ReadFile(goldenFile)
			assert.NoError(t, err)
			want := string(b)
			assert.Equal(t, want, got)

			model, err = NewQuestionModel(fname, WithIgnoreSealed())
			assert.NoError(t, err)
			got, err = model.ToHTML(true) // with marked
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestToHTMLWithAnswersMarkedSealErr(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/question1-sealed.md"
	model, err := NewQuestionModel(fname)
	assert.NoError(t, err)
	_, err = model.ToHTML(true) // with marked
	assert.Error(t, ErrSealedAnswerNoKey, err)

	model, err = NewQuestionModel(fname, WithIgnoreSealed())
	assert.NoError(t, err)
	got, err := model.ToHTML(true) // with marked
	assert.NoError(t, err)

	goldenFile := "testdata/golden/question-with-marked/question1-sealed-unsealed-only.html"
	b, err := os.ReadFile(goldenFile)
	assert.NoError(t, err)
	want := string(b)
	assert.Equal(t, want, got)
}

func TestToHTMLWithAnswersMarkedSealed(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/question1-sealed.md"
	model, err := NewQuestionModel(fname, WithPrivateKey(testKeyPrivate))
	assert.NoError(t, err)
	got, err := model.ToHTML(true) // withAnswersMarked
	assert.NoError(t, err)

	goldenFile := "testdata/golden/question-with-marked/question1-sealed.html"
	b, err := os.ReadFile(goldenFile)
	assert.NoError(t, err)
	want := string(b)
	assert.Equal(t, want, got)
}
