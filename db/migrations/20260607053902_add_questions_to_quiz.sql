-- +goose Up
CREATE TABLE IF NOT EXISTS tb_quiz_questions (
    id          VARCHAR PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id     VARCHAR NOT NULL,
    position    INT NOT NULL,
    text        TEXT NOT NULL,
    options     JSONB NOT NULL,
    correct_index INT NOT NULL,
    explanation TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (quiz_id) REFERENCES tb_quizzes(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS tb_quiz_questions;
