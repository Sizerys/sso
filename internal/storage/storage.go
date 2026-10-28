package storage

import (
	"context"
	"errors"
	"fmt"

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

func (p *PostgreSQLWrapper) GetUserPassHash(ctx context.Context, email string) ([]byte, error) {
	return GetUserPassHash(ctx, p.pool, email)
}

func (p *PostgreSQLWrapper) GetAppSecret(ctx context.Context, name string) (string, error) {
	return GetAppSecret(ctx, p.pool, name)
}

func SaveUser(ctx context.Context, storage *pgxpool.Pool, email string, passHash []byte) (int64, error) {
	const sqlQuery = "INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id;"
	var id int64
	if err := storage.QueryRow(ctx, sqlQuery, email, passHash).Scan(&id); err != nil {
		return 0, fmt.Errorf("Cannot execute sql query %w", err)
	}

	return id, nil

}

func GetUserPassHash(ctx context.Context, storage *pgxpool.Pool, email string) ([]byte, error) {
	const sqlQuery = `SELECT pass_hash FROM users WHERE email = $1`
	var passHash []byte
	if err := storage.QueryRow(ctx, sqlQuery, email).Scan(&passHash); err != nil {
		return nil, fmt.Errorf("Cannot execute sql query %w", err)
	}
	return passHash, nil

}

func GetAppSecret(ctx context.Context, storage *pgxpool.Pool, name string) (string, error) {
	var secret string
	const sqlQuery = `SELECT secret FROM apps WHERE name = $1`
	if err := storage.QueryRow(ctx, sqlQuery, name).Scan(&secret); err != nil {
		return "", fmt.Errorf("Cannot execute sql query %w", err)
	}

	return secret, nil
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
