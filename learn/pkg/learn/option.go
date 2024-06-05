package learn

// Option is used on QuestionModel creation to set optional parameters.
type Option func(configurableModel)

type configurableModel interface {
	setPrivateKey(string)
	setIgnoreSealed()
	setRawMD(rawFrontmatter string, rawMD string)
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

// WithRawMD stores given frontmatter and markdown Contents, useful if
// Markdown file has already been read.
func WithRawMD(rawFrontmatter string, rawMD string) Option {
	return func(m configurableModel) {
		m.setRawMD(rawFrontmatter, rawMD)
	}
}
