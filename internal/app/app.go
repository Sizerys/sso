package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"time"
)

type APP struct {
	GRPCServer *grpcapp.GRPCAPP
}

func New(
	log *slog.Logger,
	port int,
	storage string,
	tokenTTL time.Duration,
) *APP {
	grpcApp := grpcapp.New(log, port)

	return &APP{
		GRPCServer: grpcApp,
	}
}
