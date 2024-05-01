package answer

import (
	"fmt"
	"slices"
)

var ErrNoFrontmatter = fmt.Errorf("no frontmatter found (---)")
var ErrInvalidFrontmatter = fmt.Errorf("invalid frontmatter")
var ErrWrongFrontmatterType = fmt.Errorf("wrong frontmatter type")
var ErrSealedAnswerNoKey = fmt.Errorf("sealed answer without key")
var ErrNoAnswer = fmt.Errorf("answer not found")

type Frontmatter struct {
	Type FrontmatterType // course, unit, exercise, question
}

type QuestionFrontmatter struct {
	Type         FrontmatterType     `yaml:"type,omitempty"` // question
	Difficulty   Difficulty          `yaml:"difficulty,omitempty"`
	Unverifiable bool                `yaml:"unverifiable,omitempty"`
	Substituions []map[string]string `yaml:"substitutions,omitempty"`
	AnswerType   AnswerType          `yaml:"answer-type,omitempty"` // single-choice, multiple-choice, free-text, multiple-free-texts, program
	Answer       string              `yaml:"answer,omitempty"`
	SealedAnswer string              `yaml:"sealed-answer,omitempty"`
}

func (f *QuestionFrontmatter) Validate() error {
	if f.Type != "question" {
		return fmt.Errorf("%w: want: %q, got: %q", ErrWrongFrontmatterType, "question", f.Type)
	}
	if f.Answer == "" && f.SealedAnswer == "" {
		return fmt.Errorf("no answer found: %w", ErrNoAnswer)
	}
	if f.Answer != "" && f.SealedAnswer != "" {
		return fmt.Errorf("%w: sealed and unsealed answer found, only one allowed", ErrInvalidFrontmatter)
	}
	return nil
}

func (f *QuestionFrontmatter) GetAnswer(privateKey string) (Answer, error) {
	text := f.Answer
	if f.SealedAnswer != "" && privateKey == "" {
		return Answer{}, ErrSealedAnswerNoKey
	}
	if f.SealedAnswer != "" {
		var err error
		text, err = Decrypt(privateKey, f.SealedAnswer)
		if err != nil {
			return Answer{}, err
		}
	}
	if text == "" {
		return Answer{}, fmt.Errorf("cannot get answerkey: %w", ErrNoAnswer)
	}
	return NewAnswer(f.AnswerType, text)
}

func (f *QuestionFrontmatter) Seal(publicKey string) error {
	if f.Answer == "" && f.SealedAnswer != "" {
		return nil // already sealed
	}
	if f.Answer == "" {
		return fmt.Errorf("cannot seal empty answer: %w", ErrNoAnswer)
	}
	sealed, err := Encrypt(publicKey, f.Answer)
	if err != nil {
		return err
	}
	f.SealedAnswer = sealed
	f.Answer = ""
	return nil
}

func (f *QuestionFrontmatter) Unseal(privateKey string) error {
	if f.Answer != "" && f.SealedAnswer == "" {
		return nil // already unsealed
	}
	if f.SealedAnswer == "" {
		return fmt.Errorf("cannot unseal empty sealed-answer: %w", ErrNoAnswer)
	}
	unsealed, err := Decrypt(privateKey, f.SealedAnswer)
	if err != nil {
		return err
	}
	f.SealedAnswer = ""
	f.Answer = unsealed
	return nil
}

type FrontmatterType string

var validFrontmatterTypes = []string{"course", "unit", "exercise", "question"}

func (s FrontmatterType) MarshalText() ([]byte, error) {
	return marshalText("frontmatter 'type'", string(s), validFrontmatterTypes)
}

func (s *FrontmatterType) UnmarshalText(text []byte) error {
	return unmarshalText("frontmatter 'type'", validFrontmatterTypes, text, (*string)(s))
}

type AnswerType string

var validAnswerTypes = []string{"single-choice", "multiple-choice", "free-text", "multiple-free-texts", "program"}

func (s AnswerType) MarshalText() ([]byte, error) {
	return marshalText("frontmatter 'sub-type'", string(s), validAnswerTypes)
}

func (s *AnswerType) UnmarshalText(text []byte) error {
	return unmarshalText("frontmatter 'sub-type'", validAnswerTypes, text, (*string)(s))
}

type Difficulty string

var validDifficultys = []string{"easy", "medium", "hard", "retriable"}

func (s Difficulty) MarshalText() ([]byte, error) {
	return marshalText("frontmatter 'difficulty'", string(s), validDifficultys)
}

func (s *Difficulty) UnmarshalText(text []byte) error {
	return unmarshalText("frontmatter 'difficulty'", validDifficultys, text, (*string)(s))
}

func marshalText(fieldName, str string, validStrings []string) ([]byte, error) {
	if !slices.Contains(validStrings, str) {
		return nil, fmt.Errorf(`marshal: invalid %s: %q, use one of %v`, fieldName, str, validStrings)
	}
	return []byte(str), nil
}

func unmarshalText(fieldName string, validStrings []string, text []byte, s *string) error {
	str := string(text)
	if !slices.Contains(validStrings, str) {
		return fmt.Errorf(`unmarshal: invalid %s: %q, use one of %v`, fieldName, str, validStrings)
	}
	*s = str
	return nil
}
