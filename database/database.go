package database

import (
	"encoding/base64"
	"fmt"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/models"
	"github.com/sharify-labs/spine/utils"
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
		&models.Token{},
		&models.Upload{},
		&models.Host{},
		&models.DnsRecord{},
	)
	if err != nil {
		panic(err)
	}
}

func DB() *gorm.DB {
	return db
}

func getAllHosts(userID string) ([]*models.Host, error) {
	var hosts []*models.Host
	err := db.Where(&models.Host{UserID: userID}).Find(&hosts).Error
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

func GetAllHostnames(userID string) ([]string, error) {
	hosts, err := getAllHosts(userID)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, host := range hosts {
		names = append(names, utils.CompileHostname(host.Sub, host.Root))
	}
	return names, nil
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
