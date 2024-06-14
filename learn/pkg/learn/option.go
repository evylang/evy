package learn

// Option is used on QuestionModel creation to set optional parameters.
type Option func(*configurableModel)

type configurableModel struct {
	filename       string
	privateKey     string
	ignoreSealed   bool
	rawFrontmatter string
	rawMD          string
	cache          map[string]model
}

func (m *configurableModel) Filename() string {
	return m.filename
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

func (m *configurableModel) setCache(cache map[string]model) {
	m.cache = cache
}

func newConfigurableModel(filename string, options []Option) *configurableModel {
	m := &configurableModel{
		filename: filename,
		cache:    map[string]model{},
	}
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

func withCache(cache map[string]model) Option {
	return func(m *configurableModel) {
		m.setCache(cache)
	}
}

func newOptions(ignoreSealed bool, privateKey string, cache map[string]model) []Option {
	var options []Option
	if ignoreSealed {
		options = append(options, WithIgnoreSealed())
	}
	if !ignoreSealed && privateKey != "" {
		options = append(options, WithPrivateKey(privateKey))
	}
	if cache != nil {
		options = append(options, withCache(cache))
	}
	return options
}
