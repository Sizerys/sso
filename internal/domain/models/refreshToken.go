package models

import "time"

type RefreshToken struct {
	ID         int64
	TokenHash  []byte
	UserID     int64
	created_at time.Time
	expires_at time.Time
	revoked_at *time.Time
}

func (t RefreshToken) IsExpired() bool {
	return time.Now().After(t.expires_at)
}

func (t RefreshToken) IsRevoked() bool {
	return t.revoked_at != nil
}

func (t RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}
