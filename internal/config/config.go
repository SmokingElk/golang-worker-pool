package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogsPath       string       `yaml:"logs_path" binding:"required"`
	TimeoutSeconds int          `yaml:"timeout_seconds" binding:"required"`
	Worker         WorkerConfig `yaml:"worker" binding:"required"`
}

type WorkerConfig struct {
	QueueSize       int `yaml:"queue_size" binding:"required"`
	NumberOfWorkers int `yaml:"number_of_workers" binding:"required"`
}

func MustLoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	return &cfg
}
