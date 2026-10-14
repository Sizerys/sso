package auth

import (
	"context"
	ssov1 "sso/protos/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const EmptyID = 0

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}

func Register(gRPCServer *grpc.Server) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if req.GetUserName() == "" {
		return nil, status.Error(codes.InvalidArgument, "user name required ")
	}

	if req.GetUserPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "user password required")
	}

	if req.GetServiceId() == EmptyID {
		return nil, status.Error(codes.InvalidArgument, "service id required")
	}

	// Убрать первый nil
	return nil, nil
}

func (s *serverAPI) Logout(
	ctx context.Context,
	req *ssov1.LogoutRequest,
) (*ssov1.LogoutResponse, error) {
	panic("TO DO LOGOUT")
}
