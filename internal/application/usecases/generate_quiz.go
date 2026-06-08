package usecases

import (
	"context"
	"strings"
	"unicode"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/commands"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/dto"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/mappers"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/application/services"
	domainerrors "github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/errors"
	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/repositories"
	"github.com/google/uuid"
)

type GenerateQuizUseCase struct {
	aiService      services.IAIService
	quizRepository repositories.IQuizRepository
}

func NewGenerateQuizUseCase(aiService services.IAIService, quizRepository repositories.IQuizRepository) *GenerateQuizUseCase {
	return &GenerateQuizUseCase{
		aiService:      aiService,
		quizRepository: quizRepository,
	}
}

func (usc *GenerateQuizUseCase) Run(ctx context.Context, input commands.GenerateQuizCommand, userID string) (*dto.GenerateQuizResponse, error) {
	if err := validateTopic(input.Topic); err != nil {
		return nil, err
	}

	existentQuiz, err := usc.quizRepository.FindByUserTopicAndDifficulty(ctx, userID, input.Topic, input.Difficulty)
	if err != nil {
		return nil, err
	}
	if existentQuiz != nil {
		return mappers.ToGenerateQuizResponseFromEntity(existentQuiz), nil
	}

	aiQuestions, err := usc.aiService.GenerateQuiz(ctx, input.Topic, input.Difficulty)
	if err != nil {
		return nil, err
	}

	quizID := uuid.New().String()
	quizEntity := mappers.ToGenerateQuizEntity(quizID, userID, input.Topic, input.Difficulty, aiQuestions)

	created, err := usc.quizRepository.Create(ctx, *quizEntity)
	if err != nil {
		return nil, err
	}

	return &dto.GenerateQuizResponse{
		ID:         created.ID,
		Topic:      created.Topic,
		Difficulty: string(created.Difficulty),
		Total:      created.Total,
		Questions:  aiQuestions,
	}, nil
}

// Padrões de prompt injection a serem bloqueados
var injectionPatterns = []string{
	"ignore previous", "ignore all", "forget everything",
	"you are now", "act as", "pretend to be",
	"system:", "assistant:", "human:", "###",
	"jailbreak", "bypass", "override instruction",
	"<script", "javascript:", "eval(",
	"prompt injection", "ignore instructions",
	"disregard", "do not follow",
}

// validateTopic valida o tema contra tamanho, padrões de injeção e proporção de caracteres válidos.
func validateTopic(topic string) error {
	trimmed := strings.TrimSpace(topic)

	if len(trimmed) < 2 || len(trimmed) > 100 {
		return domainerrors.ErrInvalidTopic
	}

	lower := strings.ToLower(trimmed)
	for _, pattern := range injectionPatterns {
		if strings.Contains(lower, pattern) {
			return domainerrors.ErrInvalidTopic
		}
	}

	valid := 0
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' || r == '.' || r == ',' || r == '\'' || r == '(' || r == ')' {
			valid++
		}
	}
	if float64(valid)/float64(len(trimmed)) < 0.7 {
		return domainerrors.ErrInvalidTopic
	}

	return nil
}
