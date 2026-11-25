package auth

import (
	"context"
	"errors"
	"fmt"
	er "sso/internal/domain/errors"
	"sso/internal/domain/models"
	ssov1 "sso/protos/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int32,
	) (models.TokenPair, error)
	Refresh(ctx context.Context,
		refreshToken string,
		appID int32,
	) (models.TokenPair, error)
	RegisterNewUser(
		context context.Context,
		email string,
		password string,
	) (int64, error)
	Logout(
		ctx context.Context,
		rt []byte,
	) error
	IsAdmin(ctx context.Context, ID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

const EmptyID = 0

func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
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

	pairtoken, err := s.auth.Login(ctx, req.GetUserName(), req.GetUserPassword(), req.GetServiceId())
	if err != nil {
		if errors.Is(err, er.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "login error")
	}

	return &ssov1.LoginResponse{
		AccesToken:   pairtoken.Acces,
		RefreshToken: pairtoken.Refresh,
	}, nil
}

func (s *serverAPI) Logout(
	ctx context.Context,
	req *ssov1.LogoutRequest,
) (*ssov1.LogoutResponse, error) {
	const op = `internal/grpc/auth/server.go.Logout`
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is empty")
	}
	if err := s.auth.Logout(ctx, []byte(req.RefreshToken)); err != nil {
		if errors.Is(err, er.ErrRefreshTokenNotFound) {
			return nil, fmt.Errorf("Cannot getting refresh token %s %w", op, err)
		}
		if errors.Is(err, er.ErrInvalidCredentials) {
			return nil, fmt.Errorf("refresh token is not valid")
		}
		if errors.Is(err, er.ErrRefreshTokenByIDNotFound) {
			return nil, fmt.Errorf("Cannot getting refresh token by id %s %w", op, err)
		}
		return nil, fmt.Errorf("Logout error %s %w", op, err)
	}
	return &ssov1.LogoutResponse{
		Succes: "succes",
	}, nil
}
