-- +goose Up
CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE IF NOT EXISTS tb_quizzes (
id VARCHAR PRIMARY KEY DEFAULT gen_random_uuid(),
user_id VARCHAR NOT NULL,
topic VARCHAR(255) NOT NULL,
difficulty difficulty_level NOT NULL,
score INT NOT NULL,
total INT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES tb_users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS tb_quizzes;
DROP TYPE  IF EXISTS difficulty_level;