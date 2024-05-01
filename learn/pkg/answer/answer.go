package answer

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

var ErrSingleChoice = errors.New("single-choice answer must be a single character a-z")

type Answerkey map[string]Coursekey
type Coursekey map[string]Unitkey
type Unitkey map[string]Exercisekey
type Exercisekey map[string]Answer
type Answer struct {
	Single  string   `firestore:"single,omitempty" json:"single,omitempty"`
	Multi   []string `firestore:"multi,omitempty" json:"multi,omitempty"`
	Text    string   `firestore:"text,omitempty" json:"text,omitempty"`
	Texts   []string `firestore:"texts,omitempty" json:"texts,omitempty"`
	Program string   `firestore:"program,omitempty" json:"program,omitempty"`

	Type AnswerType `firestore:"-" json:"-"`
}

func NewAnswer(answerType AnswerType, text string) (Answer, error) {
	a := Answer{Type: answerType}
	switch answerType {
	case "single-choice":
		if err := validateSingle(text); err != nil {
			return Answer{}, fmt.Errorf("invalid single-choice answer: %w", err)
		}
		a.Single = text
	case "multiple-choice":
		multi := splitTrim(text)
		if err := validateMulti(multi); err != nil {
			return Answer{}, fmt.Errorf("invalid single-choice answer: %w", err)
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

func NewAnswerkey(filename string, answer Answer) (Answerkey, error) {
	qp, err := newAnswerPath(filename)
	if err != nil {
		return nil, err
	}
	answerkey := Answerkey{
		qp.course: Coursekey{
			qp.unit: Unitkey{
				qp.exercise: Exercisekey{
					qp.question: answer,
				},
			},
		},
	}
	return answerkey, nil
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

func (a Answer) Equals(other Answer) bool {
	if a.Type != other.Type {
		return false
	}
	switch a.Type {
	case "single-choice":
		return a.Single == other.Single
	case "text":
		return a.Text == other.Text
	case "program":
		return a.Program == other.Program // not quite correct!
	case "multiple-choice":
		for i, v := range a.Multi {
			if v != other.Multi[i] {
				return false
			}
		}
		return true
	case "texts":
		for i, v := range a.Texts {
			if v != other.Texts[i] {
				return false
			}
		}
		return true
	}
	return false

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
	return "UNKOWN ANSWER"
}

func indexToLetter(i int) string {
	return string(rune('a' + i))
}

func newAnswerPath(filename string) (answerPath, error) {
	// path: /..../COURSE/UNIT/EXERCISE/questions/QUESTION.md
	path, err := filepath.Abs(filename)
	if err != nil {
		return answerPath{}, err
	}
	segments := split(path)
	if len(segments) < 5 {
		return answerPath{}, fmt.Errorf("not enough directories in path: %v, want /..../COURSE/UNIT/EXERCISE/questions/QUESTION.md", segments)
	}
	slices.Reverse(segments)
	if segments[1] != "questions" {
		return answerPath{}, fmt.Errorf("expected 'questions' directory, found %q", segments[1])
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
	if last == "" {
		return []string{last}
	}
	return append(split(filepath.Clean(dir)), last)
}
