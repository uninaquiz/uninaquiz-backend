package dto

type QuizQuestion struct {
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correctIndex"`
	Explanation  string   `json:"explanation"`
}

type QuizHistoryResponse struct {
	ID         string `json:"id"`
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
	Score      int    `json:"score"`
	Total      int    `json:"total"`
	CreatedAt  int64  `json:"createdAt"`
}

type GenerateQuizResponse struct {
	ID         string         `json:"id"`
	Topic      string         `json:"topic"`
	Difficulty string         `json:"difficulty"`
	Total      int            `json:"total"`
	Questions  []QuizQuestion `json:"questions"`
}

type GetQuizQuestionResponse struct {
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correctIndex"`
	Explanation  string   `json:"explanation"`
}

type GetQuizResponse struct {
	ID         string                    `json:"id"`
	Topic      string                    `json:"topic"`
	Difficulty string                    `json:"difficulty"`
	Score      int                       `json:"score"`
	Total      int                       `json:"total"`
	Questions  []GetQuizQuestionResponse `json:"questions"`
	CreatedAt  int64                     `json:"createdAt"`
}
