package auth

import (
	"context"
	"errors"
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
		rt string,
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
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "user email required ")
	}

	if req.GetUserPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "user password required")
	}

	if req.GetAppId() == EmptyID {
		return nil, status.Error(codes.InvalidArgument, "service id required")
	}

	pairtoken, err := s.auth.Login(ctx, req.GetEmail(), req.GetUserPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, er.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "login error")
	}

	return &ssov1.LoginResponse{
		AccessToken:  pairtoken.Acces,
		RefreshToken: pairtoken.Refresh,
	}, nil
}

func (s *serverAPI) Logout(
	ctx context.Context,
	req *ssov1.LogoutRequest,
) (*ssov1.LogoutResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	if err := s.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, status.Error(codes.Internal, "logout failed")
	}

	return &ssov1.LogoutResponse{}, nil
}
