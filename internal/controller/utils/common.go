package utils

import "strings"

// IsReservedLabel checks if a label has a protected prefix.
func IsReservedLabel(label string, protectedPrefixes map[string]string) bool {
	for prefix := range protectedPrefixes {
		if strings.HasPrefix(label, prefix) {
			return true
		}
	}
	return false
}

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
