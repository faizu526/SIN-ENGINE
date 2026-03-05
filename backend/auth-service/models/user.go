package models

import (
    "time"
    
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type User struct {
    ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
    Username     string    `gorm:"uniqueIndex;size:50" json:"username"`
    Email        string    `gorm:"uniqueIndex;size:255" json:"email"`
    PasswordHash string    `gorm:"size:255" json:"-"`
    FullName     string    `gorm:"size:100" json:"fullName"`
    Role         string    `gorm:"size:20;default:'user'" json:"role"`
    APIKey       string    `gorm:"uniqueIndex;size:64" json:"apiKey"`
    RateLimit    int       `gorm:"default:1000" json:"rateLimit"`
    Active       bool      `gorm:"default:true" json:"active"`
    LastLogin    *time.Time `json:"lastLogin"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.ID = uuid.New().String()
    return nil
}

func (u *User) SetPassword(password string) error {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.PasswordHash = string(hashed)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
    return err == nil
}