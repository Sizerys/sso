package refresh

import (
	"crypto/rand"
	"encoding/base64"
)

func RefreshToken(lenght int) string {
	b := make([]byte, lenght)
	if _, err := rand.Read(b); err != nil {
		panic("Failed to generate refresh token")
		// точно ли стоит оставлять panic
	}
	return base64.URLEncoding.EncodeToString(b)
}
