package learn

// Option is used on QuestionModel creation to set optional parameters.
type Option func(configurableModel)

type configurableModel interface {
	setPrivateKey(string)
	setIgnoreSealed()
}

// WithPrivateKey sets privateKey and all follow-up method invocations attempt
// to unseal sealed answers.
func WithPrivateKey(privateKey string) Option {
	return func(m configurableModel) {
		m.setPrivateKey(privateKey)
	}
}

// WithIgnoreSealed sets explicit flag to ignore all sealed question.
func WithIgnoreSealed() Option {
	return func(m configurableModel) {
		m.setIgnoreSealed()
	}
}
