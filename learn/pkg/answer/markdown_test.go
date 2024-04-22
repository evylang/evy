package answer

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestNewQuestionMarkdown(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/question1.md"
	got, err := NewQuestionMarkdown(fname)
	assert.NoError(t, err)

	assert.Equal(t, fname, got.Filename)
	want := FrontmatterType("question")
	assert.Equal(t, want, got.Frontmatter.Type)
}

func TestExportAnswer(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/question1.md"
	md, err := NewQuestionMarkdown(fname)
	assert.NoError(t, err)
	privateKey := ""
	gotAnswerkey, err := md.ExportAnswerKey(privateKey)
	assert.NoError(t, err)
	want := Answer{Single: "c"}
	got := gotAnswerkey["course1"]["unit1"]["exercise1"]["question1"]
	assert.Equal(t, true, want.Equals(got))
}

func TestSealAnswer(t *testing.T) {
	fname := "testdata/course1/unit1/exercise1/questions/question1.md"
	md, err := NewQuestionMarkdown(fname)
	assert.NoError(t, err)

	err = md.Seal(testKeyPublic)
	assert.NoError(t, err)
	assert.Equal(t, "", md.Frontmatter.Answer)
	unsealedAnswer, err := Decrypt(testKeyPrivate, md.Frontmatter.SealedAnswer)
	assert.NoError(t, err)
	assert.Equal(t, "c", unsealedAnswer)
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
	gotAnswerkey, err := md.ExportAnswerKey(testKeyPrivate)
	assert.NoError(t, err)
	want := Answer{Single: "c"}
	got := gotAnswerkey["course1"]["unit1"]["exercise1"]["question1-sealed"]
	assert.Equal(t, true, want.Equals(got))
}
