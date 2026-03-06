package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User Model - Like Django's User model
type User struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	FullName     string     `gorm:"size:100" json:"fullName"`
	Role         string     `gorm:"size:20;default:'user'" json:"role"`
	APIKey       string     `gorm:"uniqueIndex;size:64" json:"apiKey"`
	RateLimit    int        `gorm:"default:1000" json:"rateLimit"`
	IsActive     bool       `gorm:"default:true" json:"isActive"`
	IsVerified   bool       `gorm:"default:false" json:"isVerified"`
	LastLogin    *time.Time `json:"lastLogin"`
	Preferences  JSON       `gorm:"type:jsonb;default:'{}'" json:"preferences"`

	// Stats
	TotalScans    int `gorm:"default:0" json:"totalScans"`
	TotalFindings int `gorm:"default:0" json:"totalFindings"`

	// Timestamps
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Sessions []Session `json:"sessions,omitempty"`
	Scans    []Scan    `json:"scans,omitempty"`
	APIKeys  []APIKey  `json:"apiKeys,omitempty"`
}

// TableName - Like Django's Meta.db_table
func (User) TableName() string {
	return "users"
}

// BeforeCreate - Like Django's pre_save signal
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// SetPassword - Like Django's set_password()
func (u *User) SetPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashed)
	return nil
}

// CheckPassword - Like Django's check_password()
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Session Model
type Session struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"type:uuid;index;not null" json:"userId"`
	Token        string    `gorm:"uniqueIndex;size:512;not null" json:"token"`
	RefreshToken string    `gorm:"uniqueIndex;size:512;not null" json:"refreshToken"`
	UserAgent    string    `gorm:"size:255" json:"userAgent"`
	ClientIP     string    `gorm:"size:45" json:"clientIp"`
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Session) TableName() string {
	return "sessions"
}

// APIKey Model
type APIKey struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"type:uuid;index;not null" json:"userId"`
	Name      string     `gorm:"size:100;not null" json:"name"`
	Key       string     `gorm:"uniqueIndex;size:64;not null" json:"key"`
	LastUsed  *time.Time `json:"lastUsed"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
