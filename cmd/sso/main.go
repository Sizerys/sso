package main

import (
	"log"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("/home/sterris/sso/.env"); err != nil {
		log.Fatal("Cannot load .env")
	}

	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("logger start")

	log.Debug("debug logger start")

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	application.GRPCServer.MustRun()

}
