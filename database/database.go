package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/posty/spine/config"
	"github.com/posty/spine/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

// db MySQL/MariaDB gorm connector
var (
	db *gorm.DB
)

func Setup() {
	connectSQL()
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
		panic(err)
	}

	// Migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Plan{},
		&models.Domain{},
		&models.Token{},
		&models.Upload{},
		&models.Host{},
	)
	if err != nil {
		panic(err)
	}
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

func GetUserUploads(userID string) ([]*models.Upload, error) {
	var uploads []*models.Upload
	err := db.Where(&models.Upload{UserID: userID}).Find(&uploads).Error
	if err != nil {
		return nil, err
	}
	return uploads, nil
}

func GetOrCreateUser(email string) (*models.User, error) {
	var user models.User
	err := db.Where(&models.User{
		Email: strings.TrimSpace(strings.ToLower(email)),
	}).FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserToken retrieves user and updates their upload-key.
// TODO: Modify this so it's done in 1 query
func UpdateUserToken(userID string, hash []byte, salt []byte) error {
	token := models.Token{}
	err := db.Where(&models.Token{UserID: userID}).FirstOrCreate(&token).Error
	if err != nil {
		return err
	}
	token.Hash = base64.URLEncoding.EncodeToString(hash)
	token.Salt = base64.URLEncoding.EncodeToString(salt)

	return db.Save(&token).Error
}
