package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zverobot1k/threat-service-blokhin/internal/dto"
	"github.com/zverobot1k/threat-service-blokhin/internal/service"
)

type ThreatHandler struct {
	service *service.ThreatService
}

func NewThreatHandler(threatService *service.ThreatService) *ThreatHandler {
	return &ThreatHandler{service: threatService}
}

func (h *ThreatHandler) GetByMD5(c *gin.Context) {
	result, err := h.service.GetByMD5(c.Request.Context(), c.Param("md5"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ThreatHandler) Create(c *gin.Context) {
	var request dto.ThreatCreateRequest
	if err := decodeJSON(c, &request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}
	result, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *ThreatHandler) Update(c *gin.Context) {
	var request dto.ThreatUpdateRequest
	if err := decodeJSON(c, &request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}
	result, err := h.service.Update(c.Request.Context(), c.Param("md5"), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func decodeJSON(c *gin.Context, target any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalid):
		writeError(c, http.StatusBadRequest, "invalid request", err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(c, http.StatusNotFound, "threat not found", err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		writeError(c, http.StatusConflict, "threat already exists", err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "internal server error", "an internal error occurred")
	}
}

func writeError(c *gin.Context, status int, message, details string) {
	c.JSON(status, dto.ErrorResponse{Message: message, Details: details})
}
