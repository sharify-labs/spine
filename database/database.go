package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/posty/spine/config"
	"github.com/posty/spine/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strings"
)

// db MySQL/MariaDB gorm connector
var (
	db       *gorm.DB
	sessions *session.Store
)

func Setup() {
	connectSQL()
	connectSessions()
}

func connectSQL() {
	var err error

	db, err = gorm.Open(mysql.New(mysql.Config{
		ServerVersion: "MariaDB",
		DSN: fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetStr("MYSQL_USER"),
			config.GetStr("MYSQL_PASS"),
			config.GetStr("MYSQL_HOST"),
			config.GetStr("MYSQL_PORT"),
			config.GetStr("MYSQL_DB"),
		),
	}),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Migrations
	if !fiber.IsChild() {
		err = db.AutoMigrate(
			&models.User{},
			&models.Plan{},
			&models.Domain{},
			&models.Key{},
			&models.Upload{},
			&models.Host{},
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func connectSessions() {
	sessions = session.New()
}

func GetSession(c *fiber.Ctx) (*session.Session, error) {
	return sessions.Get(c)
}

func InsertDomain(name string) error {
	return db.Create(&models.Domain{
		Name: strings.ToLower(strings.TrimSpace(name)),
	}).Error
}

func GetDomain(name string) (*models.Domain, error) {
	var domain models.Domain
	err := db.Where(&models.Domain{
		Name: strings.ToLower(strings.TrimSpace(name))}).First(&domain).Error
	if err != nil {
		log.Println("failed to find domain")
		return nil, err
	}
	return &domain, nil
}

func GetDomainsAvailable() ([]*models.Domain, error) {
	var domains []*models.Domain
	err := db.Find(&domains).Error
	if err != nil {
		return nil, err
	}
	return domains, nil
}

func InsertHost(userID string, sub string, root string) error {
	// Ensure root domain exists
	domain, err := GetDomain(root)
	if err != nil {
		return errors.New("root domain does not exist")
	}

	// Create host
	host := models.Host{
		UserID: userID,
		Root:   domain.Name,
		Sub:    strings.TrimSpace(strings.ToLower(sub)),
	}
	err = db.Create(&host).Error
	if err != nil {
		return err
	}
	return nil
}

func DeleteHost(userID string, hostname string) error {
	return db.Where(&models.Host{
		UserID: userID,
		Full:   strings.TrimSpace(strings.ToLower(hostname)),
	}).Delete(&models.Host{}).Error
}

func GetHost(userID string, sub string, root string) (*models.Host, error) {
	var host models.Host
	err := db.Where(&models.Host{
		UserID: userID, Sub: sub, Root: root,
	}).First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func GetAllHosts(userID string) ([]*models.Host, error) {
	var hosts []*models.Host
	err := db.Where(&models.Host{UserID: userID}).Find(&hosts).Error
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

func UpdateUserKey(userID string, hash []byte, salt []byte) error {
	return db.Save(&models.Key{
		UserID: userID,
		Hash:   base64.URLEncoding.EncodeToString(hash),
		Salt:   base64.URLEncoding.EncodeToString(salt),
	}).Error
}
