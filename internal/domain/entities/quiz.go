package entities

import (
	"fmt"
	"time"
)

type TQuizDifficulty string

const (
	DifficultyEasy   TQuizDifficulty = "easy"
	DifficultyMedium TQuizDifficulty = "medium"
	DifficultyHard   TQuizDifficulty = "hard"
)

type Quiz struct {
	ID         string
	UserID     string
	Topic      string
	Difficulty TQuizDifficulty
	Score      int
	Total      int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (qd TQuizDifficulty) String() string {
	return string(qd)
}

func ParseDifficulty(difficulty string) (TQuizDifficulty, error) {
	switch TQuizDifficulty(difficulty) {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return TQuizDifficulty(difficulty), nil
	}
	return "", fmt.Errorf("invalid difficulty: %s", difficulty)
}
