package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DatabaseURL  string `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret    string `envconfig:"JWT_SECRET_KEY" default:"supersecret"`
	GeminiAPIKey string `envconfig:"GEMINI_API_KEY" required:"true"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}
