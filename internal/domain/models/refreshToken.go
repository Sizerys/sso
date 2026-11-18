package models

import "time"

type RefreshToken struct {
	ID         int64
	TokenHash  []byte
	UserID     int64
	CreatedAt time.Time
	ExpiredAt time.Time
	RevokedAt *time.Time
}

func (t RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiredAt)
}

func (t RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}
