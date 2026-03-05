package models

import (
    "time"
    
    "github.com/google/uuid"
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

type Scan struct {
    ID          string                 `bson:"_id" json:"id"`
    UserID      string                 `bson:"user_id" json:"userId"`
    Name        string                 `bson:"name" json:"name"`
    Target      string                 `bson:"target" json:"target"`
    ScanType    []string               `bson:"scan_type" json:"scanType"`
    Status      ScanStatus              `bson:"status" json:"status"`
    Config      map[string]interface{} `bson:"config" json:"config"`
    Progress    float64                 `bson:"progress" json:"progress"`
    StartedAt   *time.Time              `bson:"started_at" json:"startedAt"`
    CompletedAt *time.Time              `bson:"completed_at" json:"completedAt"`
    Duration    int                     `bson:"duration" json:"duration"` // in seconds
    Stats       ScanStats               `bson:"stats" json:"stats"`
    CreatedAt   time.Time               `bson:"created_at" json:"createdAt"`
    UpdatedAt   time.Time               `bson:"updated_at" json:"updatedAt"`
}

type ScanStats struct {
    URLsScanned     int            `bson:"urls_scanned" json:"urlsScanned"`
    ParametersFound int            `bson:"parameters_found" json:"parametersFound"`
    FormsFound      int            `bson:"forms_found" json:"formsFound"`
    EndpointsFound  int            `bson:"endpoints_found" json:"endpointsFound"`
    Vulnerabilities map[string]int `bson:"vulnerabilities" json:"vulnerabilities"`
    Requests        int64          `bson:"requests" json:"requests"`
    Errors          int            `bson:"errors" json:"errors"`
}

type Vulnerability struct {
    ID          string                 `bson:"_id" json:"id"`
    ScanID      string                 `bson:"scan_id" json:"scanId"`
    Type        string                 `bson:"type" json:"type"`
    Severity    Severity                `bson:"severity" json:"severity"`
    Name        string                 `bson:"name" json:"name"`
    URL         string                 `bson:"url" json:"url"`
    Parameter   string                 `bson:"parameter" json:"parameter,omitempty"`
    Payload     string                 `bson:"payload" json:"payload,omitempty"`
    Evidence    string                 `bson:"evidence" json:"evidence,omitempty"`
    Description string                 `bson:"description" json:"description"`
    Remediation string                 `bson:"remediation" json:"remediation"`
    CWE         string                 `bson:"cwe" json:"cwe"`
    CVSS        float64                `bson:"cvss" json:"cvss"`
    Request     string                 `bson:"request" json:"request,omitempty"`
    Response    string                 `bson:"response" json:"response,omitempty"`
    Headers     map[string]string      `bson:"headers" json:"headers,omitempty"`
    Metadata    map[string]interface{} `bson:"metadata" json:"metadata,omitempty"`
    Verified    bool                   `bson:"verified" json:"verified"`
    Exploitable bool                   `bson:"exploitable" json:"exploitable"`
    CreatedAt   time.Time              `bson:"created_at" json:"createdAt"`
}

func NewScan() *Scan {
    return &Scan{
        ID:        uuid.New().String(),
        Status:    StatusQueued,
        Stats:     ScanStats{Vulnerabilities: make(map[string]int)},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

func NewVulnerability() *Vulnerability {
    return &Vulnerability{
        ID:        uuid.New().String(),
        CreatedAt: time.Now(),
    }
}