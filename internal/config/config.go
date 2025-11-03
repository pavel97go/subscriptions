package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AppPort  string `yaml:"app_port"`
	LogLevel string `yaml:"log_level"`
	DB       struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"password"`
		Name string `yaml:"name"`
		DSN  string `yaml:"dsn"`
	} `yaml:"db"`
}

func Load() *Config {
	cfg := &Config{}
	if _, err := os.Stat("config.yaml"); err == nil {
		f, err := os.ReadFile("config.yaml")
		if err != nil {
			log.Fatalf("error reading config.yaml: %v", err)
		}
		if err := yaml.Unmarshal(f, cfg); err != nil {
			log.Fatalf("error parsing YAML: %v", err)
		}
	}
	_ = godotenv.Load()

	if v := os.Getenv("APP_PORT"); v != "" {
		cfg.AppPort = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DB.Host = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DB.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DB.Pass = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DB.Name = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.DB.DSN = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.DB.Port = port
		}
	}
	if cfg.DB.DSN == "" {
		cfg.DB.DSN = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.DB.User, cfg.DB.Pass, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name,
		)
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return cfg
}
