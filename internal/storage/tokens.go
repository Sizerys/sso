package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (p *PostgreSQLWrapper) SaveRefreshToken(ctx context.Context, userID int64, tokenHash []byte, expires time.Time) error {
	return SaveRefreshToken(ctx, p.pool, userID, tokenHash, expires)
}

func (p *PostgreSQLWrapper) RefreshToken(ctx context.Context, refreshToken []byte) (*models.RefreshToken, error) {
	return RefreshToken(ctx, p.pool, refreshToken)
}

func (p *PostgreSQLWrapper) RevokeRefreshToken(ctx context.Context, userID int64) (int64, error) {
	return RevokeRefreshToken(ctx, p.pool, userID)
}

func (p *PostgreSQLWrapper) RevokeAllUsersRefreshToken(ctx context.Context, storage *pgxpool.Pool) error {
	return RevokeAllUsersRefreshToken(ctx, p.pool)
}

func SaveRefreshToken(ctx context.Context, storage *pgxpool.Pool,
	userID int64, tokenHash []byte, expires time.Time) error {
	const op = `internal/storage/tokens.SaveRefreshToken`
	const sqlQuery = `INSERT INTO refresh_tokens 
	(user_id, app_id, token_hash, created_at, expires_at, revoked_at) VALUES($1, $2, NOW(), $3, NULL) RETURNING id;`
	_, err := storage.Exec(ctx, sqlQuery, userID, tokenHash, expires)
	if err != nil {
		return fmt.Errorf("cannot execute sql query %s %w", op, err)
	}
	return nil
}

func RefreshToken(ctx context.Context, storage *pgxpool.Pool, refreshToken []byte) (*models.RefreshToken, error) {
	const op = `internal/storage/tokens.RefreshToken`
	var token *models.RefreshToken
	const sqlQuery = `SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL ORDER BY created_at DESC LIMIT 1;`

	if err := storage.QueryRow(ctx, sqlQuery, refreshToken).Scan(&token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s %s %w", op, ErrRefreshTokenNotFound, err)
		}
		return nil, fmt.Errorf("cannot execute sql query %s %w", op, err)
	}

	return token, nil
}

func RevokeRefreshToken(ctx context.Context, storage *pgxpool.Pool, userID int64) (int64, error) {
	const op = `internal/storage/tokens.RevokeRefreshToken`
	const sqlQuery = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL;`
	var id int64
	if err := storage.QueryRow(ctx, sqlQuery, userID).Scan(&id); err != nil {
		return 0, fmt.Errorf("Cannot execute sql query %s %w", op, err)
	}
	return id, nil
}

func RevokeAllUsersRefreshToken(ctx context.Context, storage *pgxpool.Pool) error {
	const op = `internal/storage/tokens.RevokeAllUsersRefreshToken`
	const sqlQuery = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE revoked_at IS NULL;`
	_, err := storage.Exec(ctx, sqlQuery)
	if err != nil {
		return fmt.Errorf("cannot execute sql query %s %w", op, err)
	}
	return nil
}
