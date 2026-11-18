package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/logger"
	"sso/internal/storage"
	"syscall"

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
	log.Info("config loaded", slog.String("storage_path", cfg.StoragePath))

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

	application := app.New(log, cfg.GRPC.Port, storagePostgreSQL, cfg.TokenTTL, cfg.RefreshTokenTTL)

	go func() {
		application.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	log.Info("gprpc server stoped")

	application.GRPCServer.Stop()

}
