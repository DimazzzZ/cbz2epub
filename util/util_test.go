package util

import (
	"regexp"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	// Test that GenerateUUID returns a valid UUID
	uuid := GenerateUUID()

	// Check that the UUID matches the expected format (RFC 4122 version 4)
	// Format: 8-4-4-4-12 hexadecimal digits
	pattern := "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	matched, err := regexp.MatchString(pattern, uuid)
	if err != nil {
		t.Fatalf("Error matching UUID pattern: %v", err)
	}

	if !matched {
		t.Errorf("Generated UUID %s does not match expected format", uuid)
	}

	// Test that multiple calls generate different UUIDs
	uuid2 := GenerateUUID()
	if uuid == uuid2 {
		t.Errorf("Generated UUIDs are not unique: %s == %s", uuid, uuid2)
	}
}
