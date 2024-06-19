package learn

import (
	"fmt"
)

type questionFrontmatter struct {
	Type              frontmatterType `yaml:"type,omitempty"` // question
	Difficulty        difficulty      `yaml:"difficulty,omitempty"`
	AnswerType        answerType      `yaml:"answer-type,omitempty"` // single-choice, multiple-choice, text, multiple-texts, program
	GenerateQuestions string          `yaml:"generate-questions,omitempty"`
	Answer            string          `yaml:"answer,omitempty"`
	SealedAnswer      string          `yaml:"sealed-answer,omitempty"`
	Verification      verification    `yaml:"verification,omitempty"`
}

func (f *questionFrontmatter) validate() error {
	if f.Type != "question" {
		return fmt.Errorf("%w: want: %q, got: %q", ErrWrongFrontmatterType, "question", f.Type)
	}
	if f.Answer == "" && f.SealedAnswer == "" && f.GenerateQuestions == "" {
		return fmt.Errorf("no answer found: %w", ErrNoFrontmatterAnswer)
	}
	if f.Answer != "" && f.SealedAnswer != "" {
		return fmt.Errorf("%w: sealed and unsealed answer found, only one allowed", ErrInvalidFrontmatter)
	}
	if f.GenerateQuestions != "" && f.AnswerType != "single-choice" {
		return fmt.Errorf("%w: answers can only be generated for single-choice answer-type", ErrInvalidFrontmatter)
	}
	if f.GenerateQuestions != "" && f.Answer != "" {
		return fmt.Errorf("%w: answer field must be empty when generating questions", ErrInvalidFrontmatter)
	}
	if f.GenerateQuestions != "" && f.SealedAnswer != "" {
		return fmt.Errorf("%w: sealed-answer field must be empty when generating questions", ErrInvalidFrontmatter)
	}
	return validateGeneratedQuestions(f.GenerateQuestions)
}

func (f *questionFrontmatter) getAnswer(privateKey string) (Answer, error) {
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

func (f *questionFrontmatter) Seal(publicKey string) error {
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

func (f *questionFrontmatter) Unseal(privateKey string) error {
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

func validateGeneratedQuestions(gen string) error {
	if gen == "" {
		return nil
	}
	questions := splitTrim(gen)
	if len(questions) == 1 && questions[0] != "all" {
		return fmt.Errorf("%w: more than one question must be generated for generate-questions", ErrInvalidFrontmatter)
	}
	if len(questions) != 1 {
		for _, q := range questions {
			if err := validateSingle(q); err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidFrontmatter, err)
			}
		}
	}
	return nil
}
