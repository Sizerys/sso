package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/logger"
	"sso/internal/storage"

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

	ctx := context.Background()

	pool, err := storage.Connect(ctx, cfg.StoragePath)
	if err != nil {
		log.Error("Cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	defer pool.Close()

	storagePostgreSQL := storage.NewPostgreSQLWrapper(pool)

	if err := storagePostgreSQL.Ping(ctx); err != nil {
		log.Error("Cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	application.GRPCServer.MustRun()

}
