package learn

import (
	"bytes"
	"fmt"
	"math/rand"
	"slices"
	"strconv"

	"gopkg.in/yaml.v3"
)

// DifficultyCount represents an element of the exercise's, quiz' or
// unittest's question composition: How many questions of a certain
// difficulty should be picked as part of this exercise or quiz. The
// following composition specifies that the exercise should have 2 easy
// followed by 3 hard questions.
//
//	composition:
//	  - easy: 2
//	  - hard: 3
type DifficultyCount struct {
	Difficulty string
	Count      int
}

// UnmarshalYAML unmarshals a DifficultyCount from a YAML node. In the source
// YAML it expects a map with a single entry. The key is the difficulty and
// the value is the count, this makes it easy and tight to manually write
// YAML. UnmarshalYAML returns an error if the map contains more than entry
// or if the key is not a difficulty.
func (d *DifficultyCount) UnmarshalYAML(value *yaml.Node) error {
	var plain map[difficulty]int
	if err := value.Decode(&plain); err != nil {
		return fmt.Errorf("%w: cannot unmarshal composition element: %w", ErrInvalidFrontmatter, err)
	}
	if len(plain) != 1 {
		return fmt.Errorf("%w: composition element must contain exactly one element", ErrInvalidFrontmatter)
	}
	for difficulty, count := range plain {
		d.Difficulty = string(difficulty)
		d.Count = count
	}
	return nil
}

type questionsByDifficulty map[string][]*QuestionModel

func (m questionsByDifficulty) merge(other questionsByDifficulty) {
	filenames := m.questionFilenames()
	for difficulty, questions := range other {
		for _, question := range questions {
			if !filenames[question.Filename] {
				m[difficulty] = append(m[difficulty], question)
			}
		}
	}
}

func (m questionsByDifficulty) validate(composition []DifficultyCount) error {
	totalByDifficulty := map[string]int{}
	for _, c := range composition {
		totalByDifficulty[c.Difficulty] += c.Count
	}
	for difficulty, total := range totalByDifficulty {
		if len(m[difficulty]) < total {
			return fmt.Errorf("%w: not enough questions of difficulty %q, expected %d, got %d", ErrInconsistentMdoel, difficulty, total, len(m[difficulty]))
		}
	}
	return nil
}

func (m questionsByDifficulty) questionFilenames() map[string]bool {
	filenames := map[string]bool{}
	for _, questions := range m {
		for _, question := range questions {
			filenames[question.Filename] = true
		}
	}
	return filenames
}

func (m questionsByDifficulty) PrintHTML(buf *bytes.Buffer) {
	buf.WriteString("<h3>Question Pool</h3>\n")
	buf.WriteString("<table>\n")

	for _, difficulty := range validDifficulties {
		questions := m[difficulty]
		filesnames := make([]string, len(questions))
		for i, q := range questions {
			filesnames[i] = q.Filename
		}
		slices.Sort(filesnames)
		for _, filename := range filesnames {
			buf.WriteString("<tr>")
			buf.WriteString("<td >" + filename + "</td>")
			buf.WriteString("<td >" + difficulty + " " + difficultyAsChillies(difficulty) + "</td>")
			buf.WriteString("</tr>\n")
		}
	}
	buf.WriteString("</table>\n")
}

// GenerateQuestionSet returns a list of question IDs based on the
// composition: If the composition is defined as 3 easy and 2 hard questions,
// then this function will return a list of 3 easy and 2 hard question IDs.
// The IDs are the question base filenames without extension, so the are only unique
// within the context of this exercise.
// The generated question set will eventually be tracked and persisted for returning
// students on the internet. This means we need to generate the question set before
// we start the exercise.
func GenerateQuestionSet(questionsByDifficulty questionsByDifficulty, composition []DifficultyCount) []*QuestionModel {
	permByDifficulty := map[string][]int{}
	for difficulty, quesitons := range questionsByDifficulty {
		permByDifficulty[difficulty] = rand.Perm(len(quesitons))
	}
	difficultyIdx := map[string]int{}
	var result []*QuestionModel
	for _, c := range composition {
		// Track progress index difficultyIdx[c.Difficulty] as there could be
		// multiple sets of same difficulty, e.g.: 2 easy, 3 medium, 1 easy,...
		idx := difficultyIdx[c.Difficulty]
		difficultyIdx[c.Difficulty] += c.Count
		perm := permByDifficulty[c.Difficulty][idx : idx+c.Count]
		questions := questionsByDifficulty[c.Difficulty]
		for _, i := range perm {
			result = append(result, questions[i])
		}
	}
	return result
}

func printComposition(w *bytes.Buffer, composition []DifficultyCount) {
	w.WriteString("<h3>Question Composition</h3>\n")
	w.WriteString("<table>\n")
	for _, el := range composition {
		w.WriteString("<tr>")
		w.WriteString("<td >" + el.Difficulty + " " + difficultyAsChillies(el.Difficulty) + "</td>")
		w.WriteString("<td>" + strconv.Itoa(el.Count) + "</td>")
		w.WriteString("</tr>\n")
	}
	w.WriteString("</table>\n")
}

func difficultyAsChillies(difficulty string) string {
	switch difficulty {
	case "easy":
		return "🌶️"
	case "medium":
		return "🌶️🌶️"
	case "hard":
		return "🌶️🌶️🌶️"
	}
	return ""
}
