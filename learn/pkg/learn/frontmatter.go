package learn

import (
	"fmt"
	"slices"
)

type (
	frontmatterType string
	answerType      string
	difficulty      string
	verification    string
)

var (
	validFrontmatterTypes = []string{"course", "unit", "unittest", "quiz", "exercise", "question"}
	validAnswerTypes      = []string{"single-choice", "multiple-choice", "text", "multiple-texts", "program"}
	validDifficulties     = []string{"easy", "medium", "hard", "retriable"}
	validVerifications    = []string{"match" /* default */, "none", "parse-error", "no-parse-error"}
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

func (s verification) MarshalText() ([]byte, error) {
	return marshalText("difficulty", string(s), validVerifications)
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

func (s *verification) UnmarshalText(text []byte) error {
	return unmarshalText("difficulty", validVerifications, text, (*string)(s))
}

func unmarshalText(fieldName string, validStrings []string, text []byte, s *string) error {
	str := string(text)
	if !slices.Contains(validStrings, str) {
		return fmt.Errorf(`%w: unmarshal: invalid frontmatter %q: %q, use one of %v`, ErrInvalidFrontmatter, fieldName, str, validStrings)
	}
	*s = str
	return nil
}
