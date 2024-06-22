package learn

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestGenerateQuestionSet(t *testing.T) {
	questions := questionsByDifficulty{
		"easy": []*QuestionModel{
			{configurableModel: &configurableModel{filename: "easy1.md"}},
			{configurableModel: &configurableModel{filename: "easy2.md"}},
			{configurableModel: &configurableModel{filename: "easy3.md"}},
			{configurableModel: &configurableModel{filename: "easy4.md"}},
		},
		"medium": []*QuestionModel{
			{configurableModel: &configurableModel{filename: "medium1.md"}},
			{configurableModel: &configurableModel{filename: "medium2.md"}},
			{configurableModel: &configurableModel{filename: "medium3.md"}},
			{configurableModel: &configurableModel{filename: "medium4.md"}},
		},
	}
	composition := []DifficultyCount{
		{Difficulty: "easy", Count: 2},
		{Difficulty: "medium", Count: 1},
		{Difficulty: "easy", Count: 2},
		{Difficulty: "medium", Count: 3},
	}

	repeats := 10
	repeatsByFilename := map[string]int{}

	for range repeats {
		questionSet := GenerateQuestionSet(questions, composition)
		for _, question := range questionSet {
			repeatsByFilename[question.Filename()]++
		}
	}
	assert.Equal(t, 8, len(repeatsByFilename), "%#v", repeatsByFilename)
	for filename, count := range repeatsByFilename {
		assert.Equal(t, repeats, count, filename)
	}
}
