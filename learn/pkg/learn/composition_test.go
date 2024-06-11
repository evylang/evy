package learn

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestGenerateQuestionSet(t *testing.T) {
	questions := questionsByDifficulty{
		"easy": []*QuestionModel{
			{Filename: "easy1.md"},
			{Filename: "easy2.md"},
			{Filename: "easy3.md"},
			{Filename: "easy4.md"},
		},
		"medium": []*QuestionModel{
			{Filename: "medium1.md"},
			{Filename: "medium2.md"},
			{Filename: "medium3.md"},
			{Filename: "medium4.md"},
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
			repeatsByFilename[question.Filename]++
		}
	}
	assert.Equal(t, 8, len(repeatsByFilename), "%#v", repeatsByFilename)
	for filename, count := range repeatsByFilename {
		assert.Equal(t, repeats, count, filename)
	}
}
