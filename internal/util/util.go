package util

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	mathrand "math/rand"
)

// GenerateID generates a random ID
func GenerateID(sz int) string {
	buf := make([]byte, sz)
	if _, err := cryptorand.Read(buf); err != nil {
		_, _ = mathrand.Read(buf)
	}
	return hex.EncodeToString(buf)
}
