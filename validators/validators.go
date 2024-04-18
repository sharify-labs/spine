package validators

import (
	"strings"
	"unicode"
)

// firstNChars returns the first n number of characters from a string.
func firstNChars(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:63])
	}
	return s
}

// SanitizeSubdomain validates and sanitizes subdomains.
//  1. Trims spaces
//  2. Makes all characters lowercase
//  3. Replaces all periods with hyphens to ensure only 1 level.
//  4. Removes invalid characters (allows a-z, 0-9, and hyphens)
//  5. Removes leading/trailing hyphens.
func SanitizeSubdomain(sub string) string {
	var sb strings.Builder
	str := strings.ReplaceAll(strings.TrimSpace(strings.ToLower(sub)), ".", "-")
	for i, c := range str {
		if i >= 200 {
			break // Only check first 200 characters for safety
		}
		if unicode.IsLetter(c) || unicode.IsDigit(c) || (c == '-') {
			sb.WriteRune(c)
		}
	}
	// Ensure subdomain doesn't start/end with a hyphen
	sanitized := strings.Trim(sb.String(), "-")

	// Enforce maximum length of 63 characters
	sanitized = firstNChars(sanitized, 63)

	// Trim hyphens again in case slicing string resulted in trailing hyphen.
	sanitized = strings.TrimRight(sanitized, "-")
	return sanitized
}
