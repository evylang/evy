package learn

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
	case "text":
		a.Text = text
	case "multiple-texts":
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
	AnswerPath, err := NewAnswerPath(filename)
	if err != nil {
		return nil, err
	}
	answerKey := AnswerKey{}
	answerKey.add(AnswerPath, answer)
	return answerKey, nil
}

func (key AnswerKey) add(p AnswerPath, answer Answer) {
	if key[p.Course] == nil {
		key[p.Course] = CourseKey{}
	}
	if key[p.Course][p.Unit] == nil {
		key[p.Course][p.Unit] = UnitKey{}
	}
	if key[p.Course][p.Unit][p.Exercise] == nil {
		key[p.Course][p.Unit][p.Exercise] = ExerciseKey{}
	}
	key[p.Course][p.Unit][p.Exercise][p.Question] = answer
}

// merge merges the the given AnswerKeys parameter into the answerkey
// receiver. The parameter is not modified, the receiver is updated.
func (key AnswerKey) merge(answerKey AnswerKey) {
	for course, courseKey := range answerKey {
		if key[course] == nil {
			key[course] = CourseKey{}
		}
		for unit, unitKey := range courseKey {
			if key[course][unit] == nil {
				key[course][unit] = UnitKey{}
			}
			for exercise, exerciseKey := range unitKey {
				if key[course][unit][exercise] == nil {
					key[course][unit][exercise] = ExerciseKey{}
				}
				for question, answer := range exerciseKey {
					key[course][unit][exercise][question] = answer
				}
			}
		}
	}
}

// Type returns the type of the answer, one of:
// - single-choice
// - multiple-choice
// - text
// - multiple-texts
// - program.
func (a Answer) Type() string {
	switch {
	case a.Single != "":
		return "single-choice"
	case len(a.Multi) > 0:
		return "multiple-choice"
	case a.Text != "":
		return "text"
	case len(a.Texts) > 0:
		return "multiple-texts"
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

// AnswerPath is a composite key of a course, unit, exercise and question,
// which uniquely identifies a question-answer pair.
type AnswerPath struct {
	Course   string
	Unit     string
	Exercise string
	Question string
}

// NewAnswerPath creates a new AnswerPath from a question markdown filename.
// The required path structure is
//
//	/..../COURSE/UNIT/EXERCISE/QUESTION.md
func NewAnswerPath(filename string) (AnswerPath, error) {
	path, err := filepath.Abs(filename)
	if err != nil {
		return AnswerPath{}, err
	}
	segments := splitPath(path)
	if len(segments) < 4 {
		return AnswerPath{}, fmt.Errorf("%w: not enough directories in path: %v, want /..../COURSE/UNIT/EXERCISE/QUESTION.md", ErrBadDirectoryStructure, segments)
	}
	slices.Reverse(segments)
	return AnswerPath{
		Course:   segments[3],
		Unit:     segments[2],
		Exercise: segments[1],
		Question: baseNoExt(segments[0]),
	}, nil
}

func baseNoExt(filename string) string {
	filename = filepath.Base(filename)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func splitPath(path string) []string {
	dir, last := filepath.Split(path)
	if dir == "" || dir == "/" {
		return []string{last}
	}
	return append(splitPath(filepath.Clean(dir)), last)
}
