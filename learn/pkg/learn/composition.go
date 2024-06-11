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
// followed by 3 hard questions and a final easy question.
//
//	composition:
//	  - easy: 2
//	  - hard: 3
//	  - easy: 1
type DifficultyCount struct {
	Difficulty string
	Count      int
}

// UnmarshalYAML unmarshals a DifficultyCount from a YAML node. In the source
// YAML it expects a map with a single entry. The key is the difficulty and
// the value is the count, this makes it easy and tight to manually write
// YAML. UnmarshalYAML returns an error if the map contains more than one
// entry or if the key is not a valid difficulty.
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

func (m questionsByDifficulty) PrintHTML(buf *bytes.Buffer) {
	buf.WriteString("<h3>Question Pool</h3>\n")
	buf.WriteString("<table>\n")

	for _, difficulty := range validDifficulties {
		questions := m[difficulty]
		filenames := make([]string, len(questions))
		for i, q := range questions {
			filenames[i] = q.Filename
		}
		slices.Sort(filenames)
		for _, filename := range filenames {
			buf.WriteString("<tr>")
			buf.WriteString("<td >" + filename + "</td>")
			buf.WriteString("<td >" + difficulty + " " + difficultyAsChillies(difficulty) + "</td>")
			buf.WriteString("</tr>\n")
		}
	}
	buf.WriteString("</table>\n")
}

// GenerateQuestionSet returns a randomized list of question IDs following the
// composition requirements: If the composition is defined as 3 easy and 2
// hard questions, then this function will return a list of 3 easy and 2 hard
// question IDs. The IDs are the question base filenames without extension,
// so they are only unique within the context of this exercise. No question
// ID may be repeated.
//
// The generated question set will eventually be tracked and persisted for returning
// students on the internet. This means we need to generate the question set before
// we start the exercise.
func GenerateQuestionSet(questionsByDifficulty questionsByDifficulty, composition []DifficultyCount) []*QuestionModel {
	permByDifficulty := map[string][]int{}
	for difficulty, quesitons := range questionsByDifficulty {
		permByDifficulty[difficulty] = rand.Perm(len(quesitons))
	}
	var result []*QuestionModel
	for _, c := range composition {
		// Consume the next N questions at the difficulty level.
		perm := permByDifficulty[c.Difficulty][:c.Count]
		permByDifficulty[c.Difficulty] = permByDifficulty[c.Difficulty][c.Count:]
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
