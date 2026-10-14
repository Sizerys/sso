package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	"sso/internal/grpc/auth"

	"google.golang.org/grpc"
)

type GRPCAPP struct {
	log        *slog.Logger
	GRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
) *GRPCAPP {
	GRPCServer := grpc.NewServer()

	auth.Register(GRPCServer)

	return &GRPCAPP{
		log:        log,
		GRPCServer: GRPCServer,
		port:       44044,
	}

}

func (g *GRPCAPP) Run() error {
	const op = "internal/app/grpc/app.Run"
	log := g.log.With(
		slog.String("op", op),
		slog.Int("port", g.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf("%d", g.port))
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	log.Info("server is running", slog.String("addres", l.Addr().String()))

	if err = g.GRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (g *GRPCAPP) Stop() {
	const op = "internal/app/grpc/app.Stop"

	log := g.log.With(
		slog.String("op", op),
		slog.Int("port", g.port),
	)

	g.GRPCServer.GracefulStop()

	log.Info("Server Stop", slog.String("op", op), slog.Int("port", g.port))
}

func (g GRPCAPP) MustRun() error {
	err := g.Run()
	if err != nil {
		panic("start grpc server error")
	}

	return nil
}
