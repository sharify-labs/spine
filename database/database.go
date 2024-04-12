package database

import (
	"fmt"
	goccy "github.com/goccy/go-json"
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

// cache local memory storage connector
// db SQL gorm connector
var (
	cache *memory.Storage
	db    *gorm.DB
)

func Setup() {
	var err error
	cache = memory.New(memory.Config{GCInterval: time.Minute * 5})
	db, err = gorm.Open(sqlite.New(sqlite.Config{
		DSN:        config.Get[string]("TURSO_DSN"),
		DriverName: "libsql",
	}), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.LogLevel(config.Get[int]("LOG_LEVEL")),
				Colorful:                  true,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		panic(err)
	}
}

// DB retrieves gorm connector for SQL Database.
func DB() *gorm.DB {
	return db
}

func AddToCache(key string, data interface{}, exp time.Duration) {
	var serialized []byte
	var err error
	switch v := data.(type) {
	case []byte:
		serialized = v
	case string:
		serialized = []byte(v)
	default:
		serialized, err = goccy.Marshal(data)
		if err != nil {
			fmt.Printf("unable to marshal data for cache key %s: %v", key, err)
			return
		}
	}
	if err = cache.Set(key, serialized, exp); err != nil {
		fmt.Printf("unable to cache data for key %s: %v", key, err)
	}
}

// GetFromCache retrieves an object from the cache.
// If an error occurs, logs it and does not store anything in output object.
func GetFromCache(key string, output interface{}) {
	data, err := cache.Get(key)
	if err != nil {
		fmt.Printf("unable to get %s from cache: %v", key, err)
		return
	}
	if data != nil {
		if err = goccy.Unmarshal(data, output); err != nil {
			fmt.Printf("unable to unmarshal cached data for %s: %v", key, err)
		}
	}
}

func GetAllHostnames(userID string) ([]string, error) {
	var hosts []*Host
	if err := db.Clauses(clause.Locking{
		Strength: clause.LockingStrengthShare,
	}).Where(&Host{
		UserID: userID,
	}).Find(&hosts).Error; err != nil {
		return nil, err
	}

	var names []string
	for _, h := range hosts {
		if h.Sub != "" {
			names = append(names, h.Sub+"."+h.Root)
		} else {
			names = append(names, h.Root)
		}
	}
	return names, nil
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
