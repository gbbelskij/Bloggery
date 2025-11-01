package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                string `yaml:"env" env-required:"true"`
	PostgresConnString string `yaml:"postgres_conn_string" env-required:"true"`
	JWTSecret          string `yaml:"jwt_secret"`
	HTTPServer         `yaml:"http_server"`
}

type HTTPServer struct {
	Address  string `yaml:"address" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("Config is nil")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("Config does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	return &cfg
}
