package mappers

import (
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
	"github.com/google/uuid"
)

func ToQuizHistoryResponse(quiz entities.Quiz) *dto.QuizHistoryResponse {
	return &dto.QuizHistoryResponse{
		ID:         quiz.ID,
		Topic:      quiz.Topic,
		Difficulty: string(quiz.Difficulty),
		Score:      quiz.Score,
		Total:      quiz.Total,
		CreatedAt:  quiz.CreatedAt.UnixMilli(),
	}
}

func ToGetQuizResponse(quiz entities.Quiz) *dto.GetQuizResponse {
	questions := make([]dto.GetQuizQuestionResponse, 0, len(quiz.Questions))
	for _, q := range quiz.Questions {
		questions = append(questions, dto.GetQuizQuestionResponse{
			Text:         q.Text,
			Options:      q.Options,
			CorrectIndex: q.CorrectIndex,
			Explanation:  q.Explanation,
		})
	}
	return &dto.GetQuizResponse{
		ID:         quiz.ID,
		Topic:      quiz.Topic,
		Difficulty: string(quiz.Difficulty),
		Score:      quiz.Score,
		Total:      quiz.Total,
		Questions:  questions,
		CreatedAt:  quiz.CreatedAt.UnixMilli(),
	}
}

func ToGenerateQuizResponseFromEntity(quiz *entities.Quiz) *dto.GenerateQuizResponse {
	questions := make([]dto.QuizQuestion, 0, len(quiz.Questions))
	for _, q := range quiz.Questions {
		questions = append(questions, dto.QuizQuestion{
			Text:         q.Text,
			Options:      q.Options,
			CorrectIndex: q.CorrectIndex,
			Explanation:  q.Explanation,
		})
	}
	return &dto.GenerateQuizResponse{
		ID:         quiz.ID,
		Topic:      quiz.Topic,
		Difficulty: string(quiz.Difficulty),
		Total:      quiz.Total,
		Questions:  questions,
	}
}

// ToGenerateQuizEntity creates a Quiz entity from AI-generated questions, ready to be persisted.
func ToGenerateQuizEntity(quizID, userID, topic, difficulty string, aiQuestions []dto.QuizQuestion) *entities.Quiz {
	questions := make([]entities.QuizQuestion, 0, len(aiQuestions))
	for i, q := range aiQuestions {
		questions = append(questions, entities.QuizQuestion{
			ID:           uuid.New().String(),
			QuizID:       quizID,
			Position:     i,
			Text:         q.Text,
			Options:      q.Options,
			CorrectIndex: q.CorrectIndex,
			Explanation:  q.Explanation,
		})
	}
	return &entities.Quiz{
		ID:         quizID,
		UserID:     userID,
		Topic:      topic,
		Difficulty: entities.TQuizDifficulty(difficulty),
		Score:      0,
		Total:      len(aiQuestions),
		Questions:  questions,
		CreatedAt:  time.Now(),
	}
}
