package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DBHost       string `envconfig:"DB_HOST" default:"localhost"`
	DBUser       string `envconfig:"DB_USER" default:"root"`
	DBPassword   string `envconfig:"DB_PASSWORD" default:"root"`
	DBName       string `envconfig:"DB_NAME" default:"db_uninaquiz"`
	DBPort       string `envconfig:"DB_PORT" default:"5432"`
	JWTSecret    string `envconfig:"JWT_SECRET_KEY" default:"secret-key"`
	GeminiAPIKey string `envconfig:"GEMINI_API_KEY" default:""`
}

func LoadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}
