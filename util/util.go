package util

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateUUID generates a proper UUID for the EPUB
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		// Fallback to a timestamp-based ID if random generation fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}

	// Set version (4) and variant (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
