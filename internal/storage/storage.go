package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound         = errors.New("User not found")
	ErrInvalidCredentials   = errors.New("Invalid credentials")
	ErrUserExist            = errors.New("User already exist")
	ErrRefreshTokenNotFound = errors.New("Refresh token not found")
	ErrSecretNotFound       = errors.New("Secrets not found")
	ErrAppNotFound          = errors.New("App not found")
)

type PostgreSQLWrapper struct {
	pool *pgxpool.Pool
}

func NewPostgreSQLWrapper(pool *pgxpool.Pool) *PostgreSQLWrapper {
	return &PostgreSQLWrapper{pool: pool}
}

func (p *PostgreSQLWrapper) Ping(ctx context.Context) error {
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("cannot connect to database %w", err)
	}
	return nil
}

func (p *PostgreSQLWrapper) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	return SaveUser(ctx, p.pool, email, passHash)
}

func (p *PostgreSQLWrapper) UserPassHash(ctx context.Context, email string) ([]byte, error) {
	return UserPassHash(ctx, p.pool, email)
}

func (p *PostgreSQLWrapper) AppSecret(ctx context.Context, name string) (string, error) {
	return AppSecret(ctx, p.pool, name)
}

func (p *PostgreSQLWrapper) UserByID(ctx context.Context, storage *pgxpool.Pool,
	email string) (models.User, error) {
	return UserByID(ctx, storage, email)
}

func (p *PostgreSQLWrapper) App(ctx context.Context, appID int32) (models.App, error) {
	return App(ctx, p.pool, appID)
}

func SaveUser(ctx context.Context, storage *pgxpool.Pool, email string, passHash []byte) (int64, error) {
	const op = `internal/storage/storage.SaveUser`
	const sqlQuery = "INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id;"
	var id int64
	if err := storage.QueryRow(ctx, sqlQuery, email, passHash).Scan(&id); err != nil {
		return 0, fmt.Errorf("сannot execute sql query %s %w", op, err)
	}

	return id, nil

}

func UserPassHash(ctx context.Context, storage *pgxpool.Pool, email string) ([]byte, error) {
	const op = `internal/storage/storage.UserPassHash`
	const sqlQuery = `SELECT pass_hash FROM users WHERE email = $1`
	var passHash []byte
	if err := storage.QueryRow(ctx, sqlQuery, email).Scan(&passHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s %s %w", op, ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("сannot execute sql query %s %w", op, err)
	}
	return passHash, nil

}

func AppSecret(ctx context.Context, storage *pgxpool.Pool, name string) (string, error) {
	const op = `internal/storage/storage.AppSecret`
	var secret string
	const sqlQuery = `SELECT secret FROM apps WHERE name = $1`
	if err := storage.QueryRow(ctx, sqlQuery, name).Scan(&secret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s %s %w", op, ErrSecretNotFound, err)
		}
		return "", fmt.Errorf("сannot execute sql query %x %w", op, err)
	}

	return secret, nil
}

func App(ctx context.Context, storage *pgxpool.Pool, appID int32) (models.App, error) {
	const op = `internal/storage/storage.App`
	var app models.App
	const sqlQuery = `SELECT * FROM apps WHERE id = $1`
	if err := storage.QueryRow(ctx, sqlQuery, appID).Scan(&app.ID, &app.Name, &app.Secret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s %s %w", op, ErrAppNotFound, err)
		}
		return models.App{}, fmt.Errorf("сannot execute sql query %s %w", op, err)
	}
	return app, nil
}

func UserByID(ctx context.Context, storage *pgxpool.Pool, email string) (models.User, error) {
	const op = `internal/storage/storage.UserID`
	const sqlQuery = `SELECT * FROM users WHERE email = $1`
	var user models.User
	if err := storage.QueryRow(ctx, sqlQuery).Scan(&user.ID, &user.Email, &user.PassHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s %s %w", op, ErrUserNotFound, err)
		}
		return models.User{}, fmt.Errorf("cannot execute sql query %s %w", op, err)
	}
	return user, nil
}

func Connect(ctx context.Context, dbPath string) (*pgxpool.Pool, error) {
	const op = `internal/storage/storage.Connect`
	pool, err := pgxpool.New(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new database connection %s %w", op, err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("cannot connect to database %w", err)
	}

	return pool, nil

}
