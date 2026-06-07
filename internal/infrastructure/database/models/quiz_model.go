package models

import (
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

type QuizModel struct {
	ID         string              `gorm:"column:id;primaryKey"`
	UserID     string              `gorm:"column:user_id;not null;index"`
	Topic      string              `gorm:"column:topic;not null"`
	Difficulty string              `gorm:"column:difficulty;not null"`
	Score      int                 `gorm:"column:score;not null"`
	Total      int                 `gorm:"column:total;not null"`
	Questions  []QuizQuestionModel `gorm:"foreignKey:QuizID;references:ID"`
	CreatedAt  time.Time           `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time           `gorm:"column:updated_at;autoUpdateTime"`
}

func (QuizModel) TableName() string { return "tb_quizzes" }

func (m *QuizModel) ToDomain() *entities.Quiz {
	questions := make([]entities.QuizQuestion, 0, len(m.Questions))
	for _, q := range m.Questions {
		questions = append(questions, q.ToDomain())
	}
	return &entities.Quiz{
		ID:         m.ID,
		UserID:     m.UserID,
		Topic:      m.Topic,
		Difficulty: entities.TQuizDifficulty(m.Difficulty),
		Score:      m.Score,
		Total:      m.Total,
		Questions:  questions,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func QuizToModel(q entities.Quiz) *QuizModel {
	questionModels := make([]QuizQuestionModel, 0, len(q.Questions))
	for _, qq := range q.Questions {
		questionModels = append(questionModels, *QuizQuestionToModel(qq))
	}
	return &QuizModel{
		ID:         q.ID,
		UserID:     q.UserID,
		Topic:      q.Topic,
		Difficulty: string(q.Difficulty),
		Score:      q.Score,
		Total:      q.Total,
		Questions:  questionModels,
		CreatedAt:  q.CreatedAt,
		UpdatedAt:  q.UpdatedAt,
	}
}
