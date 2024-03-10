package dto

import "strings"

// HostDTO helps parse a hostname string into a usable "object".
type HostDTO struct {
	Full string
	Sub  string
	Root string
}

func NewHost(hostname string) *HostDTO {
	full := strings.ToLower(strings.TrimSpace(hostname))
	parts := strings.Split(full, ".")
	var sub, root string
	switch {
	case len(parts) == 2:
		sub = ""
		root = full
	case len(parts) > 2:
		sub = strings.Join(parts[:len(parts)-2], ".")  // all except last 2 elements
		root = strings.Join(parts[len(parts)-2:], ".") // last 2 elements
	default:
		return nil // not long enough
	}
	return &HostDTO{
		Full: full,
		Sub:  sub,
		Root: root,
	}
}
