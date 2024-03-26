package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

// StorageKey represents a used R2 key.
// The purpose of this table is to keep a running log of all used keys.
// The rows in this table cannot be deleted.
// This is to prevent users from reusing old host/secret combinations.
// Note: This table is disconnected from user data,
// allowing us to keep it even if a user purges their account.
// ID: Represents the value of the storage key (SHA-512 hash).
type StorageKey struct {
	gorm.Model
	ID string `gorm:"primaryKey;<-:create"` // allow read and create
}

// Upload represents an upload to our app.
type Upload struct {
	gorm.Model
	ID         uint `gorm:"primaryKey;autoIncrement"`
	Exp        *time.Time
	StorageKey string `gorm:"unique;not null"`
	UserID     string `gorm:"index"` // fk -> User.ID
	User       User   // required for M-1 relationship (I think)
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
	ID          uint `gorm:"primaryKey;autoincrement"`
	Sub         string
	Root        string
	UserID      string     `gorm:"index"` // fk -> User.ID  (I don't know why but "index" tags required for fk to work)
	User        User       // required for M-1 relationship (I think)
	DnsRecordID string     `gorm:"index"`
	DnsRecord   *DnsRecord // can be null if hostname has no subdomain
}

func (h *Host) BeforeCreate(_ *gorm.DB) (err error) {
	h.Sub = strings.ToLower(strings.TrimSpace(h.Sub))
	h.Root = strings.ToLower(strings.TrimSpace(h.Root))
	return
}

type DnsRecord struct {
	gorm.Model
	ID       string `gorm:"primaryKey"` // DNS Record ID returned from Cloudflare API Request
	ZoneID   string `gorm:"not null"`
	Hostname string `gorm:"not null"`
}

func (dr *DnsRecord) BeforeCreate(_ *gorm.DB) (err error) {
	dr.Hostname = strings.ToLower(strings.TrimSpace(dr.Hostname))
	return
}

// Token represents a user's upload token.
type Token struct {
	gorm.Model
	ID     string `gorm:"primaryKey"`
	Hash   string `gorm:"unique;not null"`
	UserID string `gorm:"unique;index"`
	User   User
}

// User represents a person registered on our platform.
type User struct {
	gorm.Model
	ID      string  `gorm:"primaryKey"`
	Email   string  `gorm:"unique;not null"`
	TokenID *string `gorm:"unique;index"`
	Token   *Token  `gorm:"foreignKey:TokenID"`
	PlanID  *uint   // temp nullable
	Plan    *Plan   // temp nullable
	Hosts   []Host
	Uploads []Upload
}

func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
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
