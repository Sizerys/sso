package refresh

import (
	"crypto/rand"
	"encoding/base64"
)

func Token(lenght int) string {
	b := make([]byte, lenght)
	if _, err := rand.Read(b); err != nil {

	}
	return base64.URLEncoding.EncodeToString(b)
}
