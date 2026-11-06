package service

import (
	"context"
	"errors"
	"testing"

	"github.com/zverobot1k/threat-service-blokhin/internal/dto"
	"github.com/zverobot1k/threat-service-blokhin/internal/models"
	"github.com/zverobot1k/threat-service-blokhin/internal/repository"
)

type repositoryMock struct {
	threat    *models.Threat
	findErr   error
	createErr error
	updateErr error
	created   *models.Threat
	updated   *models.Threat
}

func (m *repositoryMock) Create(_ context.Context, threat *models.Threat) error {
	m.created = threat
	return m.createErr
}

func (m *repositoryMock) FindByMD5(_ context.Context, _ string) (*models.Threat, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.threat == nil {
		return nil, repository.ErrNotFound
	}
	return m.threat, nil
}

func (m *repositoryMock) Update(_ context.Context, threat *models.Threat) error {
	m.updated = threat
	return m.updateErr
}

func TestThreatServiceCreateNormalizesAndValidates(t *testing.T) {
	repo := &repositoryMock{}
	svc := NewThreatService(repo)

	result, err := svc.Create(context.Background(), dto.ThreatCreateRequest{
		MD5: "D41D8CD98F00B204E9800998ECF8427E", Name: "  Trojan  ",
		Severity: "HIGH", Status: "active", Confidence: 90,
	})

	if err != nil {
		t.Fatal(err)
	}
	if result.MD5 != "d41d8cd98f00b204e9800998ecf8427e" || result.Name != "Trojan" || result.Severity != "high" {
		t.Fatalf("unexpected result: %+v", result)
	}
	expected := &models.Threat{
		MD5: "d41d8cd98f00b204e9800998ecf8427e", Name: "Trojan",
		Severity: "high", Status: "active", Confidence: 90,
	}
	if *repo.created != *expected {
		t.Fatalf("unexpected stored threat: %+v", repo.created)
	}
}

func TestThreatServiceRejectsInvalidInput(t *testing.T) {
	svc := NewThreatService(&repositoryMock{})

	_, err := svc.Create(context.Background(), dto.ThreatCreateRequest{
		MD5: "invalid", Name: "Threat",
	})

	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestThreatServiceReturnsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")
	svc := NewThreatService(&repositoryMock{findErr: repositoryErr})

	_, err := svc.GetByMD5(context.Background(), "d41d8cd98f00b204e9800998ecf8427e")

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
