package models

import "time"

const (
	SeverityUnknown  = "unknown"
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"

	StatusUnknown  = "unknown"
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Threat struct {
	MD5         string `gorm:"primaryKey;size:32"`
	Name        string `gorm:"not null;size:255"`
	Description string `gorm:"size:5000"`
	Type        string `gorm:"size:100"`
	Severity    string `gorm:"not null;size:20;default:unknown"`
	Status      string `gorm:"not null;size:20;default:unknown"`
	Source      string `gorm:"size:255"`
	Confidence  int    `gorm:"not null;default:0"`
	FirstSeen   time.Time
	LastSeen    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
