package entities

import (
	"time"
)

type TQuizDifficulty string

type QuizQuestion struct {
	ID           string
	QuizID       string
	Position     int
	Text         string
	Options      []string
	CorrectIndex int
	Explanation  string
}

type Quiz struct {
	ID         string
	UserID     string
	Topic      string
	Difficulty TQuizDifficulty
	Score      int
	Total      int
	Questions  []QuizQuestion
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (qd TQuizDifficulty) String() string {
	return string(qd)
}
