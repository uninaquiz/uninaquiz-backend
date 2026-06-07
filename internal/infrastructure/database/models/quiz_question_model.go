package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

// StringSlice is a custom type for serializing []string as JSONB in PostgreSQL.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte for StringSlice, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

type QuizQuestionModel struct {
	ID           string      `gorm:"column:id;primaryKey"`
	QuizID       string      `gorm:"column:quiz_id;not null;index"`
	Position     int         `gorm:"column:position;not null"`
	Text         string      `gorm:"column:text;not null"`
	Options      StringSlice `gorm:"column:options;type:jsonb;not null"`
	CorrectIndex int         `gorm:"column:correct_index;not null"`
	Explanation  string      `gorm:"column:explanation;not null;default:''"`
}

func (QuizQuestionModel) TableName() string { return "tb_quiz_questions" }

func (m *QuizQuestionModel) ToDomain() entities.QuizQuestion {
	return entities.QuizQuestion{
		ID:           m.ID,
		QuizID:       m.QuizID,
		Position:     m.Position,
		Text:         m.Text,
		Options:      []string(m.Options),
		CorrectIndex: m.CorrectIndex,
		Explanation:  m.Explanation,
	}
}

func QuizQuestionToModel(q entities.QuizQuestion) *QuizQuestionModel {
	return &QuizQuestionModel{
		ID:           q.ID,
		QuizID:       q.QuizID,
		Position:     q.Position,
		Text:         q.Text,
		Options:      StringSlice(q.Options),
		CorrectIndex: q.CorrectIndex,
		Explanation:  q.Explanation,
	}
}
