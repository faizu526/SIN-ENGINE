package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportFormat string

const (
	FormatPDF  ReportFormat = "pdf"
	FormatHTML ReportFormat = "html"
	FormatJSON ReportFormat = "json"
	FormatCSV  ReportFormat = "csv"
	FormatXML  ReportFormat = "xml"
)

type ReportStatus string

const (
	ReportPending    ReportStatus = "pending"
	ReportGenerating ReportStatus = "generating"
	ReportCompleted  ReportStatus = "completed"
	ReportFailed     ReportStatus = "failed"
)

type Report struct {
	ID            string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        string       `gorm:"type:uuid;index;not null" json:"userId"`
	Name          string       `gorm:"size:255;not null" json:"name"`
	Type          string       `gorm:"size:50;not null" json:"type"` // scan, search, exploit, custom
	Format        ReportFormat `gorm:"size:10;not null" json:"format"`
	Status        ReportStatus `gorm:"size:20;default:'pending'" json:"status"`
	SourceID      string       `gorm:"size:255" json:"sourceId"` // scan ID, search ID, etc.
	Config        JSON         `gorm:"type:jsonb;default:'{}'" json:"config"`
	Summary       JSON         `gorm:"type:jsonb" json:"summary"`
	FilePath      string       `gorm:"size:512" json:"filePath"`
	FileSize      int64        `gorm:"default:0" json:"fileSize"`
	DownloadURL   string       `gorm:"size:512" json:"downloadUrl"`
	DownloadCount int          `gorm:"default:0" json:"downloadCount"`
	GeneratedAt   *time.Time   `json:"generatedAt"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Report) TableName() string {
	return "reports"
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
