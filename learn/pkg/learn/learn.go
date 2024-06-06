// Package learn provides data structures and tools for Evy learn resources.
// Question, exercises, units and courses are parsed from Markdown files with
// YAML frontmatter. The frontmatter serves as a small set of structured data
// associated with the unstructured Markdown file.
//
// Question can be verified to have the expected correct answer output match
// the question output. Questions, can seal (encrypt) their answers in the
// Frontmatter or unsealed (decrypted) them. We use this to avoid openly
// publishing the answerKey. Questions can also export their AnswerKeys into
// single big JSON object as used in Evy's persistent data store(Firestore).
// See the testdata/ directory for sample question and answers.
package learn

import "errors"

// Errors for the learn package.
var (
	ErrBadMarkdownStructure = errors.New("bad Markdown structure")
	ErrInconsistentMdoel    = errors.New("inconsistency")
	ErrWrongAnswer          = errors.New("wrong answer")

	ErrSingleChoice          = errors.New("single-choice answer must be a single character a-z")
	ErrBadDirectoryStructure = errors.New("bad directory structure for course layout")

	ErrNoFrontmatter        = errors.New("no frontmatter found")
	ErrInvalidFrontmatter   = errors.New("invalid frontmatter")
	ErrWrongFrontmatterType = errors.New("wrong frontmatter type")
	ErrNoFrontmatterAnswer  = errors.New("no answer in frontmatter")
	ErrSealedAnswerNoKey    = errors.New("sealed answer without key")
	ErrSealedTooShort       = errors.New("sealed data is too short")
)
