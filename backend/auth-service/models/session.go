package models

import (
    "time"
    
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Session struct {
    ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
    UserID       string    `gorm:"type:uuid;index" json:"userId"`
    Token        string    `gorm:"uniqueIndex;size:512" json:"token"`
    RefreshToken string    `gorm:"uniqueIndex;size:512" json:"refreshToken"`
    UserAgent    string    `gorm:"size:255" json:"userAgent"`
    ClientIP     string    `gorm:"size:45" json:"clientIp"`
    ExpiresAt    time.Time `json:"expiresAt"`
    CreatedAt    time.Time `json:"createdAt"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
    s.ID = uuid.New().String()
    return nil
}