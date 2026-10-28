package storage

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound       = errors.New("User not found")
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrUserExist          = errors.New("User already exist")
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

func SaveUser(ctx context.Context, storage *pgxpool.Pool, email string, passHash []byte) (int64, error) {
	const sqlQuery = "INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id;"
	var id int64
	if err := storage.QueryRow(ctx, sqlQuery, email, passHash).Scan(&id); err != nil {
		return 0, fmt.Errorf("сannot execute sql query %w", err)
	}

	return id, nil

}

func UserPassHash(ctx context.Context, storage *pgxpool.Pool, email string) ([]byte, error) {
	const sqlQuery = `SELECT pass_hash FROM users WHERE email = $1`
	var passHash []byte
	if err := storage.QueryRow(ctx, sqlQuery, email).Scan(&passHash); err != nil {
		return nil, fmt.Errorf("сannot execute sql query %w", err)
	}
	return passHash, nil

}

func AppSecret(ctx context.Context, storage *pgxpool.Pool, name string) (string, error) {
	var secret string
	const sqlQuery = `SELECT secret FROM apps WHERE name = $1`
	if err := storage.QueryRow(ctx, sqlQuery, name).Scan(&secret); err != nil {
		return "", fmt.Errorf("сannot execute sql query %w", err)
	}

	return secret, nil
}

func SaveRefreshToken(ctx context.Context, storage *pgxpool.Pool,
	userID int64, appID int32, tokenHash string, expires time.Time) (int64, error) {
	var id int64
	const sqlQuery = `INSERT INTO refresh_tokens 
	(user_id, app_id, token_hash, created_at, expires_at, revoked_at) VALUES($1, $2, $3, NOW(), $4, NULL) RETURNING id;`
	if err := storage.QueryRow(ctx, sqlQuery, userID, appID, tokenHash, expires).Scan(&id); err != nil {
		return 0, fmt.Errorf("сannot execute sql query %w", err)
	}
	return id, nil
}

func RefreshToken(ctx context.Context, storage *pgxpool.Pool, userID int64) (*models.Token, error) {
	var token *models.Token
	const sqlQuery = `SELECT * FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL ORDER BY created_at DESC LIMIT 1;`

	if err := storage.QueryRow(ctx, sqlQuery, userID).Scan(&token); err != nil {
		return nil, fmt.Errorf("cannot execute sql query %w", err)
	}

	return token, nil
}

func Connect(ctx context.Context, dbPath string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new database connection %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("cannot connect to database %w", err)
	}

	return pool, nil

}
