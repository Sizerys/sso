package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	sl "sso/internal/lib/error"
	"sso/internal/lib/jwt"
	"sso/internal/lib/refresh"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredit = errors.New("Invalid credentials")
)

type Auth struct {
	log             *slog.Logger
	saverUser       SaverUser
	userProvider    UserProvider
	appProvider     AppProvider
	rtProvider      RefreshTokenProvider
	accesTokenTTL   time.Duration
	refreshTokenTTL time.Duration
}

type SaverUser interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (int64, error)
}

type RefreshTokenProvider interface {
	SaveRefreshToken(ctx context.Context, userID int64,
		tokenHash []byte, expires time.Time) error
	RefreshToken(ctx context.Context, refreshToken []byte) (models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, userID int64) (int64, error)
	RevokeAllUsersRefreshToken(ctx context.Context) error
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, ID int64) (bool, error)
	UserByID(ctx context.Context, userID int64) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

func New(
	log *slog.Logger,
	saverUser SaverUser,
	userProvider UserProvider,
	appProvider AppProvider,
	rtProvider RefreshTokenProvider,
	accesTokenTTL time.Duration,
	refreshTokenTTL time.Duration,

) *Auth {
	return &Auth{
		log:             log,
		saverUser:       saverUser,
		userProvider:    userProvider,
		appProvider:     appProvider,
		rtProvider:      rtProvider,
		accesTokenTTL:   accesTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int32,
) (models.TokenPair, error) {
	const op = "internal/service/auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("name", email),
	)

	log.Info("Start login seccion")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User already exist")

			return models.TokenPair{}, fmt.Errorf("%s, %w", op, ErrInvalidCredit)
		}
		log.Error("User founding error", sl.Err(err))

		return models.TokenPair{}, fmt.Errorf("%s, %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return models.TokenPair{}, fmt.Errorf("%s: %w", op, ErrInvalidCredit)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		log.Error("Get app ID error")

		return models.TokenPair{}, fmt.Errorf("get app ID error %s %d %w", op, appID, err)
	}

	token, err := jwt.NewToken(user, app, a.accesTokenTTL)
	if err != nil {
		log.Error("Token generate error")

		return models.TokenPair{}, fmt.Errorf("token generate error %s, %s, %d, %w", op, email, app.ID, err)
	}

	refresh := refresh.RefreshToken(32)
	refreshHash := sha256.Sum256([]byte(refresh))

	if err := a.rtProvider.SaveRefreshToken(ctx, user.ID, refreshHash[:],
		time.Now().Add(a.refreshTokenTTL)); err != nil {
		log.Error("Fail save refresh token")
		return models.TokenPair{}, fmt.Errorf("fail save refresh token")
	}

	return models.TokenPair{
		Acces:   token,
		Refresh: refresh,
	}, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string, appID int32) (models.TokenPair, error) {
	const op = "internal/service/auth/auth.Refresh"
	log := a.log.With(
		slog.String("op:", op),
	)
	log.Debug("Start refresh session")
	refreshTokenHash := sha256.Sum256([]byte(refreshToken))
	rt, err := a.rtProvider.RefreshToken(ctx, refreshTokenHash[:])
	if valid := rt.IsValid(); valid != true {
		log.Warn("refresh token is not valid")
		slog.Bool("IsExpired", rt.IsExpired())
		slog.Bool("IsReovked", rt.IsRevoked())
		return models.TokenPair{}, fmt.Errorf("%s %w", op, storage.ErrInvalidCredentials)
	}
	if err != nil {
		if errors.Is(err, storage.ErrRefreshTokenNotFound) {
			log.Warn("Refresh token not found")
			return models.TokenPair{}, fmt.Errorf("%s %w", op, err)
		}
		log.Error("cannot get refresh token")
		return models.TokenPair{}, fmt.Errorf("%s %w", op, err)

	}

	user, err := a.userProvider.UserByID(ctx, rt.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found")
			return models.TokenPair{}, fmt.Errorf("%s %w", op, err)
		}
		log.Error("cannot get user")
		return models.TokenPair{}, fmt.Errorf("%s %w", op, err)
	}
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("App not found")
			return models.TokenPair{}, fmt.Errorf("%s %w", op, err)
		}
		log.Error("Cannot get app")
		return models.TokenPair{}, fmt.Errorf("cannot geta app %s %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.accesTokenTTL)
	if err != nil {
		log.Error("Token generate error")

		return models.TokenPair{}, fmt.Errorf("token generate error %s, %d, %w", op, app.ID, err)
	}

	return models.TokenPair{
		Acces:   token,
		Refresh: refreshToken,
	}, nil

}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "internal/service/auth.Register"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("Register new user seccion start")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Hashing password error")

		return 0, fmt.Errorf("Hashing password error %s %w", op, err)
	}

	id, err := a.saverUser.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			log.Warn("User already exist")

			return 0, fmt.Errorf("%s %w", op, err)
		}
		log.Error("Save user error", sl.Err(err))

		return 0, fmt.Errorf("%s %w", op, err)
	}

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, ID int64) (bool, error) {
	const op = "internal/service/auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("ID", ID),
	)

	isAdmin, err := a.userProvider.IsAdmin(ctx, ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User does not exist", sl.Err(err))

			return false, fmt.Errorf("%s %w", op, err)
		}
		log.Error("Is admin check error", sl.Err(err))

		return false, fmt.Errorf("isAdmin check error %s %w", op, err)
	}

	log.Info("Checked isAdmin", slog.Int64("user ID", ID), slog.Bool("isAdmin:", isAdmin))

	return isAdmin, nil

}
