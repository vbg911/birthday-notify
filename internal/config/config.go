package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env             string `yaml:"env" env-default:"local"`
	PostgresConnect string `yaml:"postgres_connect" env:"CONNECT_URL" env-required:"true"`
	HTTPServer      `yaml:"http_server"`
	SMTPServer      `yaml:"smtp_server"`
}

type HTTPServer struct {
	FiberAddress string        `yaml:"fiber_address" env-default:"localhost:8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"4s"`
}

type SMTPServer struct {
	SMTPAddress string `yaml:"smtp_address" env-required:"true"`
	SMTPPort    string `yaml:"smtp_port" env-required:"true"`
	From        string `yaml:"from" env-required:"true"`
	Password    string `yaml:"password" env-required:"true" env:"SMTP_PASS"`
}

func MustReadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

func GetJWTSecret() string {
	secret := os.Getenv("JWT")
	if secret == "" {
		return "secret"
	}
	return secret
}
