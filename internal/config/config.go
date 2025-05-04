package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	ServerPort      string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-fs.json"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

var cfg Config

func LoadConfig() (*Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "address and port to run server")
	flag.StringVar(&cfg.ServerPort, "a", cfg.ServerPort, "address and port for result url")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "the full name of the file where the data is saved")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "PostgresSQL DSN")
	flag.Parse()
	return &cfg, nil
}
