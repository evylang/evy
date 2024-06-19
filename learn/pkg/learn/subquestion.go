package learn

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"strings"

	"golang.org/x/tools/txtar"
	"rsc.io/markdown"
)

func newSubQuestions(m *QuestionModel) ([]*QuestionModel, error) {
	if m.Frontmatter.GenerateQuestions == "" {
		return nil, nil
	}
	txtarContent := m.Question.(*txtarContent)
	archive := txtarContent.archive
	genSubQuestion := map[string]bool{}

	for _, question := range splitTrim(m.Frontmatter.GenerateQuestions) {
		genSubQuestion[question] = true
	}

	base := baseNoExt(m.Filename())
	hashed, err := hashedFilenames(base, archive)
	if err != nil {
		return nil, err
	}
	filenameNoExt := strings.TrimSuffix(m.Filename(), ".md")
	subQuestions := make([]*QuestionModel, 0, len(archive.Files))
	for _, file := range archive.Files {
		if !strings.HasSuffix(file.Name, ".evy") {
			continue
		}
		genQuestion := baseNoExt(file.Name)
		if !genSubQuestion["all"] && !genSubQuestion[genQuestion] {
			continue
		}
		filename := filenameNoExt + "-" + hashed[file.Name] + ".md"
		sub := newSubQuestion(m, genQuestion, filename, file.Data, txtarContent.ResultType)
		subQuestions = append(subQuestions, sub)
	}

	return subQuestions, nil
}

func newSubQuestion(m *QuestionModel, question, filename string, txtarEvyFile []byte, resultType ResultType) *QuestionModel {
	opts := newOptions(m.ignoreSealed, m.privateKey, m.cache)
	model := &QuestionModel{
		embeds:            map[markdown.Block]Renderer{},
		configurableModel: newConfigurableModel(filename, opts),
	}
	model.Doc = m.Doc
	model.ResultType = m.ResultType
	model.AnswerChoices = m.AnswerChoices
	model.answerBlock = m.answerBlock

	fm := *m.Frontmatter
	fm.Answer = question
	fm.GenerateQuestions = ""
	model.Frontmatter = &fm
	model.parentQuestion = m
	questionRenderer := newRendererFromEvyBytes(txtarEvyFile, resultType)
	model.Question = questionRenderer
	model.embeds = maps.Clone(m.embeds)
	for block, renderer := range model.embeds {
		if renderer == m.Question {
			// Update txtarContent renderer to specific txtar file, e.g.
			// `-- a.evy --` for which we are generating a new sub question.
			model.embeds[block] = questionRenderer
		}
	}
	return model
}

func hashedFilenames(modelFilename string, archive *txtar.Archive) (map[string]string, error) {
	filenames := map[string]string{}
	for _, file := range archive.Files {
		if !strings.HasSuffix(file.Name, ".evy") {
			continue
		}
		if err := validateSingle(baseNoExt(file.Name)); err != nil {
			return nil, fmt.Errorf("%w: %s", err, file.Name)
		}
		if _, ok := filenames[file.Name]; ok {
			return nil, fmt.Errorf("%w: duplicate filename %q inside txtar file %q", ErrInconsistentMdoel, modelFilename, file.Name)
		}
		h := sha256.New()
		h.Write([]byte(modelFilename + "-" + file.Name))
		filenames[file.Name] = hex.EncodeToString(h.Sum(nil))
	}
	charLength := 1
	for {
		shortFilenames, ok := shortenHashedFilenames(filenames, charLength)
		if ok {
			filenames = shortFilenames
			break
		}
		charLength++
	}
	return filenames, nil
}

func shortenHashedFilenames(filenames map[string]string, charLength int) (map[string]string, bool) {
	shortFilenames := map[string]bool{}
	result := map[string]string{}
	for k, filename := range filenames {
		shortFilename := filename[:charLength]
		if shortFilenames[shortFilename] {
			return nil, false
		}
		shortFilenames[shortFilename] = true
		result[k] = shortFilename
	}
	return result, true
}
