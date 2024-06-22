package learn

import (
	"path/filepath"
)

// Course represents the bare bones structure of a course as needed on the
// learn landing page and against progress tracking.
type Course struct {
	Name string `firestore:"name,omitempty" json:"name,omitempty"`
	// PartialID  is the last path segment of the directory in which course
	// Markdown file is located, as used in answerkey and AnswerPath.
	// It is used in the same way for Unit and Exercise.
	PartialID string          `firestore:"partialID,omitempty" json:"partialID,omitempty"`
	MaxScore  int             `firestore:"maxScore,omitempty" json:"maxScore,omitempty"`
	Units     map[string]Unit `firestore:"units,omitempty" json:"units,omitempty"`
	UnitOrder []string        `firestore:"unitOrder,omitempty" json:"unitOrder,omitempty"`
}

// Unit represents the a unit within a course catalog.
type Unit struct {
	Name          string              `firestore:"name,omitempty" json:"name,omitempty"`
	PartialID     string              `firestore:"partialID,omitempty" json:"partialID,omitempty"` // see Catalog.PartialID
	MaxScore      int                 `firestore:"maxScore,omitempty" json:"maxScore,omitempty"`
	Exercises     map[string]Exercise `firestore:"exercises,omitempty" json:"exercises,omitempty"`
	ExerciseOrder []string            `firestore:"exerciseOrder,omitempty" json:"exerciseOrder,omitempty"`
}

// Exercise represents an exercise, a quiz or a unittest within a course
// catalog.
type Exercise struct {
	PartialID   string            `firestore:"partialID,omitempty" json:"partialID,omitempty"` // see Catalog.PartialID
	Type        string            `firestore:"type,omitempty" json:"type,omitempty"`           // "exercise", "quiz" or "unittest"
	MaxScore    int               `firestore:"maxScore,omitempty" json:"maxScore,omitempty"`
	Composition []DifficultyCount `firestore:"composition,omitempty" json:"composition,omitempty"`
}

const exerciseScore = 10

// NewCourseCatalog creates a course catalog from a course model.
func NewCourseCatalog(courseModel *CourseModel) Course {
	course := Course{
		Name:      courseModel.Name(),
		PartialID: filepath.Base(filepath.Dir(courseModel.Filename())),
		Units:     map[string]Unit{},
	}
	for _, unitModel := range courseModel.Units {
		unit := Unit{
			Name:      unitModel.Name(),
			PartialID: filepath.Base(filepath.Dir(unitModel.Filename())),
			Exercises: map[string]Exercise{},
		}
		for _, model := range unitModel.OrderedModels {
			exercise := newExerciseCatalogEntry(model)
			unit.Exercises[exercise.PartialID] = exercise
			unit.ExerciseOrder = append(unit.ExerciseOrder, exercise.PartialID)
			unit.MaxScore += exercise.MaxScore
		}
		course.Units[unit.PartialID] = unit
		course.UnitOrder = append(course.UnitOrder, unit.PartialID)
		course.MaxScore += unit.MaxScore
	}
	return course
}

func newExerciseCatalogEntry(model model) Exercise {
	exercise := Exercise{}
	switch m := model.(type) {
	case *ExerciseModel:
		exercise.PartialID = filepath.Base(filepath.Dir(m.Filename()))
		exercise.Composition = m.Frontmatter.Composition
		exercise.MaxScore = exerciseScore
		exercise.Type = "exercise"
	case *QuizModel:
		exercise.PartialID = baseNoExt(m.Filename())
		exercise.Composition = m.Frontmatter.Composition
		exercise.Type = "quiz"
	case *UnittestModel:
		exercise.PartialID = baseNoExt(m.Filename())
		exercise.Composition = m.Frontmatter.Composition
		exercise.Type = "unittest"
	}
	return exercise
}
