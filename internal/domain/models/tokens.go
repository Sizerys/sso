package models

import "time"

type AccesToken struct {
	ID         int64
	TokenHash  []byte
	UserID     int64
	AppID      int32
	created_at time.Time
	expires_at time.Time
	revoked_at *time.Time
}

type RefreshToken struct {
	ID         int64
	TokenHash  []byte
	UserID     int64
	created_at time.Time
	expires_at time.Time
	revoked_at *time.Time
}

type PairTokens struct {
	acces   AccesToken
	refresh RefreshToken
}
