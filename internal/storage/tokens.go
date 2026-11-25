package storage

import (
	"context"
	"errors"
	"fmt"
	er "sso/internal/domain/errors"
	"sso/internal/domain/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (p *PostgreSQLWrapper) SaveRefreshToken(ctx context.Context, userID int64, tokenHash []byte, expires time.Time) error {
	return SaveRefreshToken(ctx, p.pool, userID, tokenHash, expires)
}

func (p *PostgreSQLWrapper) RefreshToken(ctx context.Context, refreshToken []byte) (models.RefreshToken, error) {
	return RefreshToken(ctx, p.pool, refreshToken)
}

func (p *PostgreSQLWrapper) RevokeRefreshToken(ctx context.Context, userID int64) (int64, error) {
	return RevokeRefreshToken(ctx, p.pool, userID)
}

func (p *PostgreSQLWrapper) RevokeAllUsersRefreshToken(ctx context.Context) error {
	return RevokeAllUsersRefreshToken(ctx, p.pool)
}

func (p *PostgreSQLWrapper) UserByID(ctx context.Context, userID int64) (models.User, error) {
	return UserByID(ctx, p.pool, userID)
}

func SaveRefreshToken(ctx context.Context, storage *pgxpool.Pool,
	userID int64, tokenHash []byte, expires time.Time) error {
	const op = `internal/storage/tokens.SaveRefreshToken`
	const sqlQuery = `INSERT INTO refresh_tokens 
	(user_id, token_hash, created_at, expires_at, revoked_at) VALUES($1, $2, NOW(), $3, NULL);`
	_, err := storage.Exec(ctx, sqlQuery, userID, tokenHash, expires)
	if err != nil {
		return fmt.Errorf("cannot execute sql query %s %w", op, err)
	}
	return nil
}

func RefreshToken(ctx context.Context, storage *pgxpool.Pool, refreshToken []byte) (models.RefreshToken, error) {
	const op = `internal/storage/tokens.RefreshToken`
	var t models.RefreshToken
	const sqlQuery = `SELECT id, user_id, token_hast, created_at, expires_at, revoked_at
	 FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL ORDER BY created_at DESC LIMIT 1;`

	if err := storage.QueryRow(ctx, sqlQuery, refreshToken).Scan(&t.ID, &t.UserID, &t.TokenHash,
		&t.CreatedAt, &t.ExpiredAt, &t.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshToken{}, fmt.Errorf("%s %s %w", op, er.ErrRefreshTokenNotFound, err)
		}
		return models.RefreshToken{}, fmt.Errorf("cannot execute sql query %s %w", op, err)
	}

	return t, nil
}

func RevokeRefreshToken(ctx context.Context, storage *pgxpool.Pool, userID int64) (int64, error) {
	const op = `internal/storage/tokens.RevokeRefreshToken`
	const sqlQuery = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL;`
	tag, err := storage.Exec(ctx, sqlQuery, userID)
	if err != nil {
		return 0, fmt.Errorf("Cannot execute sql query %s %w", op, err)
	}
	return tag.RowsAffected(), nil
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

func UserByID(ctx context.Context, storage *pgxpool.Pool, userID int64) (models.User, error) {
	const op = `internal/storage/tokens.UserByID`
	var user models.User
	const sqlQuery = `SELECT id, email, pass_hash FROM refresh_tokens WHERE user_id = $1;`
	if err := storage.QueryRow(ctx, sqlQuery, userID).Scan(&user.ID, &user.Email, &user.PassHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s %s %w", op, er.ErrUserNotFound, err)
		}
		return models.User{}, fmt.Errorf("cannot execute sql query %s %w", op, err)
	}
	return user, nil
}
