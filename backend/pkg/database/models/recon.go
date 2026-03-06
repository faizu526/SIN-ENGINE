package models

import (
	"time"
)

type ReconType string

const (
	ReconDNS       ReconType = "dns"
	ReconSubdomain ReconType = "subdomain"
	ReconPort      ReconType = "port"
	ReconService   ReconType = "service"
	ReconTech      ReconType = "technology"
	ReconWhois     ReconType = "whois"
	ReconSSL       ReconType = "ssl"
	ReconEmail     ReconType = "email"
	ReconSocial    ReconType = "social"
	ReconGitHub    ReconType = "github"
)

type ReconTarget struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string     `gorm:"type:uuid;index;not null" json:"userId"`
	Target      string     `gorm:"size:255;not null;index" json:"target"` // domain, IP, email, etc.
	Type        ReconType  `gorm:"size:50;not null" json:"type"`
	Status      string     `gorm:"size:20;default:'pending'" json:"status"`
	Progress    float64    `gorm:"default:0" json:"progress"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (ReconTarget) TableName() string {
	return "recon_targets"
}

type ReconResult struct {
	ID         string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TargetID   string      `gorm:"type:uuid;index;not null" json:"targetId"`
	Type       ReconType   `gorm:"size:50;not null" json:"type"`
	Name       string      `gorm:"size:255" json:"name"`
	Value      string      `gorm:"type:text" json:"value"`
	Data       JSON        `gorm:"type:jsonb" json:"data"`
	Source     string      `gorm:"size:100" json:"source"`
	Confidence float64     `gorm:"default:1.0" json:"confidence"`
	Tags       StringArray `gorm:"type:text[]" json:"tags"`
	FoundAt    time.Time   `json:"foundAt"`
	CreatedAt  time.Time   `json:"createdAt"`
}

func (ReconResult) TableName() string {
	return "recon_results"
}

type Subdomain struct {
	ID          string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Domain      string      `gorm:"size:255;index;not null" json:"domain"`
	Subdomain   string      `gorm:"size:255;index;not null" json:"subdomain"`
	IPAddresses StringArray `gorm:"type:text[]" json:"ipAddresses"`
	Resolves    bool        `gorm:"default:true" json:"resolves"`
	Source      string      `gorm:"size:100" json:"source"`
	FoundAt     time.Time   `json:"foundAt"`
}

func (Subdomain) TableName() string {
	return "subdomains"
}
