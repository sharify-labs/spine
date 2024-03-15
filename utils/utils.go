package utils

import "strings"

func CompileHostname(sub string, root string) string {
	sub = strings.ToLower(strings.TrimSpace(sub))
	root = strings.ToLower(strings.TrimSpace(root))
	if sub != "" {
		return sub + "." + root
	}
	return root
}
