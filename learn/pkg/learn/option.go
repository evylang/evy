package learn

// Option is used on QuestionModel creation to set optional parameters.
type Option func(*configurableModel)

type configurableModel struct {
	privateKey     string
	ignoreSealed   bool
	rawFrontmatter string
	rawMD          string
}

func (m *configurableModel) setPrivateKey(privateKey string) {
	m.privateKey = privateKey
}

func (m *configurableModel) setIgnoreSealed() {
	m.ignoreSealed = true
}

func (m *configurableModel) setRawMD(rawFrontmatter string, rawMD string) {
	m.rawFrontmatter = rawFrontmatter
	m.rawMD = rawMD
}

func newConfigurableModel(options []Option) *configurableModel {
	m := &configurableModel{}
	for _, opt := range options {
		opt(m)
	}
	return m
}

// WithPrivateKey sets privateKey and all follow-up method invocations attempt
// to unseal sealed answers.
func WithPrivateKey(privateKey string) Option {
	return func(m *configurableModel) {
		m.setPrivateKey(privateKey)
	}
}

// WithIgnoreSealed sets explicit flag to ignore all sealed question.
func WithIgnoreSealed() Option {
	return func(m *configurableModel) {
		m.setIgnoreSealed()
	}
}

// WithRawMD stores given frontmatter and markdown Contents, useful if
// Markdown file has already been read.
func WithRawMD(rawFrontmatter string, rawMD string) Option {
	return func(m *configurableModel) {
		m.setRawMD(rawFrontmatter, rawMD)
	}
}
