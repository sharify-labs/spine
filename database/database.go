package database

import (
	"errors"
	"github.com/gofiber/storage/memory/v2"
	"github.com/markbates/goth"
	"github.com/sharify-labs/spine/config"
	"github.com/sharify-labs/spine/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		DSN:           config.Str("MYSQL_DSN"),
		ServerVersion: "MariaDB",
	}), &gorm.Config{TranslateError: true})
	if err != nil {
		panic(err)
	}

	// Migrations
	//err = db.AutoMigrate(
	//	&StorageKey{},
	//	&User{},
	//	&Token{},
	//	&Plan{},
	//	&Upload{},
	//	&Host{},
	//	&DnsRecord{},
	//)
	//if err != nil {
	//	panic(err)
	//}
}

func connectCache() {
	cache = memory.New(memory.Config{GCInterval: time.Minute * 5})
}

func getAllHosts(userID string) ([]*Host, error) {
	var hosts []*Host
	err := db.Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
	}).Where(&Host{
		UserID: userID,
	}).Find(&hosts).Error
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
	err := db.Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
	}).Where(&Upload{UserID: userID}).Find(&uploads).Error
	if err != nil {
		return nil, err
	}
	return uploads, nil
}

func GetOrCreateUser(gothUser goth.User) (*User, error) {
	var user User
	err := db.Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
	}).Where(&User{
		Email:     strings.TrimSpace(strings.ToLower(gothUser.Email)),
		DiscordID: &gothUser.UserID,
	}).FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
