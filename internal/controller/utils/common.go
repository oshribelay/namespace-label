package utils

import (
	"math/rand"
	"strings"
)

// IsReservedLabel checks if a label has a protected prefix.
func IsReservedLabel(label string, protectedPrefixes map[string]string) bool {
	for prefix := range protectedPrefixes {
		if strings.HasPrefix(label, prefix) {
			return true
		}
	}
	return false
}

// EqualLabels checks if two maps of labels are equal to each other.
func EqualLabels(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		if valueB, exists := b[key]; !exists || valueA != valueB {
			return false
		}
	}

	return true
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateRandomString generates a random string of length n.
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
