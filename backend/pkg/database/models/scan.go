package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScanStatus string

const (
	StatusQueued    ScanStatus = "queued"
	StatusRunning   ScanStatus = "running"
	StatusPaused    ScanStatus = "paused"
	StatusCompleted ScanStatus = "completed"
	StatusFailed    ScanStatus = "failed"
	StatusCancelled ScanStatus = "cancelled"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// Scan Model - Like Django's Scan model
type Scan struct {
	ID       string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID   string      `gorm:"type:uuid;index;not null" json:"userId"`
	Name     string      `gorm:"size:255;not null" json:"name"`
	Target   string      `gorm:"size:2048;not null" json:"target"`
	ScanType StringArray `gorm:"type:text[]" json:"scanType"`
	Status   ScanStatus  `gorm:"size:20;default:'queued'" json:"status"`
	Priority int         `gorm:"default:0" json:"priority"`
	Progress float64     `gorm:"default:0" json:"progress"`
	Config   JSON        `gorm:"type:jsonb;default:'{}'" json:"config"`

	// Timing
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	Duration    int        `json:"duration"` // in seconds

	// Stats
	URLsScanned     int   `gorm:"default:0" json:"urlsScanned"`
	ParametersFound int   `gorm:"default:0" json:"parametersFound"`
	FormsFound      int   `gorm:"default:0" json:"formsFound"`
	EndpointsFound  int   `gorm:"default:0" json:"endpointsFound"`
	RequestsCount   int64 `gorm:"default:0" json:"requestsCount"`
	ErrorsCount     int   `gorm:"default:0" json:"errorsCount"`

	// Results
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	Findings        JSON            `gorm:"type:jsonb;default:'{}'" json:"findings"`

	// Timestamps
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Scan) TableName() string {
	return "scans"
}

func (s *Scan) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// Vulnerability Model
type Vulnerability struct {
	ID          string   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ScanID      string   `gorm:"type:uuid;index;not null" json:"scanId"`
	Type        string   `gorm:"size:100;not null" json:"type"`
	Severity    Severity `gorm:"size:20;not null" json:"severity"`
	Name        string   `gorm:"size:255;not null" json:"name"`
	URL         string   `gorm:"size:2048;not null" json:"url"`
	Parameter   string   `gorm:"size:255" json:"parameter"`
	Payload     string   `gorm:"type:text" json:"payload"`
	Evidence    string   `gorm:"type:text" json:"evidence"`
	Description string   `gorm:"type:text" json:"description"`
	Remediation string   `gorm:"type:text" json:"remediation"`
	CWE         string   `gorm:"size:20" json:"cwe"`
	CVSS        float64  `gorm:"default:0" json:"cvss"`
	Request     string   `gorm:"type:text" json:"request"`
	Response    string   `gorm:"type:text" json:"response"`
	Headers     JSON     `gorm:"type:jsonb" json:"headers"`
	Metadata    JSON     `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Verified    bool     `gorm:"default:false" json:"verified"`
	Exploitable bool     `gorm:"default:false" json:"exploitable"`

	CreatedAt time.Time `json:"createdAt"`

	Scan Scan `gorm:"foreignKey:ScanID" json:"-"`
}

func (Vulnerability) TableName() string {
	return "vulnerabilities"
}

func (v *Vulnerability) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}
