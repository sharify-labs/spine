package services

import (
	"errors"
	"fmt"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/database"
	"gorm.io/gorm"
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

func (h *Host) dnsRecord() *database.DnsRecord {
	var record database.DnsRecord
	err := database.DB().Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
	}).Where(&database.DnsRecord{
		Hostname: h.Full,
	}).First(&record).Error
	if err != nil {
		return nil
	}
	return &record
}

func (h *Host) sendToCF() error {
	// Create Cloudflare DNS Record
	record, err := clients.Cloudflare.CreateCNAME(h.UserID, h.Sub, h.Root)
	if err != nil {
		return err
	}

	// Store DNS Record in Database
	// TODO: Double-check no lock is necessary here.
	err = database.DB().Create(&database.DnsRecord{
		ID:       record.ID,
		ZoneID:   record.ZoneID,
		Hostname: h.Full,
	}).Error

	if err != nil {
		// Database insert failed, roll back Cloudflare change
		err = clients.Cloudflare.RemoveCNAME(record.ZoneID, record.ID)
		if err != nil {
			// Created CNAME, Database failed, and now can't remove CNAME -> this edge case will create mess.
			// TODO: Consider putting details of this event somewhere safe so we can manually fix it
			return fmt.Errorf("cf record created, database failed, unable to remove cf record: %v", err)
		}
	}

	return nil
}

func (h *Host) sendToDB() error {
	record := h.dnsRecord()
	if record != nil {
		// Lock dns_records table from getting updated while creating host
		// to ensure host is not getting created while dns_record is getting deleted elsewhere.
		return database.DB().Clauses(clause.Locking{
			Strength: clause.LockingStrengthShare,
			Table:    clause.Table{Name: "dns_records"},
		}).Create(&database.Host{
			DnsRecordID: record.ID,
			UserID:      h.UserID,
			Root:        h.Root,
			Sub:         h.Sub,
		}).Error
	}
	return fmt.Errorf("missing cloudflare recordID for %s", h.Full)
}

func (h *Host) removeFromCF() error {
	// Get dns record from database
	if record := h.dnsRecord(); record != nil {
		if err := clients.Cloudflare.RemoveCNAME(record.ZoneID, record.ID); err != nil {
			return err
		}
		// Successfully removed CNAME DNS Record from Cloudflare
		// -> Remove DnsRecord entry from Database too
		// Lock dns_records table to prevent multiple txs from trying to delete the same record.
		if err := database.DB().Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: clause.CurrentTable},
		}).Where(&database.DnsRecord{
			Hostname: h.Full,
		}).Delete(&database.DnsRecord{}).Error; err != nil {
			// Cloudflare removal success but can't remove DnsRecord from database
			// TODO: Although rare, it's possible so need to add cleanup/rollback here too.
			return err
		}
		// Successfully removed DNS Record from Cloudflare and Database
		return nil
	}
	return fmt.Errorf("failed to delete %s from cloudflare: database is missing dns record", h.Full)
}
func (h *Host) removeFromDB() error {
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

// Register writes the host to the database and creates a Cloudflare CNAME entry.
func (h *Host) Register() error {
	dnsRecord := h.dnsRecord()
	// If host has subdomain -> might need to create DNS entry
	if h.Sub != "" {
		// TODO: Add check for limit of 1000 DNS records per domain.
		if dnsRecord == nil {
			// CNAME Record missing -> send to CF (this will also create DnsRecord db entry)
			if err := h.sendToCF(); err != nil {
				return err
			}
		}
	}

	if err := h.sendToDB(); err != nil {
		// TODO: Consider adding logic to undo Cloudflare change if database fails
		return err
	}

	return nil
}

func (h *Host) Delete() error {
	// Root-only hostnames don't have CNAME records -> Skip Cloudflare.
	// Subdomain hostnames should only be removed from Cloudflare if only 1 entry exists in Hosts DB table
	var count int64

	// Lock hosts table from getting updated while counting records to prevent race condition (I think?)
	err := database.DB().Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
		Table:    clause.Table{Name: "hosts"},
	}).Model(&database.Host{}).Where(&database.Host{
		Sub:  h.Sub,
		Root: h.Root,
	}).Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // no records exist with given hostname
		}
		return err
	}

	// TODO: Not sure if 0 results will return ErrRecordNotFound or just return count of 0. Adding this for safety.
	if count == 0 {
		return nil
	}

	// If hostname has subdomain & only 1 record exists in Hosts table
	// (no other users need it) -> delete dns record from Cloudflare
	if h.Sub != "" && count == 1 {
		if err = h.removeFromCF(); err != nil {
			return err
		}
	}

	if err = h.removeFromDB(); err != nil {
		// TODO: Add logic to undo Cloudflare removal if DB removal fails.
		// This is important because Zephyr uploads strictly rely on the database for validating target hostname.
		// If CF record is removed but DB record persists,
		// users will still be able to upload to 'deleted' hostnames but images will not be viewable.
		return err
	}

	return nil
}
