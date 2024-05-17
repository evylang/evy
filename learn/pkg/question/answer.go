package question

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type (
	// AnswerKey is the top-level map of multiple course answerKeys.
	AnswerKey map[string]CourseKey
	// CourseKey is the full AnswerKey of a course indexed by unit ID.
	CourseKey map[string]UnitKey
	// UnitKey is the AnswerKey of a single unit indexed by exercise ID.
	UnitKey map[string]ExerciseKey
	// ExerciseKey is the AnswerKey of a single exercise indexed by question ID.
	ExerciseKey map[string]Answer
	// Answer a single, concrete answer to question. It may be only one of: single
	// or multiple choice, free text, multiple free texts or program.
	Answer struct {
		Single  string   `firestore:"single,omitempty" json:"single,omitempty"`
		Multi   []string `firestore:"multi,omitempty" json:"multi,omitempty"`
		Text    string   `firestore:"text,omitempty" json:"text,omitempty"`
		Texts   []string `firestore:"texts,omitempty" json:"texts,omitempty"`
		Program string   `firestore:"program,omitempty" json:"program,omitempty"`
	}
)

// NewAnswer creates a new Answer from an answer type and a raw answer string.
func NewAnswer(answerType answerType, text string) (Answer, error) {
	a := Answer{}
	switch answerType {
	case "single-choice":
		if err := validateSingle(text); err != nil {
			return Answer{}, fmt.Errorf("invalid single-choice answer: %w", err)
		}
		a.Single = text
	case "multiple-choice":
		multi := splitTrim(text)
		if err := validateMulti(multi); err != nil {
			return Answer{}, fmt.Errorf("invalid multiple-choice answer: %w", err)
		}
		a.Multi = multi
	case "free-text":
		a.Text = text
	case "multiple-free-text":
		a.Texts = splitTrim(text)
	case "program":
		a.Program = text
	}
	return a, nil
}

// NewAnswerKey creates a new AnswerKey from a filename and an answer. The
// filename is used to generate the composite key of the answer. It is split
// into course, unit, exercise and answer key and an Answer.
func NewAnswerKey(filename string, answer Answer) (AnswerKey, error) {
	qp, err := newAnswerPath(filename)
	if err != nil {
		return nil, err
	}
	answerKey := AnswerKey{
		qp.course: CourseKey{
			qp.unit: UnitKey{
				qp.exercise: ExerciseKey{
					qp.question: answer,
				},
			},
		},
	}
	return answerKey, nil
}

// Type returns the type of the answer, one of:
// - single-choice
// - multiple-choice
// - free-text
// - multiple-free-text
// - program.
func (a Answer) Type() string {
	switch {
	case a.Single != "":
		return "single-choice"
	case len(a.Multi) > 0:
		return "multiple-choice"
	case a.Text != "":
		return "free-text"
	case len(a.Texts) > 0:
		return "multiple-free-text"
	case a.Program != "":
		return "program"
	}
	return "UNKNOWN ANSWER TYPE"
}

// Equals returns true if the two answers are equal.
func (a Answer) Equals(other Answer) bool {
	return a.Single == other.Single && a.Text == other.Text && a.Program == other.Program &&
		slices.Equal(a.Multi, other.Multi) && slices.Equal(a.Texts, other.Texts)
}

func (a Answer) correctAnswerIndices() map[int]bool {
	m := make(map[int]bool)
	switch {
	case a.Single != "":
		m[int(a.Single[0]-'a')] = true
	case len(a.Multi) > 0:
		for _, s := range a.Multi {
			m[int(s[0]-'a')] = true
		}
	}
	return m
}

func (a Answer) correctAnswers() string {
	switch {
	case a.Single != "":
		return a.Single
	case len(a.Multi) > 0:
		return strings.Join(a.Multi, ", ")
	case a.Text != "":
		return a.Text
	case len(a.Texts) > 0:
		return strings.Join(a.Texts, ", ")
	case a.Program != "":
		return a.Program
	}
	return "UNKNOWN ANSWER"
}

func indexToLetter(i int) string {
	return string(rune('a' + i))
}

type answerPath struct {
	course   string
	unit     string
	exercise string
	question string
}

func validateSingle(str string) error {
	if len(str) != 1 || str[0] < 'a' || str[0] > 'z' {
		return ErrSingleChoice
	}
	return nil
}

func validateMulti(ss []string) error {
	for _, s := range ss {
		if err := validateSingle(s); err != nil {
			return err
		}
	}
	return nil
}

func splitTrim(str string) []string {
	ss := strings.Split(str, ",")
	for i, s := range ss {
		ss[i] = strings.TrimSpace(s)
	}
	return ss
}

func newAnswerPath(filename string) (answerPath, error) {
	// path: /..../COURSE/UNIT/EXERCISE/questions/QUESTION.md
	path, err := filepath.Abs(filename)
	if err != nil {
		return answerPath{}, err
	}
	segments := split(path)
	if len(segments) < 5 {
		return answerPath{}, fmt.Errorf("%w: not enough directories in path: %v, want /..../COURSE/UNIT/EXERCISE/questions/QUESTION.md", ErrBadDirectoryStructure, segments)
	}
	slices.Reverse(segments)
	if segments[1] != "questions" {
		return answerPath{}, fmt.Errorf("%w: expected 'questions' directory, found %q", ErrBadDirectoryStructure, segments[1])
	}
	return answerPath{
		course:   segments[4],
		unit:     segments[3],
		exercise: segments[2],
		question: strings.TrimSuffix(segments[0], filepath.Ext(segments[0])),
	}, nil
}

func split(path string) []string {
	dir, last := filepath.Split(path)
	if dir == "" || dir == "/" {
		return []string{last}
	}
	return append(split(filepath.Clean(dir)), last)
}
