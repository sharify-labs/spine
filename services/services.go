package services

import (
	"errors"
	"fmt"
	"github.com/sharify-labs/spine/clients"
	"github.com/sharify-labs/spine/database"
	"github.com/sharify-labs/spine/models"
	"gorm.io/gorm"
	"strings"
)

// HostDTO helps parse a hostname string into a usable "object".
type HostDTO struct {
	Full   string
	Sub    string
	Root   string
	UserID string
}

func NewHostDTO(hostname string, userID string) *HostDTO {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	parts := strings.Split(hostname, ".")
	var sub, root string
	switch {
	case len(parts) == 2:
		sub = ""
		root = hostname
	case len(parts) == 3:
		sub = parts[0]
		root = parts[1] + "." + parts[2]
		//sub = strings.Join(parts[:len(parts)-2], ".")  // all except last 2 elements
		//root = strings.Join(parts[len(parts)-2:], ".") // last 2 elements
	default:
		return nil // not long enough or more than 1 level to subdomain
	}
	return &HostDTO{
		Full:   hostname,
		Sub:    sub,
		Root:   root,
		UserID: userID,
	}
}

func (h *HostDTO) dnsRecord() *models.DnsRecord {
	var record models.DnsRecord
	err := database.DB().Where(&models.DnsRecord{Hostname: h.Full}).First(&record).Error
	if err != nil {
		return nil
	}
	return &record
}

func (h *HostDTO) sendToCF() error {
	// Create Cloudflare DNS Record
	record, err := clients.Cloudflare.CreateCNAME(h.UserID, h.Sub, h.Root)
	if err != nil {
		return err
	}

	// Store DNS Record in Database
	err = database.DB().Create(&models.DnsRecord{
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

func (h *HostDTO) sendToDB() error {
	record := h.dnsRecord()
	if record != nil {
		return database.DB().Create(&models.Host{
			DnsRecordID: record.ID,
			UserID:      h.UserID,
			Root:        h.Root,
			Sub:         h.Sub,
		}).Error
	}
	return fmt.Errorf("missing cloudflare recordID for %s", h.Full)
}

func (h *HostDTO) removeFromCF() error {
	// Get dns record from database
	if record := h.dnsRecord(); record != nil {
		if err := clients.Cloudflare.RemoveCNAME(record.ZoneID, record.ID); err != nil {
			return err
		}
		// Successfully removed CNAME DNS Record from Cloudflare
		// -> Remove DnsRecord entry from Database too
		if err := database.DB().Where(
			&models.DnsRecord{Hostname: h.Full},
		).Delete(&models.DnsRecord{}).Error; err != nil {
			// Cloudflare removal success but can't remove DnsRecord from database
			// TODO: Although rare, it's possible so need to add cleanup/rollback here too.
			return err
		}
		// Successfully removed DNS Record from Cloudflare and Database
		return nil
	}
	return fmt.Errorf("failed to delete %s from cloudflare: database is missing dns record", h.Full)
}
func (h *HostDTO) removeFromDB() error {
	return database.DB().Where(&models.Host{
		Sub:    h.Sub,
		Root:   h.Root,
		UserID: h.UserID,
	}).Delete(&models.Host{}).Error
}

// Register writes the host to the database and creates a Cloudflare CNAME entry.
func (h *HostDTO) Register() error {
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

func (h *HostDTO) Delete() error {
	// Root-only hostnames don't have CNAME records -> Skip Cloudflare.
	// Subdomain hostnames should only be removed from Cloudflare if only 1 entry exists in Hosts DB table
	var (
		count           int64
		unique          = false
		missingDBRecord = false
	)
	err := database.DB().Model(&models.Host{}).Where(&models.Host{Sub: h.Sub, Root: h.Root}).Count(&count).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		missingDBRecord = true
	}
	if count == 1 {
		unique = true
	}

	// If hostname has subdomain & only 1 record exists in Hosts table
	// (no other users need it) -> delete dns record from Cloudflare
	if h.Sub != "" && unique {
		if err = h.removeFromCF(); err != nil {
			return err
		}
	}

	if !missingDBRecord {
		if err = h.removeFromDB(); err != nil {
			// TODO: Add logic to undo Cloudflare removal if DB removal fails.
			// This is important because Zephyr uploads strictly rely on the database for validating target hostname.
			// If CF record is removed but DB record persists,
			// users will still be able to upload to 'deleted' hostnames but images will not be viewable.
			return err
		}
	}

	return nil
}
