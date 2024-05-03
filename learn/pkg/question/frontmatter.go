package question

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Type         frontmatterType `yaml:"type,omitempty"` // question
	Difficulty   difficulty      `yaml:"difficulty,omitempty"`
	AnswerType   answerType      `yaml:"answer-type,omitempty"` // single-choice, multiple-choice, free-text, multiple-free-texts, program
	Answer       string          `yaml:"answer,omitempty"`
	SealedAnswer string          `yaml:"sealed-answer,omitempty"`
}

func (f *frontmatter) validate() error {
	if f.Type != "question" {
		return fmt.Errorf("%w: want: %q, got: %q", ErrWrongFrontmatterType, "question", f.Type)
	}
	if f.Answer == "" && f.SealedAnswer == "" {
		return fmt.Errorf("no answer found: %w", ErrNoFrontmatterAnswer)
	}
	if f.Answer != "" && f.SealedAnswer != "" {
		return fmt.Errorf("%w: sealed and unsealed answer found, only one allowed", ErrInvalidFrontmatter)
	}
	return nil
}

func (f *frontmatter) getAnswer(privateKey string) (Answer, error) {
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
		return Answer{}, fmt.Errorf("cannot get answerkey: %w", ErrNoFrontmatterAnswer)
	}
	return NewAnswer(f.AnswerType, text)
}

func (f *frontmatter) Seal(publicKey string) error {
	if f.Answer == "" && f.SealedAnswer != "" {
		return nil // already sealed
	}
	if f.Answer == "" {
		return fmt.Errorf("cannot seal empty answer: %w", ErrNoFrontmatterAnswer)
	}
	sealed, err := Encrypt(publicKey, f.Answer)
	if err != nil {
		return err
	}
	f.SealedAnswer = sealed
	f.Answer = ""
	return nil
}

func (f *frontmatter) Unseal(privateKey string) error {
	if f.Answer != "" && f.SealedAnswer == "" {
		return nil // already unsealed
	}
	if f.SealedAnswer == "" {
		return fmt.Errorf("cannot unseal empty sealed-answer: %w", ErrNoFrontmatterAnswer)
	}
	unsealed, err := Decrypt(privateKey, f.SealedAnswer)
	if err != nil {
		return err
	}
	f.SealedAnswer = ""
	f.Answer = unsealed
	return nil
}

type (
	frontmatterType string
	answerType      string
	difficulty      string
)

var (
	validFrontmatterTypes = []string{"course", "unit", "exercise", "question"}
	validAnswerTypes      = []string{"single-choice", "multiple-choice", "free-text", "multiple-free-texts", "program"}
	validDifficulties     = []string{"easy", "medium", "hard", "retriable"}
)

func (s frontmatterType) MarshalText() ([]byte, error) {
	return marshalText("type", string(s), validFrontmatterTypes)
}

func (s answerType) MarshalText() ([]byte, error) {
	return marshalText("answer-type", string(s), validAnswerTypes)
}

func (s difficulty) MarshalText() ([]byte, error) {
	return marshalText("difficulty", string(s), validDifficulties)
}

func marshalText(fieldName, str string, validStrings []string) ([]byte, error) {
	if !slices.Contains(validStrings, str) {
		return nil, fmt.Errorf(`%w: marshal: invalid frontmatter %q: %q, use one of %v`, ErrInvalidFrontmatter, fieldName, str, validStrings)
	}
	return []byte(str), nil
}

func (s *frontmatterType) UnmarshalText(text []byte) error {
	return unmarshalText("type", validFrontmatterTypes, text, (*string)(s))
}

func (s *answerType) UnmarshalText(text []byte) error {
	return unmarshalText("answer-type", validAnswerTypes, text, (*string)(s))
}

func (s *difficulty) UnmarshalText(text []byte) error {
	return unmarshalText("difficulty", validDifficulties, text, (*string)(s))
}

func unmarshalText(fieldName string, validStrings []string, text []byte, s *string) error {
	str := string(text)
	if !slices.Contains(validStrings, str) {
		return fmt.Errorf(`%w: unmarshal: invalid frontmatter %q: %q, use one of %v`, ErrInvalidFrontmatter, fieldName, str, validStrings)
	}
	*s = str
	return nil
}

func parseFrontmatter(str string) (*frontmatter, error) {
	fm := &frontmatter{}
	if err := yaml.Unmarshal([]byte(str), fm); err != nil {
		return nil, fmt.Errorf("%w: cannot process Question Markdown frontmatter: %w", ErrInvalidFrontmatter, err)
	}
	if err := fm.validate(); err != nil {
		return nil, err
	}
	return fm, nil
}
