package models

import (
	"time"
)

type DorkEngine string

const (
	EngineGoogle  DorkEngine = "google"
	EngineGitHub  DorkEngine = "github"
	EngineShodan  DorkEngine = "shodan"
	EngineCensys  DorkEngine = "censys"
	EngineFOFA    DorkEngine = "fofa"
	EngineZoomEye DorkEngine = "zoomeye"
)

type Dork struct {
	ID          string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string      `gorm:"type:uuid;index;not null" json:"userId"`
	Name        string      `gorm:"size:255;not null" json:"name"`
	Query       string      `gorm:"type:text;not null" json:"query"`
	Engine      DorkEngine  `gorm:"size:50;not null" json:"engine"`
	Category    string      `gorm:"size:100" json:"category"`
	Description string      `gorm:"type:text" json:"description"`
	Tags        StringArray `gorm:"type:text[]" json:"tags"`
	IsPublic    bool        `gorm:"default:false" json:"isPublic"`
	UseCount    int         `gorm:"default:0" json:"useCount"`
	LastUsed    *time.Time  `json:"lastUsed"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Dork) TableName() string {
	return "dorks"
}

type DorkResult struct {
	ID          string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DorkID      string      `gorm:"type:uuid;index;not null" json:"dorkId"`
	UserID      string      `gorm:"type:uuid;index;not null" json:"userId"`
	URL         string      `gorm:"size:2048" json:"url"`
	Title       string      `gorm:"size:512" json:"title"`
	Description string      `gorm:"type:text" json:"description"`
	Data        JSON        `gorm:"type:jsonb" json:"data"`
	Severity    Severity    `gorm:"size:20" json:"severity"`
	Tags        StringArray `gorm:"type:text[]" json:"tags"`
	FoundAt     time.Time   `json:"foundAt"`
	CreatedAt   time.Time   `json:"createdAt"`
}

func (DorkResult) TableName() string {
	return "dork_results"
}
