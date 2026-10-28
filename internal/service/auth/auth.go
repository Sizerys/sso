package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/config"
	"sso/internal/domain/models"
	sl "sso/internal/lib/error"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredit = errors.New("Invalid credentials")
)

type Auth struct {
	log          *slog.Logger
	saverUser    SaverUser
	userProvider UserProvide
	appProvider  AppProvider
	tokenTTL     time.Duration
	
}

type SaverUser interface {
	SaveUser(
		email string,
		passHash []byte,
	) (int64, error)
}

type UserProvide interface {
	User(ctx context.Context, email string) (models.User, error)
	isAdmin(ctx context.Context, ID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

func New(
	log *slog.Logger,
	saverUser SaverUser,
	userProvider UserProvide,
	appProvider AppProvider,
	cfg config.Config,
) *Auth {
	return &Auth{
		log:          log,
		saverUser:    saverUser,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     cfg.TokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int64,
) (string, error) {
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

			return "", fmt.Errorf("%s, %w", op, &ErrInvalidCredit)
		}
		log.Error("User founding error", sl.Err(err))

		return "", fmt.Errorf("%s, %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, storage.ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		log.Error("Get app ID error")

		return "", fmt.Errorf("Get app ID error %s %d %w", op, appID, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		log.Error("Token generate error")

		return "", fmt.Errorf("Token generate error %s, %s, %d, %w", op, email, app.ID, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(
	context context.Context,
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

	id, err := a.saverUser.SaveUser(email, passHash)
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

func (a *Auth) isAdmin(ctx context.Context, ID int64) (bool, error) {
	const op = "internal/service/auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("ID", ID),
	)

	isAdmin, err := a.userProvider.isAdmin(ctx, ID)
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
