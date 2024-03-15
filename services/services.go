package services

import (
	"errors"
	"github.com/posty/spine/clients"
	"github.com/posty/spine/database"
	"github.com/posty/spine/models"
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

func (h *HostDTO) exists() (bool, error) {
	err := database.DB().Where(&models.Host{
		Sub:  h.Sub,
		Root: h.Root,
	}).First(&models.Host{}).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (h *HostDTO) sendToCF() (string, error) {
	return clients.Cloudflare.CreateCNAME(h.UserID, h.Sub, h.Root)
}

func (h *HostDTO) sendToDB(recordID string) error {
	return database.DB().Create(&models.Host{
		RecordID: recordID,
		UserID:   h.UserID,
		Root:     h.Root,
		Sub:      h.Sub,
	}).Error
}

func (h *HostDTO) removeFromCF() error {
	// Get RecordID from database
	var host models.Host
	err := database.DB().Where(&models.Host{
		Sub:    h.Sub,
		Root:   h.Root,
		UserID: h.UserID,
	}).First(&host).Error
	if err != nil {
		return err
	}

	return clients.Cloudflare.RemoveCNAME(h.Root, host.RecordID)
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
	var (
		recordID string
		err      error
	)

	if h.Sub != "" {
		// Note: Only hosts with subdomains need to be sent to CF.
		recordID, err = h.sendToCF()
		if err != nil {
			return err
		}
	}

	if err = h.sendToDB(recordID); err != nil {
		// TODO: Add logic to undo Cloudflare change or store RecordID somewhere if database fails
		return err
	}

	return nil
}

func (h *HostDTO) Delete() error {
	// Note: Only hosts with subdomains need to be deleted from CF. Root-only hostnames don't have CNAME records.
	if h.Sub != "" {
		if err := h.removeFromCF(); err != nil {
			return err
		}
	}

	if err := h.removeFromDB(); err != nil {
		// TODO: Add logic to undo Cloudflare removal if DB removal fails.
		// This is important because Zephyr uploads strictly rely on the database for validating target hostname.
		// If CF record is removed but DB record persists,
		// users will still be able to upload to 'deleted' hostnames but images will not be viewable.
		return err
	}

	return nil
}
