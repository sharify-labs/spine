package database

import (
	"errors"
	"github.com/gofiber/storage/memory/v2"
	"github.com/markbates/goth"
	"github.com/sharify-labs/spine/config"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

// db SQL gorm connector
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
	db, err = gorm.Open(sqlite.New(sqlite.Config{
		DSN:        config.Str("TURSO_DSN"),
		DriverName: "libsql",
	}), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.LogLevel(config.Int("LOG_LEVEL")),
				Colorful:      true,
			},
		),
	})
	if err != nil {
		panic(err)
	}
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
		if host.Sub != "" {
			names = append(names, host.Sub+"."+host.Root)
		} else {
			names = append(names, host.Root)
		}
	}
	return names, nil
}

func GetUserUploads(userID string) ([]*Upload, error) {
	var uploads []*Upload
	err := db.Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
	}).Where(&Upload{
		UserID: userID,
	}).Find(&uploads).Error
	if err != nil {
		return nil, err
	}
	return uploads, nil
}

// GetOrCreateUser Retrieves a user by Discord ID from the database. If not found, creates a new record.
// Also assigns the provided email to the record, regardless of if the record is found.
func GetOrCreateUser(gothUser goth.User) (*User, error) {
	var user User
	err := db.Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
	}).Where(User{
		DiscordID: &gothUser.UserID,
	}).Assign(User{
		Email: strings.TrimSpace(strings.ToLower(gothUser.Email)),
	}).FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
