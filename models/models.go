package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
)

// Upload represents an image uploaded to our app.
type Upload struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	StorageKey string `gorm:"unique"`
	UserID     string `gorm:"index"` // fk -> User.ID
	User       User   // required for M-1 relationship (I think)
}

// Key represents a User's upload key.
type Key struct {
	gorm.Model
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	Hash   string `gorm:"unique;not null"`
	Salt   string `gorm:"not null"`
	UserID string `gorm:"index"` // fk -> User.ID
}

// Domain represents a domain name that can be used by a User to build a Host.
// Example: stole-my-blow.wtf
type Domain struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"unique;not null"`
}

// Host represents a FQDN that a User can upload to.
// Example:
//
//	"id": 1,
//	"sub": "joe",
//	"root": "stole-my-blow.wtf",
//	"full": "joe.stole-my-blow.wtf"
//	"user_id": "c99e9b2c-f04b-421e-b4a5-8120d2513b93"
type Host struct {
	gorm.Model
	ID     uint   `gorm:"primaryKey;autoincrement"`
	Full   string // (Host.Sub + Host.Root) or (Host.Sub + Domain.Name)
	Sub    string
	Root   string `gorm:"index"` // fk -> Domain.Name
	UserID string `gorm:"index"` // fk -> User.ID  (I don't know why but "index" tags required for fk to work)
	Domain Domain `gorm:"foreignKey:Root;references:Name"`
	User   User   // required for M-1 relationship (I think)
}

func (ud *Host) BeforeCreate(_ *gorm.DB) (err error) {
	ud.Full = strings.ToLower(ud.Sub + "." + ud.Root)
	return
}

// User represents a person registered on our platform.
type User struct {
	gorm.Model
	ID      string `gorm:"primaryKey"`
	Email   string `gorm:"unique;not null"`
	PlanID  *uint  // temp nullable
	Plan    *Plan  // temp nullable
	Key     Key
	Hosts   []Host
	Uploads []Upload
}

func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	return
}

// Plan represents a User's plan and describes pricing & limits.
type Plan struct {
	gorm.Model
	ID         uint    `gorm:"primaryKey;autoIncrement"`
	Name       string  `gorm:"unique;not null"`
	Price      float32 `gorm:"unique;not null"`
	MaxHosts   int     `gorm:"not null"`
	MaxUploads int     `gorm:"not null"`
}
