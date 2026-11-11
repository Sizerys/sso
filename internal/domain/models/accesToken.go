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

func (t AccesToken) IsExpired() bool {
	return time.Now().After(t.expires_at)
}

func (t AccesToken) IsRevoked() bool {
	return t.revoked_at != nil
}

func (t AccesToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}
