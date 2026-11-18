package models

import "time"

type AccesToken struct {
	ID        int64
	TokenHash []byte
	UserID    int64
	AppID     int32
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (t AccesToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t AccesToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t AccesToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}
