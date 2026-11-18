package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string        `yaml:"env" env-default:"local"`
	StoragePath     string        `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	GRPC            GRPC          `yaml:"grpc" env-required:"true"`
	TokenTTL        time.Duration `yaml:"token_ttl" env-default:"0.25h"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"720h"`
}

type GRPC struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

func MustLoad() *Config {

	configPath, storagePath := fetchPath()

	if configPath == "" {
		panic("Config path is required")
	}

	if storagePath == "" {
		panic("Storage path is required")
	}

	return MustLoadPath(configPath, storagePath)

}

func MustLoadPath(configPath, storagePath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("File config does not exist %s", configPath)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatal("does not read config")
	}

	if storagePath != "" {
		config.StoragePath = storagePath
	}

	if config.StoragePath == "" {
		log.Fatal("storage path is required (YAML, STORAGE_PATH env, or -storage-path flag)")
	}

	return &config
}

func fetchPath() (string, string) {
	var configPath, storagePath string
	flag.StringVar(&configPath, "config-path", "", "path to config")
	flag.StringVar(&storagePath, "storage-path", "", "path to storage")

	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	if storagePath == "" {
		storagePath = os.Getenv("STORAGE_PATH")
	}

	return configPath, storagePath
}
