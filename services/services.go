package services

import (
	"github.com/posty/spine/clients"
	"github.com/posty/spine/database"
	"github.com/posty/spine/models"
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
		Full:   full,
		Sub:    sub,
		Root:   root,
		UserID: userID,
	}
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
	var host models.Host
	err := database.DB().Where(&models.Host{
		Full:   h.Full,
		UserID: h.UserID,
	}).First(&host).Error
	if err != nil {
		return err
	}

	return clients.Cloudflare.RemoveCNAME(h.Root, host.RecordID)
}
func (h *HostDTO) removeFromDB() error {
	return database.DB().Where(&models.Host{
		UserID: h.UserID,
		Full:   h.Full,
	}).Delete(&models.Host{}).Error
}

// Register writes the host to the database and creates a Cloudflare CNAME entry.
func (h *HostDTO) Register() error {
	recordID, err := h.sendToCF()
	if err != nil {
		return err
	}

	if err = h.sendToDB(recordID); err != nil {
		// TODO: Add logic to undo Cloudflare change or store RecordID somewhere if database fails
		return err
	}

	return nil
}

func (h *HostDTO) Delete() error {
	if err := h.removeFromCF(); err != nil {
		return err
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
