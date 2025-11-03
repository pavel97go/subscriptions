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
	AppPort string `yaml:"app_port"`
	DB      struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"password"`
		Name string `yaml:"name"`
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
		return cfg
	}
	_ = godotenv.Load()

	portStr := getenv("DB_PORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("invalid DB_PORT: %v", err)
	}

	cfg.AppPort = getenv("APP_PORT", "8080")
	cfg.DB.Host = getenv("DB_HOST", "localhost")
	cfg.DB.Port = port
	cfg.DB.User = getenv("DB_USER", "user")
	cfg.DB.Pass = getenv("DB_PASSWORD", "password")
	cfg.DB.Name = getenv("DB_NAME", "subscriptions_db")

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.Name,
	)
}

func getenv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
