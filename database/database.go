package database

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/posty/spine/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

// DB gorm connector
var DB *gorm.DB

func ConnectDB() {
	var err error

	url := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GetStr("MYSQL_USER"),
		config.GetStr("MYSQL_PASS"),
		config.GetStr("MYSQL_HOST"),
		config.GetStr("MYSQL_PORT"),
		config.GetStr("MYSQL_DB"),
	)
	DB, err = gorm.Open(mysql.Open(url), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatal(err)
	}

	// Migrations
	if !fiber.IsChild() {
		err = DB.AutoMigrate(&models.Upload{}, &models.Key{})
		if err != nil {
			log.Fatal(err)
		}
	}
}
