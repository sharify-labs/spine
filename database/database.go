package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gofiber/storage/memory/v2"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

// db MySQL/MariaDB gorm connector
// cache local memory storage connector
var (
	db    *gorm.DB
	cache *memory.Storage
)

func Setup() {
	connectDB()
	connectCache()
}

// DB retrieves gorm connector for SQL Database.
func DB() *gorm.DB {
	return db
}

// Cache retrieves memory storage connector used for caching.
func Cache() *memory.Storage {
	return cache
}

func connectDB() {
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
		&StorageKey{},
		&User{},
		&Plan{},
		&Upload{},
		&Host{},
		&DnsRecord{},
	)
	if err != nil {
		panic(err)
	}
}

func connectCache() {
	cache = memory.New(memory.Config{GCInterval: time.Minute * 5})
}

func getAllHosts(userID string) ([]*Host, error) {
	var hosts []*Host
	err := db.Where(&Host{UserID: userID}).Find(&hosts).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
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

func GetUserUploads(userID string) ([]*Upload, error) {
	var uploads []*Upload
	err := db.Where(&Upload{UserID: userID}).Find(&uploads).Error
	if err != nil {
		return nil, err
	}
	return uploads, nil
}

func GetOrCreateUser(email string) (*User, error) {
	var user User
	err := db.Where(&User{
		Email: strings.TrimSpace(strings.ToLower(email)),
	}).FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
