package services

import (
	"github.com/sharify-labs/spine/database"
	"gorm.io/gorm/clause"
	"strings"
)

// Host helps parse a hostname string into a usable "object".
type Host struct {
	Full   string
	Sub    string
	Root   string
	UserID string
}

// JoinHostname combines a sub and root domain into a single hostname string.
func JoinHostname(sub string, root string) string {
	if sub != "" {
		return sub + "." + root
	}
	return root
}

// NewHostFromFull takes in a full hostname and userID and returns a Host object.
// - Assumes the hostname is pre-validated and sanitized.
// - We only support 1st-level subdomains. We do not need to consider subs with multiple periods.
// - We only support TLDs with 1 period (ex: .co.uk is not supported)
func NewHostFromFull(hostname string, userID string) *Host {
	parts := strings.Split(hostname, ".")
	var sub, root string
	switch {
	// sharify.me
	case len(parts) == 2:
		sub = ""
		root = hostname
	// i.sharify.me
	case len(parts) == 3:
		sub = parts[0]
		root = parts[1] + "." + parts[2]
	// invalid (foo.bar.sharify.me) or (sharify)
	default:
		sub = ""
		root = ""
	}
	return &Host{
		Full:   hostname,
		Sub:    sub,
		Root:   root,
		UserID: userID,
	}
}

// NewHostFromParts takes in a hostname and userID and returns a Host object.
// Note: Assumes the sub and root are pre-validated and sanitized.
func NewHostFromParts(sub string, root string, userID string) *Host {
	return &Host{
		Full:   JoinHostname(sub, root),
		Sub:    sub,
		Root:   root,
		UserID: userID,
	}
}

// Register writes the host to the database.
// Assumes root domain is already added to map of available domains and has an A record set in Cloudflare.
func (h *Host) Register() error {
	return database.DB().Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
		Table:    clause.Table{Name: "dns_records"},
	}).Create(&database.Host{
		UserID: h.UserID,
		Root:   h.Root,
		Sub:    h.Sub,
	}).Error
}

func (h *Host) Delete() error {
	// Lock hosts table to prevent multiple txs from trying to delete the same host.
	return database.DB().Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
		Table:    clause.Table{Name: clause.CurrentTable},
	}).Where(&database.Host{
		Sub:    h.Sub,
		Root:   h.Root,
		UserID: h.UserID,
	}).Delete(&database.Host{}).Error
}
