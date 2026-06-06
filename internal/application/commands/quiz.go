package commands

type GenerateQuizCommand struct {
	Topic      string `json:"topic" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required,oneof=easy medium hard"`
}

type SaveQuizHistoryCommand struct {
	ID         string `json:"id" binding:"required"`
	Topic      string `json:"topic" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Score      int    `json:"score" binding:"min=0"`
	Total      int    `json:"total" binding:"required,min=1"`
	CreatedAt  int64  `json:"createdAt" binding:"required"`
}
