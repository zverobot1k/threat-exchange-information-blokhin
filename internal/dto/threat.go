package dto

import "time"

type ThreatCreateRequest struct {
	MD5         string     `json:"md5" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	Source      string     `json:"source"`
	Confidence  int        `json:"confidence"`
	FirstSeen   *time.Time `json:"first_seen"`
	LastSeen    *time.Time `json:"last_seen"`
}

type ThreatUpdateRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	Source      string     `json:"source"`
	Confidence  int        `json:"confidence"`
	FirstSeen   *time.Time `json:"first_seen"`
	LastSeen    *time.Time `json:"last_seen"`
}

type ThreatResponse struct {
	MD5         string    `json:"md5"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type,omitempty"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	Source      string    `json:"source,omitempty"`
	Confidence  int       `json:"confidence"`
	FirstSeen   time.Time `json:"first_seen,omitempty"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
