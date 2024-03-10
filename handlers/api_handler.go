package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/posty/spine/config"
	"github.com/posty/spine/database"
	"github.com/posty/spine/dto"
	"github.com/posty/spine/security"
	"strings"
)

// ResetKeyHandler refreshes an existing user's upload-key.
func ResetKeyHandler(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("user_id", ""))
	if userID == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	key := security.NewZephyrKey()
	if key == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if err := database.UpdateUserKey(userID, key.Hash, key.Salt); err != nil {
		log.Errorf("Failed to save key in database: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		config.UserHeader: userID,
		config.KeyHeader:  key.Key,
	})
}

// ListDomainsHandler returns a JSON array of all available root domain names.
func ListDomainsHandler(c *fiber.Ctx) error {
	domains, err := database.GetDomainsAvailable()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	var names []string
	for _, domain := range domains {
		names = append(names, domain.Name)
	}
	return c.Status(fiber.StatusOK).JSON(names)
}

// ListHostsHandler returns a JSON array of all hosts registered by a given user.
func ListHostsHandler(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("user_id", ""))
	if userID == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	hosts, err := database.GetAllHosts(userID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	var names []string
	for _, host := range hosts {
		names = append(names, host.Full)
	}
	return c.Status(fiber.StatusOK).JSON(names)
}

// CreateHostHandler creates new hosts for a user.
// Root domain of hostname must be registered first. This can be checked with ListDomainsHandler.
func CreateHostHandler(c *fiber.Ctx) error {
	hostname := c.Params("hostname")
	host := dto.NewHost(hostname)
	if host == nil {
		// Hostname does not meet format requirements. Should prob be validated on frontend too.
		return c.SendStatus(fiber.StatusBadRequest)
	}

	userID := strings.TrimSpace(c.Params("user_id"))
	err := database.InsertHost(userID, host.Sub, host.Root)
	if err != nil {
		// Unable to create host in DB. Maybe root domain doesn't exist, or user doesn't exist, or something else.
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		config.UserHeader: userID,
		config.HostHeader: host.Full,
	})
}

// DeleteHostHandler deletes a registered hostname.
func DeleteHostHandler(c *fiber.Ctx) error {
	hostname := c.Params("hostname")
	userID := strings.TrimSpace(c.Params("user_id"))
	err := database.DeleteHost(userID, hostname)
	if err != nil {
		// Unable to create host in DB. Maybe root domain doesn't exist, or user doesn't exist, or something else.
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.SendStatus(fiber.StatusOK)
}
