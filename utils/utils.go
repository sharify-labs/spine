package utils

import (
	"strings"
	"unicode"
)

// SanitizeSubdomain removes invalid characters from subdomains.
// Note: custom paths are limited to 255 characters. Anything remaining will be cut off.
// Allowed characters include: a-z, A-Z, 0-9, hyphens, and underscores.
func SanitizeSubdomain(sub string) string {
	var sb strings.Builder
	str := strings.ToLower(strings.TrimSpace(sub))
	runeCount := 0
	for i, c := range str {
		if runeCount >= 63 {
			break // Enforce maximum length of 63 characters for a subdomain
		}
		if unicode.IsLetter(c) || unicode.IsDigit(c) || (c == '-' && i != 0) {
			sb.WriteRune(c)
			runeCount++
		}
	}
	sanitized := sb.String()
	if strings.HasSuffix(sanitized, "-") {
		sanitized = sanitized[:len(sanitized)-1] // Ensure subdomain does not end with a hyphen
	}
	return sanitized
}

func CompileHostname(sub string, root string) string {
	sub = strings.ToLower(strings.TrimSpace(sub))
	root = strings.ToLower(strings.TrimSpace(root))
	if sub != "" {
		return sub + "." + root
	}
	return root
}
