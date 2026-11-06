package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zverobot1k/threat-service-blokhin/internal/models"
	"github.com/zverobot1k/threat-service-blokhin/internal/repository"
	"github.com/zverobot1k/threat-service-blokhin/internal/service"
)

type handlerRepositoryMock struct {
	threat  *models.Threat
	findErr error
}

func (m *handlerRepositoryMock) Create(context.Context, *models.Threat) error { return nil }
func (m *handlerRepositoryMock) Update(context.Context, *models.Threat) error { return nil }
func (m *handlerRepositoryMock) FindByMD5(context.Context, string) (*models.Threat, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.threat == nil {
		return nil, repository.ErrNotFound
	}
	return m.threat, nil
}

func TestGetByMD5ReturnsThreat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/info/:md5", NewThreatHandler(service.NewThreatService(&handlerRepositoryMock{
		threat: &models.Threat{MD5: "d41d8cd98f00b204e9800998ecf8427e", Name: "Threat", Severity: "high", Status: "active"},
	})).GetByMD5)
	request := httptest.NewRequest(http.MethodGet, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"name":"Threat"`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestGetByMD5ReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/info/:md5", NewThreatHandler(service.NewThreatService(&handlerRepositoryMock{})).GetByMD5)
	request := httptest.NewRequest(http.MethodGet, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"message":"threat not found"`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestGetByMD5ReturnsBadRequestForInvalidMD5(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/info/:md5", NewThreatHandler(service.NewThreatService(&handlerRepositoryMock{})).GetByMD5)
	request := httptest.NewRequest(http.MethodGet, "/info/invalid", nil)
	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestCreateRejectsUnknownJSONFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/info", NewThreatHandler(service.NewThreatService(&handlerRepositoryMock{})).Create)
	request := httptest.NewRequest(http.MethodPost, "/info", strings.NewReader(`{"md5":"d41d8cd98f00b204e9800998ecf8427e","name":"Threat","unexpected":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestGetByMD5ReturnsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/info/:md5", NewThreatHandler(service.NewThreatService(&handlerRepositoryMock{findErr: errors.New("db down")})).GetByMD5)
	request := httptest.NewRequest(http.MethodGet, "/info/d41d8cd98f00b204e9800998ecf8427e", nil)
	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", response.Code)
	}
}
