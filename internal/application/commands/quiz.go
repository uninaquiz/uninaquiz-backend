package commands

type GenerateQuizCommand struct {
	Topic      string `json:"topic" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required,oneof=easy medium hard"`
}

type SaveQuizHistoryCommand struct {
	ID    string `json:"id" binding:"required"`
	Score int    `json:"score" binding:"min=0"`
}
