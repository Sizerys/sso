package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	GRPC        GRPC   `yaml:"grpc" env-requierd:"true"`
}

type GRPC struct {
	Port    int64     `yaml:"port" env-required:"true"`
	Timeout time.Time `yaml:"timeout" env-default:"5s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("Config path does not exist")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("File config does not exist %s", configPath)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatal("cannot read config")
	}

	return &config

}
