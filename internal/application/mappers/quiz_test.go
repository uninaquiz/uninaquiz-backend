package mappers_test

import (
	"testing"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

func TestToQuizHistoryResponse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		input entities.Quiz
	}{
		{
			name: "all fields mapped correctly",
			input: entities.Quiz{
				ID:         "quiz-1",
				Topic:      "Mathematics",
				Difficulty: "easy",
				Score:      4,
				Total:      5,
				CreatedAt:  now,
			},
		},
		{
			name:  "zero-value quiz",
			input: entities.Quiz{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mappers.ToQuizHistoryResponse(tt.input)
			if resp == nil {
				t.Fatal("ToQuizHistoryResponse() returned nil")
			}
			if resp.ID != tt.input.ID {
				t.Errorf("ID = %v, want %v", resp.ID, tt.input.ID)
			}
			if resp.Topic != tt.input.Topic {
				t.Errorf("Topic = %v, want %v", resp.Topic, tt.input.Topic)
			}
			if resp.Difficulty != string(tt.input.Difficulty) {
				t.Errorf("Difficulty = %v, want %v", resp.Difficulty, tt.input.Difficulty)
			}
			if resp.Score != tt.input.Score {
				t.Errorf("Score = %v, want %v", resp.Score, tt.input.Score)
			}
			if resp.Total != tt.input.Total {
				t.Errorf("Total = %v, want %v", resp.Total, tt.input.Total)
			}
			if resp.CreatedAt != tt.input.CreatedAt.UnixMilli() {
				t.Errorf("CreatedAt = %v, want %v", resp.CreatedAt, tt.input.CreatedAt.UnixMilli())
			}
		})
	}
}

func TestToGenerateQuizEntity(t *testing.T) {
	aiQuestions := []dto.QuizQuestion{
		{Text: "Q1", Options: []string{"A", "B"}, CorrectIndex: 0, Explanation: "Exp1"},
		{Text: "Q2", Options: []string{"C", "D"}, CorrectIndex: 1, Explanation: "Exp2"},
	}

	tests := []struct {
		name       string
		quizID     string
		userID     string
		topic      string
		difficulty string
		questions  []dto.QuizQuestion
	}{
		{
			name:       "entity built correctly from AI questions",
			quizID:     "quiz-uuid-1",
			userID:     "user-uuid-1",
			topic:      "Physics",
			difficulty: "medium",
			questions:  aiQuestions,
		},
		{
			name:       "empty question list",
			quizID:     "quiz-uuid-2",
			userID:     "user-uuid-2",
			topic:      "Biology",
			difficulty: "hard",
			questions:  []dto.QuizQuestion{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := mappers.ToGenerateQuizEntity(tt.quizID, tt.userID, tt.topic, tt.difficulty, tt.questions)
			if entity == nil {
				t.Fatal("ToGenerateQuizEntity() returned nil")
			}
			if entity.ID != tt.quizID {
				t.Errorf("ID = %v, want %v", entity.ID, tt.quizID)
			}
			if entity.UserID != tt.userID {
				t.Errorf("UserID = %v, want %v", entity.UserID, tt.userID)
			}
			if entity.Topic != tt.topic {
				t.Errorf("Topic = %v, want %v", entity.Topic, tt.topic)
			}
			if string(entity.Difficulty) != tt.difficulty {
				t.Errorf("Difficulty = %v, want %v", entity.Difficulty, tt.difficulty)
			}
			if entity.Total != len(tt.questions) {
				t.Errorf("Total = %v, want %v", entity.Total, len(tt.questions))
			}
			if len(entity.Questions) != len(tt.questions) {
				t.Errorf("len(Questions) = %v, want %v", len(entity.Questions), len(tt.questions))
			}
			for i, q := range entity.Questions {
				if q.Position != i {
					t.Errorf("Questions[%d].Position = %v, want %v", i, q.Position, i)
				}
				if q.QuizID != tt.quizID {
					t.Errorf("Questions[%d].QuizID = %v, want %v", i, q.QuizID, tt.quizID)
				}
				if q.ID == "" {
					t.Errorf("Questions[%d].ID should not be empty", i)
				}
			}
		})
	}
}

func TestToGenerateQuizResponseFromEntity(t *testing.T) {
	quizWithQuestions := &entities.Quiz{
		ID:         "quiz-1",
		Topic:      "Chemistry",
		Difficulty: "medium",
		Total:      2,
		Questions: []entities.QuizQuestion{
			{ID: "qq1", Text: "Q1", Options: []string{"A", "B"}, CorrectIndex: 0, Explanation: "E1"},
			{ID: "qq2", Text: "Q2", Options: []string{"C", "D"}, CorrectIndex: 1, Explanation: "E2"},
		},
	}

	tests := []struct {
		name  string
		input *entities.Quiz
	}{
		{
			name:  "maps quiz with questions to generate response",
			input: quizWithQuestions,
		},
		{
			name:  "quiz with no questions",
			input: &entities.Quiz{ID: "quiz-2", Topic: "Art", Difficulty: "easy", Questions: []entities.QuizQuestion{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mappers.ToGenerateQuizResponseFromEntity(tt.input)
			if resp == nil {
				t.Fatal("ToGenerateQuizResponseFromEntity() returned nil")
			}
			if resp.ID != tt.input.ID {
				t.Errorf("ID = %v, want %v", resp.ID, tt.input.ID)
			}
			if len(resp.Questions) != len(tt.input.Questions) {
				t.Errorf("len(Questions) = %v, want %v", len(resp.Questions), len(tt.input.Questions))
			}
		})
	}
}
