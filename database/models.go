package database

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
// ID: Auto-incrementing integer.
// Type: Internal type [file/image/paste/redirect] (see services.UploadType)
// Size: Stored data size.
// Mime: MIME string of the stored data. NULL for pastes/redirects.
// Title: A display name for this Upload in the gallery. Defaults to filename for files/images.
// Exp: Date and Time when the Upload expires. NULL for permanent uploads.
// StorageKey: The key used for storing the Upload in R2 and Redis Cache.
// Hostname: The full hostname where the Upload is served from (ex: sharify.me)
// Secret: The secret used to view the Upload (ex: Rn1P8Zqz)
// UserID: The User.ID of the publisher (foreign key).
type Upload struct {
	gorm.Model
	ID         uint    `gorm:"primaryKey;autoincrement"`  // cannot edit
	Type       uint8   `gorm:"not null;<-:create"`        // cannot edit
	Size       int64   `gorm:"not null;<-:create"`        // cannot edit
	StorageKey string  `gorm:"unique;not null;<-:create"` // cannot edit
	Hostname   string  `gorm:"not null;<-:create"`        // cannot edit
	Secret     string  `gorm:"not null;<-:create"`        // cannot edit
	Mime       *string `gorm:"<-:create"`                 // cannot edit
	Title      string
	Exp        *time.Time
	UserID     string `gorm:"index;<-:create"` // fk -> User.ID // cannot edit
	User       User   // required for M-1 relationship (I think)
}

func (u *Upload) BeforeCreate(_ *gorm.DB) (_ error) {
	u.Hostname = strings.ToLower(u.Hostname)
	return
}

// Host represents a FQDN that a User can upload to.
// Example:
//
//	"id": 1,
//	"sub": "joe",
//	"root": "stole-my-blow.wtf",
//	"user_id": "c99e9b2c-f04b-421e-b4a5-8120d2513b93"
type Host struct {
	gorm.Model
	ID     uint   `gorm:"primaryKey;autoincrement"`
	Sub    string `gorm:"<-:create"`          // cannot edit
	Root   string `gorm:"not null;<-:create"` // cannot edit
	UserID string `gorm:"index"`              // fk -> User.ID  (I don't know why but "index" tags required for fk to work)
	User   User   // required for M-1 relationship (I think)
}

func (h *Host) BeforeCreate(_ *gorm.DB) (err error) {
	h.Sub = strings.ToLower(strings.TrimSpace(h.Sub))
	h.Root = strings.ToLower(strings.TrimSpace(h.Root))
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
	ID        string  `gorm:"primaryKey"`
	Email     string  `gorm:"unique;not null"`
	DiscordID *string `gorm:"unique;index"`
	TokenID   *string `gorm:"unique;index"`
	Token     *Token  `gorm:"foreignKey:TokenID"`
	PlanID    *uint   // temp nullable
	Plan      *Plan   // temp nullable
	Hosts     []Host
	Uploads   []Upload
}

func (u *User) BeforeCreate(_ *gorm.DB) (_ error) {
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
