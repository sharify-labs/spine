package models

import "gorm.io/gorm"

type Plan struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	Name        string
	Price       float32 `gorm:"not null"`
	UploadLimit int     `gorm:"not null"`
	DomainLimit int     `gorm:"not null"`
}

// Upload represents an image uploaded to our app
type Upload struct {
	gorm.Model
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	UserID string `gorm:"not null"`
	R2Key  string `gorm:"->;<-:create;unique;size:128"`
}

// User represents a person registered on our platform.
type User struct {
	gorm.Model
	ID                uint   `gorm:"primaryKey;autoIncrement"`
	Email             string `gorm:"not null"`
	PlanID            uint   `gorm:"not null"`
	Plan              Plan   `gorm:"foreignKey:PlanID;references;ID"`
	Uploads           []Upload
	ZephyrKeys        []ZephyrKey
	RegisteredDomains []RegisteredDomain `gorm:"many2many:user_registered_domains"`
}

// RegisteredDomain represents a FQDN that a User can upload to
type RegisteredDomain struct {
	gorm.Model
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	RootDomain string `gorm:"not null"`
	SubDomain  string
	Domain     Domain `gorm:"foreignKey:Name;references:root_domain"`
}

// Domain represents a hostname ready to be used by a User to build a RegisteredDomain
type Domain struct {
	gorm.Model
	Name string `gorm:"primaryKey;not null;"`
}

// ZephyrKey represents a user's upload key
type ZephyrKey struct {
	gorm.Model
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	UserID string `gorm:"not null"`
	Domain string `gorm:"->;<-:create;not null"` // "->;<-:create" means 'allow read and create' (no update)
	Hash   string `gorm:"->;<-:create;size:256;not null"`
	Salt   string `gorm:"->;<-:create;size:256;not null"`
}

// Key represents the application key
type Key struct {
	Hash string
	Salt string
}
