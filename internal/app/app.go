package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/service/auth"
	"sso/internal/storage"
	"time"
)

type APP struct {
	GRPCServer *grpcapp.GRPCAPP
}

func New(
	log *slog.Logger,
	port int,
	storage *storage.PostgreSQLWrapper,
	accesTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *APP {
	authSerivce := auth.New(log, storage, storage, storage, storage, accesTokenTTL, refreshTokenTTL)
	grpcApp := grpcapp.New(log, port, authSerivce)

	return &APP{
		GRPCServer: grpcApp,
	}
}
