package models

import (
	"time"
)

type BreachSource string

const (
	SourceHIBP      BreachSource = "hibp"
	SourceDehashed  BreachSource = "dehashed"
	SourceSnusbase  BreachSource = "snusbase"
	SourceLeakCheck BreachSource = "leakcheck"
	SourceBreachDir BreachSource = "breachdirectory"
)

type Breach struct {
	ID          string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string       `gorm:"type:uuid;index;not null" json:"userId"`
	Email       string       `gorm:"size:255;index" json:"email"`
	Username    string       `gorm:"size:100;index" json:"username"`
	Domain      string       `gorm:"size:255;index" json:"domain"`
	Source      BreachSource `gorm:"size:50" json:"source"`
	BreachName  string       `gorm:"size:255" json:"breachName"`
	BreachDate  *time.Time   `json:"breachDate"`
	DataClasses StringArray  `gorm:"type:text[]" json:"dataClasses"`
	Passwords   JSON         `gorm:"type:jsonb" json:"passwords"`
	Emails      StringArray  `gorm:"type:text[]" json:"emails"`
	Usernames   StringArray  `gorm:"type:text[]" json:"usernames"`
	IPAddresses StringArray  `gorm:"type:text[]" json:"ipAddresses"`
	Devices     JSON         `gorm:"type:jsonb" json:"devices"`
	IsVerified  bool         `gorm:"default:false" json:"isVerified"`
	IsSensitive bool         `gorm:"default:false" json:"isSensitive"`
	FoundAt     time.Time    `json:"foundAt"`
	CreatedAt   time.Time    `json:"createdAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Breach) TableName() string {
	return "breaches"
}

type BreachMonitor struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        string     `gorm:"type:uuid;index;not null" json:"userId"`
	Email         string     `gorm:"size:255;index" json:"email"`
	Domain        string     `gorm:"size:255;index" json:"domain"`
	IsActive      bool       `gorm:"default:true" json:"isActive"`
	NotifyEmail   bool       `gorm:"default:true" json:"notifyEmail"`
	NotifyWebhook string     `gorm:"size:512" json:"notifyWebhook"`
	LastChecked   *time.Time `json:"lastChecked"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func (BreachMonitor) TableName() string {
	return "breach_monitors"
}
