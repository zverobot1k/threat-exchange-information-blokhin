package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zverobot1k/threat-service-blokhin/internal/dto"
	"github.com/zverobot1k/threat-service-blokhin/internal/models"
	"github.com/zverobot1k/threat-service-blokhin/internal/repository"
)

var ErrNotFound = errors.New("threat not found")
var ErrAlreadyExists = errors.New("threat already exists")
var ErrInvalid = errors.New("invalid threat data")

var md5Pattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

type ThreatService struct {
	repository repository.ThreatRepository
}

func NewThreatService(repo repository.ThreatRepository) *ThreatService {
	return &ThreatService{repository: repo}
}

func (s *ThreatService) GetByMD5(ctx context.Context, md5 string) (*dto.ThreatResponse, error) {
	md5, err := normalizeMD5(md5)
	if err != nil {
		return nil, err
	}
	threat, err := s.repository.FindByMD5(ctx, md5)
	if repository.IsNotFound(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toResponse(threat), nil
}

func (s *ThreatService) Create(ctx context.Context, request dto.ThreatCreateRequest) (*dto.ThreatResponse, error) {
	md5, err := normalizeMD5(request.MD5)
	if err != nil {
		return nil, err
	}
	threat, err := validateAndBuild(md5, request.Name, request.Description, request.Type, request.Severity, request.Status, request.Source, request.Confidence, request.FirstSeen, request.LastSeen)
	if err != nil {
		return nil, err
	}
	if _, err = s.repository.FindByMD5(ctx, md5); err == nil {
		return nil, ErrAlreadyExists
	} else if !repository.IsNotFound(err) {
		return nil, err
	}
	if err = s.repository.Create(ctx, threat); err != nil {
		if repository.IsConflict(err) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return toResponse(threat), nil
}

func (s *ThreatService) Update(ctx context.Context, md5 string, request dto.ThreatUpdateRequest) (*dto.ThreatResponse, error) {
	md5, err := normalizeMD5(md5)
	if err != nil {
		return nil, err
	}
	threat, err := s.repository.FindByMD5(ctx, md5)
	if repository.IsNotFound(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	updated, err := validateAndBuild(md5, request.Name, request.Description, request.Type, request.Severity, request.Status, request.Source, request.Confidence, request.FirstSeen, request.LastSeen)
	if err != nil {
		return nil, err
	}
	updated.CreatedAt = threat.CreatedAt
	updated.UpdatedAt = threat.UpdatedAt
	if err = s.repository.Update(ctx, updated); err != nil {
		return nil, err
	}
	return toResponse(updated), nil
}

func normalizeMD5(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if !md5Pattern.MatchString(value) {
		return "", fmt.Errorf("%w: md5 must contain exactly 32 hexadecimal characters", ErrInvalid)
	}
	return value, nil
}

func validateAndBuild(md5, name, description, threatType, severity, status, source string, confidence int, firstSeen, lastSeen *time.Time) (*models.Threat, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return nil, fmt.Errorf("%w: name must contain 1 to 255 characters", ErrInvalid)
	}
	if len(description) > 5000 || len(threatType) > 100 || len(source) > 255 || confidence < 0 || confidence > 100 {
		return nil, fmt.Errorf("%w: field length or confidence is invalid", ErrInvalid)
	}
	if firstSeen != nil && lastSeen != nil && lastSeen.Before(*firstSeen) {
		return nil, fmt.Errorf("%w: last_seen cannot be before first_seen", ErrInvalid)
	}
	severity = normalizeEnum(severity, models.SeverityUnknown)
	status = normalizeEnum(status, models.StatusUnknown)
	if !validSeverity(severity) || !validStatus(status) {
		return nil, fmt.Errorf("%w: unsupported severity or status", ErrInvalid)
	}
	threat := &models.Threat{MD5: md5, Name: name, Description: description, Type: threatType, Severity: severity, Status: status, Source: source, Confidence: confidence}
	if firstSeen != nil {
		threat.FirstSeen = *firstSeen
	}
	if lastSeen != nil {
		threat.LastSeen = *lastSeen
	}
	return threat, nil
}

func normalizeEnum(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	return value
}

func validSeverity(value string) bool {
	return value == models.SeverityUnknown || value == models.SeverityLow || value == models.SeverityMedium || value == models.SeverityHigh || value == models.SeverityCritical
}
func validStatus(value string) bool {
	return value == models.StatusUnknown || value == models.StatusActive || value == models.StatusInactive
}

func toResponse(threat *models.Threat) *dto.ThreatResponse {
	return &dto.ThreatResponse{MD5: threat.MD5, Name: threat.Name, Description: threat.Description, Type: threat.Type, Severity: threat.Severity, Status: threat.Status, Source: threat.Source, Confidence: threat.Confidence, FirstSeen: threat.FirstSeen, LastSeen: threat.LastSeen, CreatedAt: threat.CreatedAt, UpdatedAt: threat.UpdatedAt}
}
